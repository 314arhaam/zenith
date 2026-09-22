package core

import (
	"sync"
	"testing"
)

func TestRegistrySnapshotAndDuplicate(t *testing.T) {
	s := NewSystem()
	first, created := s.Add("api")
	if !created {
		t.Fatal("first registration should succeed")
	}
	second, created := s.Add("api")
	if created || second != first {
		t.Fatalf("duplicate registration changed service: first=%+v second=%+v", first, second)
	}
	snapshot := s.GetAll()
	delete(snapshot, "api")
	if _, ok := s.Get("api"); !ok {
		t.Fatal("mutating GetAll snapshot mutated registry")
	}
}

func TestRegistryConcurrentUse(t *testing.T) {
	s := NewSystem()
	var wg sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			name := string(rune('a' + worker))
			for i := 0; i < 100; i++ {
				s.Add(name)
				s.Get(name)
				s.GetAll()
				s.Len()
				if i%5 == 0 {
					s.Remove(name)
				}
			}
		}(worker)
	}
	wg.Wait()
}
