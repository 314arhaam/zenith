package core

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type service struct {
	ServiceID      uint64 `json:"service_id"`
	CreateDateTime string `json:"create_datetime"`
}

type System struct {
	data map[string]service
	mu   sync.Mutex
}

func NewCustomService(serviceId uint64, dt string) service {
	return service{
		ServiceID:      serviceId,
		CreateDateTime: dt,
	}
}

func NewService() service {
	return service{
		ServiceID:      uint64(rand.Intn(1000)),
		CreateDateTime: time.Now().Format(time.DateTime),
	}
}

func NewSystem() System {
	return System{data: make(map[string]service)}
}

func (s *System) Get(key string) (service, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.data[key]
	return val, ok
}

func (s *System) Set(key string, val service) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = val
}

func (s *System) Add(serviceName string) {
	data := NewService()
	s.Set(serviceName, data)
}

func (s *System) Len() int {
	return len(s.data)
}

func (s *System) Marshal() (string, error) {
	dm, err := json.Marshal(s.data)
	if err != nil {
		return "", fmt.Errorf("Error in Marshal `System` instance")
	}
	return string(dm), nil
}

func (s *System) GetAll() map[string]service {
	return s.data
}

func (s *System) Remove(serviceName string) bool {
	if _, ok := s.Get(serviceName); !ok {
		return false
	}
	s.mu.Lock()
	for k := range s.data {
		if k == serviceName {
			delete(s.data, k)
		}
	}
	s.mu.Unlock()
	return true
}
