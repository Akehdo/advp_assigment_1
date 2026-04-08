package http

import (
	"errors"
	"fmt"
	nethttp "net/http"

	"github.com/Akendo/assigment1/appointment/internal/model"
	"github.com/Akendo/assigment1/appointment/internal/repository"
	"github.com/Akendo/assigment1/appointment/internal/transport/http/dto"
	"github.com/Akendo/assigment1/appointment/internal/usecase"
	"github.com/gin-gonic/gin"
)

type appointmentService interface {
	CreateAppointment(title, description, doctorID string) (*model.Appointment, error)
	GetAppointment(id string) (*model.Appointment, error)
	ListAppointments() ([]*model.Appointment, error)
	UpdateStatus(id string, status model.Status) (*model.Appointment, error)
}

type Handler struct {
	service appointmentService
}

func NewHandler(service appointmentService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.GET("/health", h.health)
	router.POST("/appointments", h.createAppointment)
	router.GET("/appointments/:id", h.getAppointment)
	router.GET("/appointments", h.listAppointments)
	router.PATCH("/appointments/:id/status", h.updateStatus)
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(nethttp.StatusOK, gin.H{"status": "appointment service ok"})
}

func (h *Handler) createAppointment(c *gin.Context) {
	var req dto.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appointment, err := h.service.CreateAppointment(req.Title, req.Description, req.DoctorID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrDoctorNotFound):
			c.JSON(nethttp.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, usecase.ErrDoctorServiceUnavailable):
			c.JSON(nethttp.StatusServiceUnavailable, gin.H{"error": err.Error()})
		default:
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(nethttp.StatusCreated, dto.NewAppointmentResponse(appointment))
}

func (h *Handler) getAppointment(c *gin.Context) {
	id := c.Param("id")
	appointment, err := h.service.GetAppointment(id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAppointmentNotFound):
			c.JSON(nethttp.StatusNotFound, gin.H{"error": fmt.Sprintf("appointment with id %s not found", id)})
		default:
			c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(nethttp.StatusOK, dto.NewAppointmentResponse(appointment))
}

func (h *Handler) listAppointments(c *gin.Context) {
	appointments, err := h.service.ListAppointments()
	if err != nil {
		c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(nethttp.StatusOK, dto.NewAppointmentListResponse(appointments))
}

func (h *Handler) updateStatus(c *gin.Context) {
	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appointment, err := h.service.UpdateStatus(c.Param("id"), model.Status(req.Status))
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAppointmentNotFound):
			c.JSON(nethttp.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, model.ErrInvalidStatus), errors.Is(err, model.ErrStatusTransitionNotAllowed):
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(nethttp.StatusOK, dto.NewAppointmentResponse(appointment))
}
