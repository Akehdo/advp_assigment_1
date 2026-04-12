package app

import (
	"database/sql"
	"fmt"
	"net"
	"os"

	"github.com/Akendo/assigment1/doctor/internal/repository/postgres"
	transportgrpc "github.com/Akendo/assigment1/doctor/internal/transport/grpc"
	"github.com/Akendo/assigment1/doctor/internal/usecase"
	doctorpb "github.com/Akendo/assigment1/doctor/proto"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func Run() error {
	db, err := openDB(
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5433"),
		getEnv("DB_NAME", "doctor_db"),
		getEnv("DB_USER", "doctor_user"),
		getEnv("DB_PASSWORD", "doctor_pass"),
	)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := ensureSchema(db); err != nil {
		return err
	}

	repo := postgres.NewDoctorRepository(db)
	service := usecase.NewDoctorService(repo)
	handler := transportgrpc.NewHandler(service)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	doctorpb.RegisterDoctorServiceServer(grpcServer, handler)

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
		CREATE TABLE IF NOT EXISTS doctors (
			id UUID PRIMARY KEY,
			full_name TEXT NOT NULL,
			specialization TEXT,
			email TEXT NOT NULL UNIQUE
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
