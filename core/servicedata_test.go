package core_test

import (
	"testing"
	"time"
	"zenith/core"
)

func TestCreateService(t *testing.T) {
	s0 := core.NewService()
	t.Logf("[*] service created: %+v", s0)
	time.Sleep(1 * time.Second)
	s1 := core.NewCustomService(123, time.Now().Format(time.DateTime))
	t.Logf("[*] service created: %+v", s1)
}

func TestSystem(t *testing.T) {
	system := core.NewSystem()
	system.Add("new-test-service")
	if system.Len() != 1 {
		t.Fatalf("[x] Add() doesn't work: %d", system.Len())
	}
	t.Logf("[*] %+v", system.GetAll())
	if exists := system.Remove("new-test-service"); !exists {
		t.Fatalf("[x] error in remove, service exists but remove returns false")
	}
	if system.Len() != 0 {
		t.Fatalf("[x] Remove() doesn't work: %d", system.Len())
	}
	t.Logf("[*] %+v", system.GetAll())
	if exists := system.Remove("new-non-existant-service"); exists {
		t.Fatalf("[x] error in remove, service does not exist but remove returns true")
	}
	s0 := core.NewService()
	system.Set("new-test-set", s0)
	if system.Len() != 1 {
		t.Fatalf("[x] Set() doesn't work: %d", system.Len())
	}
	t.Logf("[*] %+v", system.GetAll())
}
