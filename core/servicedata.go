package core

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Service struct {
	ServiceID      uint64 `json:"service_id"`
	CreateDateTime string `json:"create_datetime"`
}

type ServiceData struct {
	data map[string]Service
	mu   sync.Mutex
}

func NewService() Service {
	return Service{
		ServiceID:      uint64(rand.Intn(1000)),
		CreateDateTime: time.Now().Format(time.DateTime),
	}
}

func NewServiceData() ServiceData {
	return ServiceData{data: make(map[string]Service)}
}

func (s *ServiceData) Get(key string) (Service, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.data[key]
	return val, ok
}

func (s *ServiceData) Set(key string, val Service) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = val
}

func (s *ServiceData) Add(serviceName string) {
	data := NewService()
	s.Set(serviceName, data)
}

func (s *ServiceData) Len() int {
	return len(s.data)
}

func (s *ServiceData) Marshal() (string, error) {
	dm, err := json.Marshal(s.data)
	if err != nil {
		return "", fmt.Errorf("Error in Marshal `ServiceData` instance")
	}
	return string(dm), nil
}

func (s *ServiceData) GetAll() map[string]Service {
	return s.data
}

func (s *ServiceData) Remove(serviceName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.data {
		if k == serviceName {
			delete(s.data, k)
		}
	}
}
