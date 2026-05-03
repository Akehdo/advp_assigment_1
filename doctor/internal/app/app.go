package app

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"

	"github.com/Akendo/assigment1/doctor/internal/repository/postgres"
	transportgrpc "github.com/Akendo/assigment1/doctor/internal/transport/grpc"
	"github.com/Akendo/assigment1/doctor/internal/usecase"
	doctorpb "github.com/Akendo/assigment1/doctor/proto"
	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func Run() error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := runMigrations(db, "doctor"); err != nil {
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

func openDB() (*sql.DB, error) {
	dsn := firstNonEmpty(
		os.Getenv("DATABASE_URL"),
		os.Getenv("DB_DSN"),
	)
	if dsn == "" {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5433"),
			getEnv("DB_USER", "doctor_user"),
			getEnv("DB_PASSWORD", "doctor_pass"),
			getEnv("DB_NAME", "doctor_db"),
		)
	}

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

func runMigrations(db *sql.DB, serviceName string) error {
	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	migrationsPath, err := findMigrationsPath(serviceName)
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		fileURL(migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func findMigrationsPath(serviceName string) (string, error) {
	candidates := []string{
		filepath.Join(".", "migrations"),
		filepath.Join(".", serviceName, "migrations"),
	}

	if executablePath, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executablePath), "migrations"))
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err != nil || !info.IsDir() {
			continue
		}

		absolutePath, err := filepath.Abs(candidate)
		if err != nil {
			return "", fmt.Errorf("resolve migrations path %q: %w", candidate, err)
		}

		return absolutePath, nil
	}

	return "", fmt.Errorf("migrations directory not found for %s service", serviceName)
}

func fileURL(path string) string {
	return (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(path),
	}).String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
