package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Akendo/assigment1/notification-service/internal/subscriber"
	"github.com/Akendo/assigment1/pkg/messaging"
	"github.com/nats-io/nats.go"
)

const (
	maxConnectAttempts = 5
	baseBackoff        = time.Second
)

func Run() error {
	logger := log.Default()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	conn, err := connectWithRetry(natsURL, logger)
	if err != nil {
		return err
	}
	defer conn.Close()

	eventSubscriber := subscriber.New(conn, logger)
	if err := eventSubscriber.SubscribeAll(); err != nil {
		return fmt.Errorf("subscribe to NATS subjects: %w", err)
	}

	<-ctx.Done()

	if err := conn.Drain(); err != nil {
		return fmt.Errorf("drain NATS connection: %w", err)
	}

	return nil
}

func connectWithRetry(natsURL string, logger *log.Logger) (*nats.Conn, error) {
	var lastErr error

	for attempt := 1; attempt <= maxConnectAttempts; attempt++ {
		conn, err := messaging.ConnectNATS(natsURL, "notification-service", logger)
		if err == nil {
			return conn, nil
		}

		lastErr = err
		logger.Printf("failed to connect to NATS (attempt %d/%d): %v", attempt, maxConnectAttempts, err)

		if attempt == maxConnectAttempts {
			break
		}

		time.Sleep(baseBackoff * time.Duration(1<<(attempt-1)))
	}

	return nil, fmt.Errorf("notification service could not connect to NATS after %d attempts: %w", maxConnectAttempts, lastErr)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
