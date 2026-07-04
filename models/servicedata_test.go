package data

import (
	"encoding/json"
	"log"
	"strconv"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	log.Println("*** CREATE TEST ***")
	d := NewService()
	jsonData, _ := json.Marshal(d)
	log.Println(string(jsonData))
}

func TestMarshal(t *testing.T) {
	log.Println("*** MARSHAL INDENT TEST ***")
	d := NewService()
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		t.Error("Inital data are not equal")
		t.Fail()
	} else {
		log.Print(string(b))
	}
}

func TestAddService(t *testing.T) {
	d := NewServiceData()
	d.Add("test-service-main")
	if len(d) != 1 {
		t.Fatal("Add method doesn't work")
	}
	d.Add("x")
	d.Add("y")
	log.Printf("Length: %d", len(d))
	d.Remove("x")
	if len(d) != 2 {
		t.Fatal("Remove method doesn't work")
	}
}

func TestBehave(t *testing.T) {
	log.Println("*** REMOVE TEST ***")
	d1 := make(ServiceData)
	d2 := make(ServiceData)
	for i := 0; i < 3; i++ {
		dt := time.Now().Format(time.DateTime)
		name := "test-service-" + strconv.Itoa(i+1)
		d1[name] = Service{
			ServiceID:      uint64(10 + i),
			CreateDateTime: dt,
		}
		d2[name] = Service{
			ServiceID:      uint64(10 + i),
			CreateDateTime: dt,
		}
	}
	d1m, _ := json.Marshal(d1)
	d2m, _ := json.Marshal(d2)
	if string(d1m) != string(d2m) {
		t.Error("Inital data are not equal")
		t.Fail()
	}
	d1.Remove("test-service-2")
	d1m, _ = json.Marshal(d1)
	if string(d1m) == string(d2m) {
		t.Error("Removal failed")
		t.Fail()
	}
}
