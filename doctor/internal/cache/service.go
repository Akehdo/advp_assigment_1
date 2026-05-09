package cache

import (
	"context"
	"log"
	"time"

	"github.com/Akendo/assigment1/doctor/internal/model"
)

const (
	doctorsListKey        = "doctors:list"
	cacheOperationTimeout = 2 * time.Second
)

type doctorService interface {
	CreateDoctor(fullName, specialization, email string) (*model.Doctor, error)
	GetDoctor(id string) (*model.Doctor, error)
	ListDoctors() ([]*model.Doctor, error)
}

type DoctorService struct {
	next   doctorService
	cache  CacheRepository
	ttl    time.Duration
	logger *log.Logger
}

func NewDoctorService(next doctorService, cache CacheRepository, ttl time.Duration, logger *log.Logger) *DoctorService {
	if logger == nil {
		logger = log.Default()
	}

	return &DoctorService{
		next:   next,
		cache:  cache,
		ttl:    ttl,
		logger: logger,
	}
}

func (s *DoctorService) CreateDoctor(fullName, specialization, email string) (*model.Doctor, error) {
	doctor, err := s.next.CreateDoctor(fullName, specialization, email)
	if err != nil {
		return nil, err
	}

	s.delete(doctorsListKey)

	return doctor, nil
}

func (s *DoctorService) GetDoctor(id string) (*model.Doctor, error) {
	if s.cache != nil {
		key := doctorByIDKey(id)
		var cached model.Doctor

		ctx, cancel := s.cacheContext()
		found, err := s.cache.Get(ctx, key, &cached)
		cancel()
		if err != nil {
			s.logger.Printf("doctor cache get %q: %v", key, err)
		} else if found {
			s.logger.Printf("doctor cache hit %q", key)
			return &cached, nil
		}
	}

	doctor, err := s.next.GetDoctor(id)
	if err != nil {
		return nil, err
	}

	s.set(doctorByIDKey(id), doctor)

	return doctor, nil
}

func (s *DoctorService) ListDoctors() ([]*model.Doctor, error) {
	if s.cache != nil {
		var cached []*model.Doctor

		ctx, cancel := s.cacheContext()
		found, err := s.cache.Get(ctx, doctorsListKey, &cached)
		cancel()
		if err != nil {
			s.logger.Printf("doctor cache get %q: %v", doctorsListKey, err)
		} else if found {
			s.logger.Printf("doctor cache hit %q", doctorsListKey)
			return cached, nil
		}
	}

	doctors, err := s.next.ListDoctors()
	if err != nil {
		return nil, err
	}

	s.set(doctorsListKey, doctors)

	return doctors, nil
}

func (s *DoctorService) set(key string, value any) {
	if s.cache == nil {
		return
	}

	ctx, cancel := s.cacheContext()
	defer cancel()

	if err := s.cache.Set(ctx, key, value, s.ttl); err != nil {
		s.logger.Printf("doctor cache set %q: %v", key, err)
	}
}

func (s *DoctorService) delete(key string) {
	if s.cache == nil {
		return
	}

	ctx, cancel := s.cacheContext()
	defer cancel()

	if err := s.cache.Delete(ctx, key); err != nil {
		s.logger.Printf("doctor cache delete %q: %v", key, err)
	}
}

func (s *DoctorService) cacheContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), cacheOperationTimeout)
}

func doctorByIDKey(id string) string {
	return "doctor:" + id
}
