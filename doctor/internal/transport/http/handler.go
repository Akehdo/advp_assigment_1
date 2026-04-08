package http

import (
	"errors"
	"fmt"
	nethttp "net/http"

	"github.com/Akendo/assigment1/doctor/internal/model"
	"github.com/Akendo/assigment1/doctor/internal/repository"
	"github.com/Akendo/assigment1/doctor/internal/transport/http/dto"
	"github.com/gin-gonic/gin"
)

type doctorService interface {
	CreateDoctor(fullName, specialization, email string) (*model.Doctor, error)
	GetDoctor(id string) (*model.Doctor, error)
	ListDoctors() ([]*model.Doctor, error)
}

type Handler struct {
	service doctorService
}

func NewHandler(service doctorService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	router.GET("/health", h.health)
	router.POST("/doctors", h.createDoctor)
	router.GET("/doctors/:id", h.getDoctor)
	router.GET("/doctors", h.listDoctors)
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(nethttp.StatusOK, gin.H{"status": "doctor service ok"})
}

func (h *Handler) createDoctor(c *gin.Context) {
	var req dto.CreateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doctor, err := h.service.CreateDoctor(req.FullName, req.Specialization, req.Email)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDoctorEmailAlreadyExists):
			c.JSON(nethttp.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(nethttp.StatusCreated, dto.NewDoctorResponse(doctor))
}

func (h *Handler) getDoctor(c *gin.Context) {
	id := c.Param("id")
	doctor, err := h.service.GetDoctor(id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDoctorNotFound):
			c.JSON(nethttp.StatusNotFound, gin.H{"error": fmt.Sprintf("doctor with id %s not found", id)})
		default:
			c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(nethttp.StatusOK, dto.NewDoctorResponse(doctor))
}

func (h *Handler) listDoctors(c *gin.Context) {
	doctors, err := h.service.ListDoctors()
	if err != nil {
		c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(nethttp.StatusOK, dto.NewDoctorListResponse(doctors))
}
