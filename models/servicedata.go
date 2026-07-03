package data

import (
	"math/rand"
	"time"
)

type ServiceData struct {
	ServiceID      uint64 `json:"service_id"`
	CreateDateTime string `json:"create_datetime"`
}

func NewServiceData() ServiceData {
	return ServiceData{
		ServiceID:      uint64(rand.Intn(1000)),
		CreateDateTime: time.Now().Format(time.DateTime),
	}
}

func CreateServiceList() map[string]ServiceData {
	return make(map[string]ServiceData)
}

func AddService(sl map[string]ServiceData, serviceName string) map[string]ServiceData {
	data := NewServiceData()
	sl[serviceName] = data
	return sl
}

func RemoveService(sl map[string]ServiceData, serviceName string) map[string]ServiceData {
	for k := range sl {
		if k == serviceName {
			delete(sl, k)
		}
	}
	return sl
}
