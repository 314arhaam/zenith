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

func AddService(sl ServiceData, serviceName string) ServiceData {
	data := NewService()
	sl[serviceName] = data
	return sl
}

func RemoveService(sl ServiceData, serviceName string) ServiceData {
	for k := range sl {
		if k == serviceName {
			delete(sl, k)
		}
	}
	return sl
}
