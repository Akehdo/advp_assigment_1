package app

import (
	"database/sql"
	"fmt"
	"net"
	"os"

	"github.com/Akendo/assigment1/appointment/internal/client"
	"github.com/Akendo/assigment1/appointment/internal/repository/postgres"
	transportgrpc "github.com/Akendo/assigment1/appointment/internal/transport/grpc"
	"github.com/Akendo/assigment1/appointment/internal/usecase"
	appointmentpb "github.com/Akendo/assigment1/appointment/proto"
	doctorpb "github.com/Akendo/assigment1/doctor/proto"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	conn, err := grpc.NewClient(
		getEnv("DOCTOR_SERVICE_ADDR", "localhost:50051"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	doctorServiceClient := doctorpb.NewDoctorServiceClient(conn)
	doctorGateway := client.NewDoctorGRPCClient(doctorServiceClient)

	service := usecase.NewAppointmentService(repo, doctorGateway)
	handler := transportgrpc.NewHandler(service)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	appointmentpb.RegisterAppointmentServiceServer(grpcServer, handler)

	return grpcServer.Serve(lis)

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
