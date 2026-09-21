package core

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

func checkSnapshot(t *testing.T, system *System) {
	t.Helper()
	encoded, err := system.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]Service
	if err := json.Unmarshal([]byte(encoded), &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, system.GetAll()) {
		t.Fatalf("stale snapshot: %s", encoded)
	}
}

func TestZeroValueAndCacheInvalidation(t *testing.T) {
	var system System
	if encoded, err := system.Marshal(); encoded != "{}" || err != nil {
		t.Fatalf("zero snapshot = %q, %v", encoded, err)
	}
	first, _ := system.Add("api")
	checkSnapshot(t, &system)
	encoded, _ := system.Marshal()
	if _, created := system.Add("api"); created {
		t.Fatal("duplicate add")
	}
	if again, _ := system.Marshal(); again != encoded {
		t.Fatal("duplicate changed snapshot")
	}
	system.Set("api", first)
	checkSnapshot(t, &system)
	system.Set("api", NewCustomService(42, "2026-01-01T00:00:00Z"))
	checkSnapshot(t, &system)
	system.Set("worker", NewCustomService(43, "2026-01-01T00:00:00Z"))
	checkSnapshot(t, &system)
	if system.Remove("missing") {
		t.Fatal("removed missing")
	}
	checkSnapshot(t, &system)
	system.Remove("api")
	checkSnapshot(t, &system)
	system.Remove("worker")
	if encoded, _ := system.Marshal(); encoded != "{}" {
		t.Fatalf("not empty: %s", encoded)
	}
	// A previously returned string is immutable, even after cache invalidation.
	var old map[string]Service
	if err := json.Unmarshal([]byte(encoded), &old); err != nil || old["api"] != first {
		t.Fatal("old snapshot was changed")
	}
}

func TestZeroValueSet(t *testing.T) {
	var system System
	system.Set("api", NewCustomService(1, "now"))
	if system.Len() != 1 {
		t.Fatal("zero value Set failed")
	}
}

func TestConcurrentSingleWinner(t *testing.T) {
	system := NewSystem()
	var created, removed atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, ok := system.Add("same"); ok {
				created.Add(1)
			}
		}()
	}
	wg.Wait()
	if created.Load() != 1 {
		t.Fatalf("created %d times", created.Load())
	}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if system.Remove("same") {
				removed.Add(1)
			}
		}()
	}
	wg.Wait()
	if removed.Load() != 1 {
		t.Fatalf("removed %d times", removed.Load())
	}
}

func TestMarshalWhileMutating(t *testing.T) {
	system := NewSystem()
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			name := fmt.Sprintf("service-%d", worker)
			for i := 0; i < 200; i++ {
				system.Set(name, NewService())
				encoded, err := system.Marshal()
				var snapshot map[string]Service
				if err != nil || json.Unmarshal([]byte(encoded), &snapshot) != nil {
					t.Error("invalid concurrent snapshot")
					return
				}
				if i%3 == 0 {
					system.Remove(name)
				}
				system.GetAll()
			}
		}(worker)
	}
	wg.Wait()
	for i := 0; i < 8; i++ {
		system.Remove(fmt.Sprintf("service-%d", i))
	}
	system.Set("final", NewCustomService(9, "final"))
	checkSnapshot(t, system)
}

func TestConcurrentServiceIDs(t *testing.T) {
	ids := make(chan uint64, 2000)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ids <- NewService().ServiceID
			}
		}()
	}
	wg.Wait()
	close(ids)
	seen := make(map[uint64]bool)
	for id := range ids {
		if id == 0 || seen[id] {
			t.Fatalf("bad ID %d", id)
		}
		seen[id] = true
	}
}
