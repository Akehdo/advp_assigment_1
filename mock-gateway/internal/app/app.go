package app

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	defaultPort              = "8080"
	defaultFailurePercent    = 20
	defaultReadHeaderTimeout = 5 * time.Second
)

type notificationRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
	Channel        string `json:"channel"`
	Recipient      string `json:"recipient"`
	Message        string `json:"message"`
}

type notificationResponse struct {
	Status string `json:"status"`
}

type requestLog struct {
	Time           string `json:"time"`
	Level          string `json:"level"`
	Status         string `json:"status"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	Channel        string `json:"channel,omitempty"`
	Recipient      string `json:"recipient,omitempty"`
	Error          string `json:"error,omitempty"`
}

type server struct {
	mu             sync.Mutex
	processedKeys  map[string]struct{}
	random         *rand.Rand
	failurePercent int
	encoder        *json.Encoder
}

func Run() error {
	srv := &server{
		processedKeys:  make(map[string]struct{}),
		random:         rand.New(rand.NewSource(time.Now().UnixNano())),
		failurePercent: getEnvAsInt("MOCK_GATEWAY_FAILURE_PERCENT", defaultFailurePercent),
		encoder:        json.NewEncoder(os.Stdout),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/notify", srv.handleNotify)

	httpServer := &http.Server{
		Addr:              ":" + getEnv("MOCK_GATEWAY_PORT", defaultPort),
		Handler:           mux,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
	}

	return httpServer.ListenAndServe()
}

func (s *server) handleNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	var req notificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid json", "")
		return
	}

	if req.IdempotencyKey == "" || req.Channel == "" || req.Recipient == "" || req.Message == "" {
		s.respondError(w, http.StatusBadRequest, "all fields are required", req.IdempotencyKey)
		return
	}

	if s.isDuplicate(req.IdempotencyKey) {
		s.writeLog(requestLog{
			Time:           time.Now().UTC().Format(time.RFC3339),
			Level:          "info",
			Status:         "duplicate",
			IdempotencyKey: req.IdempotencyKey,
			Channel:        req.Channel,
			Recipient:      req.Recipient,
		})
		s.respondJSON(w, http.StatusOK, notificationResponse{Status: "duplicate"})
		return
	}

	if s.shouldFail() {
		s.writeLog(requestLog{
			Time:           time.Now().UTC().Format(time.RFC3339),
			Level:          "error",
			Status:         "unavailable",
			IdempotencyKey: req.IdempotencyKey,
			Channel:        req.Channel,
			Recipient:      req.Recipient,
			Error:          "random 503",
		})
		s.respondJSON(w, http.StatusServiceUnavailable, notificationResponse{Status: "unavailable"})
		return
	}

	s.markProcessed(req.IdempotencyKey)
	s.writeLog(requestLog{
		Time:           time.Now().UTC().Format(time.RFC3339),
		Level:          "info",
		Status:         "accepted",
		IdempotencyKey: req.IdempotencyKey,
		Channel:        req.Channel,
		Recipient:      req.Recipient,
	})
	s.respondJSON(w, http.StatusOK, notificationResponse{Status: "accepted"})
}

func (s *server) isDuplicate(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.processedKeys[key]
	return exists
}

func (s *server) markProcessed(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.processedKeys[key] = struct{}{}
}

func (s *server) shouldFail() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.random.Intn(100) < s.failurePercent
}

func (s *server) respondError(w http.ResponseWriter, statusCode int, errMsg, key string) {
	s.writeLog(requestLog{
		Time:           time.Now().UTC().Format(time.RFC3339),
		Level:          "error",
		Status:         "rejected",
		IdempotencyKey: key,
		Error:          errMsg,
	})
	http.Error(w, errMsg, statusCode)
}

func (s *server) respondJSON(w http.ResponseWriter, statusCode int, response notificationResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}

func (s *server) writeLog(entry requestLog) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = s.encoder.Encode(entry)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 || parsed > 100 {
		return fallback
	}

	return parsed
}
