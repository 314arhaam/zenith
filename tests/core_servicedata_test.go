package tests

import (
	"sync"
	"testing"
	"zenith/core"
)

func TestSystemAddGetRemove(t *testing.T) {
	system := core.NewSystem()
	service, created := system.Add("api")
	if !created {
		t.Fatal("first add should create the service")
	}
	if service.ServiceID == 0 || service.CreateDateTime == "" {
		t.Fatalf("invalid generated service: %+v", service)
	}

	got, ok := system.Get("api")
	if !ok || got != service {
		t.Fatalf("Get() = %+v, %v; want %+v, true", got, ok, service)
	}
	if system.Len() != 1 {
		t.Fatalf("Len() = %d; want 1", system.Len())
	}
	if !system.Remove("api") {
		t.Fatal("Remove existing service should return true")
	}
	if system.Remove("api") {
		t.Fatal("Remove missing service should return false")
	}
}

func TestSystemAddRejectsDuplicateWithoutOverwriting(t *testing.T) {
	system := core.NewSystem()
	first, created := system.Add("api")
	if !created {
		t.Fatal("first add should succeed")
	}
	second, created := system.Add("api")
	if created {
		t.Fatal("duplicate add should not create a new service")
	}
	if second != first {
		t.Fatalf("duplicate add changed service: first=%+v second=%+v", first, second)
	}
}

func TestSystemGetAllReturnsSnapshot(t *testing.T) {
	system := core.NewSystem()
	original := core.NewCustomService(42, "2026-01-01T00:00:00Z")
	system.Set("api", original)

	snapshot := system.GetAll()
	snapshot["api"] = core.NewCustomService(99, "changed")
	delete(snapshot, "api")

	got, ok := system.Get("api")
	if !ok || got != original {
		t.Fatalf("mutating snapshot changed registry: %+v, %v", got, ok)
	}
}

func TestSystemConcurrentAccess(t *testing.T) {
	system := core.NewSystem()
	var wg sync.WaitGroup
	for worker := 0; worker < 20; worker++ {
		worker := worker
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				name := string(rune('a'+worker%20)) + "-service"
				system.Add(name)
				system.Get(name)
				system.GetAll()
				system.Len()
				if i%7 == 0 {
					system.Remove(name)
				}
			}
		}()
	}
	wg.Wait()
}

func TestServiceIDsAreUnique(t *testing.T) {
	seen := make(map[uint64]struct{})
	for i := 0; i < 1000; i++ {
		service := core.NewService()
		if _, exists := seen[service.ServiceID]; exists {
			t.Fatalf("duplicate service ID: %d", service.ServiceID)
		}
		seen[service.ServiceID] = struct{}{}
	}
}
