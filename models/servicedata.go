package data

import (
	"math/rand"
	"time"
)

type Service struct {
	ServiceID      uint64 `json:"service_id"`
	CreateDateTime string `json:"create_datetime"`
}

type ServiceData map[string]Service

func NewService() Service {
	return Service{
		ServiceID:      uint64(rand.Intn(1000)),
		CreateDateTime: time.Now().Format(time.DateTime),
	}
}

func CreateServiceData() ServiceData {
	return make(ServiceData)
}

func (s *ServiceData) Add(serviceName string) {
	data := NewService()
	(*s)[serviceName] = data
}

func (s *ServiceData) Remove(serviceName string) {
	for k := range *s {
		if k == serviceName {
			delete(*s, k)
		}
	}
}
