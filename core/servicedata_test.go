package core

import (
	"testing"
	"time"
)

func TestServiceData(t *testing.T) {
	testKey := "test"
	sd := NewSystem()
	// Add
	sd.Add(testKey)
	t.Logf("[*] Data added: %s", testKey)
	// Get
	s, ok := sd.Get(testKey)
	if !ok {
		t.Fatal("Get method on ServiceData failed.")
	}
	// Len
	t.Logf("[*] Get method ok: %s %v", testKey, s)
	if sd.Len() != 1 {
		t.Fatal("Len method on ServiceData failed")
	}
	t.Log("[*] Len ok")
	// Marshal
	if val, err := sd.Marshal(); err != nil {
		t.Fatal("Marshall method on ServiceData failed")
	} else {
		t.Logf("[*] Marshal ok %s", val)
	}
	// Remove true
	nonExistsTestKey := testKey + "_prefix"
	if ok := sd.Remove(testKey); !ok {
		t.Fatal("Remove method on ServiceData failed")
	}
	// Remove false
	if ok := sd.Remove(nonExistsTestKey); !ok {
		t.Log("[*] Non existent data returned False")
	} else {
		t.Fatal("Remove method on ServiceData failed. Non existent data")
	}
	// GetAll
	curTime := time.Now().Format(time.DateTime)
	mockService := NewCustomService(1, curTime)
	data1 := NewSystem()
	data1.Set("test-02", mockService)
	data2 := make(map[string]service)
	data2["test-02"] = mockService
	if data1.GetAll()["test-02"] != data2["test-02"] {
		t.Fatal("GetAll method on ServiceData failed")
	}
	t.Logf("[*] GetAll ok %v", data1.GetAll())
}
