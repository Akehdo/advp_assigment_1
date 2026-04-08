package app

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/Akendo/assigment1/appointment/internal/gateway"
	"github.com/Akendo/assigment1/appointment/internal/repository/postgres"
	transporthttp "github.com/Akendo/assigment1/appointment/internal/transport/http"
	"github.com/Akendo/assigment1/appointment/internal/usecase"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func Run() error {
	db, err := openDB(
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5434"),
		getEnv("DB_NAME", "appointment_db"),
		getEnv("DB_USER", "appointment_user"),
		getEnv("DB_PASSWORD", "appointment_pass"),
	)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := ensureSchema(db); err != nil {
		return err
	}

	repo := postgres.NewAppointmentRepository(db)
	doctorGateway := gateway.NewDoctorRESTGateway(getEnv("DOCTOR_SERVICE_URL", "http://localhost:8081"))
	service := usecase.NewAppointmentService(repo, doctorGateway)
	handler := transporthttp.NewHandler(service)

	router := gin.Default()
	handler.RegisterRoutes(router)

	return router.Run(":8082")
}

func openDB(host, port, name, user, password string) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		user,
		password,
		name,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func ensureSchema(db *sql.DB) error {
	const query = `
		CREATE TABLE IF NOT EXISTS appointments (
			id UUID PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			doctor_id UUID NOT NULL,
			status TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`

	_, err := db.Exec(query)
	return err
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
