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
	if d.Len() != 1 {
		t.Fatal("Add method doesn't work")
	}
	d.Add("x")
	d.Add("y")
	log.Printf("Length: %d", d.Len())
	d.Remove("x")
	if d.Len() != 2 {
		t.Fatal("Remove method doesn't work")
	}
}

func TestBehave(t *testing.T) {
	log.Println("*** REMOVE TEST ***")
	d1 := NewServiceData()
	d2 := NewServiceData()
	for i := 0; i < 3; i++ {
		dt := time.Now().Format(time.DateTime)
		name := "test-service-" + strconv.Itoa(i+1)
		d1.Set(name, Service{
			ServiceID:      uint64(10 + i),
			CreateDateTime: dt,
		})
		d2.Set(name, Service{
			ServiceID:      uint64(10 + i),
			CreateDateTime: dt,
		})
	}
	d1m, _ := d1.Marshal()
	d2m, _ := d2.Marshal()
	if string(d1m) != string(d2m) {
		t.Error("Inital data are not equal")
		t.Fail()
	}
	d1.Remove("test-service-2")
	d1m, _ = d1.Marshal()
	if string(d1m) == string(d2m) {
		t.Error("Removal failed")
		t.Fail()
	}
}
