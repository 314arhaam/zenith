package core

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Service struct {
	ServiceID      uint64 `json:"service_id"`
	CreateDateTime string `json:"create_datetime"`
}

// System is an in-memory registry. Its zero value is ready to use.
// A System must not be copied after first use. Prefer NewSystem.
type System struct {
	mu         sync.RWMutex
	data       map[string]Service
	version    uint64
	cachedJSON string
	// Serialize cache fills, not ordinary registry reads/writes. Encoding and
	// network writes never hold mu, so slow clients cannot block registration.
	marshalMu sync.Mutex
}

var nextServiceID atomic.Uint64

func init() { nextServiceID.Store(uint64(time.Now().UnixNano())) }

func NewCustomService(serviceID uint64, dt string) Service {
	return Service{ServiceID: serviceID, CreateDateTime: dt}
}

// NewService returns a process-local identifier, not a durable/distributed ID.
func NewService() Service {
	return Service{ServiceID: nextServiceID.Add(1), CreateDateTime: time.Now().UTC().Format(time.RFC3339Nano)}
}

func NewSystem() *System { return &System{data: make(map[string]Service)} }

func (s *System) Get(key string) (Service, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

func (s *System) Set(key string, val Service) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.data[key]; ok && old == val {
		return
	}
	s.prepareWrite()
	s.data[key] = val
}

// prepareWrite must be called with mu exclusively locked.
func (s *System) prepareWrite() {
	if s.data == nil {
		s.data = make(map[string]Service)
	}
	s.version++
	s.cachedJSON = ""
}

// Add atomically registers a service; duplicate names preserve the original.
func (s *System) Add(serviceName string) (Service, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.data[serviceName]; ok {
		return existing, false
	}
	service := NewService()
	s.prepareWrite()
	s.data[serviceName] = service
	return service, true
}

func (s *System) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// Marshal returns an immutable JSON snapshot. Repeated reads without mutations
// reuse the encoding. A cold read is O(n log n) (JSON sorts map keys); a warm
// read is O(1). Sending the resulting bytes still costs O(response size).
func (s *System) Marshal() (string, error) {
	s.mu.RLock()
	cached := s.cachedJSON
	s.mu.RUnlock()
	if cached != "" {
		return cached, nil
	}

	s.marshalMu.Lock()
	defer s.marshalMu.Unlock()
	s.mu.RLock()
	if s.cachedJSON != "" {
		cached = s.cachedJSON
		s.mu.RUnlock()
		return cached, nil
	}
	version := s.version
	snapshot := s.snapshotLocked()
	s.mu.RUnlock()

	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return "", fmt.Errorf("marshal system: %w", err)
	}
	result := string(encoded)
	s.mu.Lock()
	// A concurrent write may have happened during encoding. Return the coherent
	// snapshot we read, but never publish it as the cache for the newer version.
	if s.version == version {
		s.cachedJSON = result
	}
	s.mu.Unlock()
	return result, nil
}

func (s *System) snapshotLocked() map[string]Service {
	out := make(map[string]Service, len(s.data))
	for key, value := range s.data {
		out[key] = value
	}
	return out
}

// GetAll returns a caller-owned snapshot, not the registry's mutable map.
func (s *System) GetAll() map[string]Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshotLocked()
}

func (s *System) Remove(serviceName string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[serviceName]; !ok {
		return false
	}
	s.prepareWrite()
	delete(s.data, serviceName)
	return true
}
