package main

import "aim-oscar/services"

type ServiceManager struct {
	services map[uint16]services.Service
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		services: make(map[uint16]services.Service),
	}
}

func (sm *ServiceManager) RegisterService(family uint16, service services.Service) {
	sm.services[family] = service
}

func (sm *ServiceManager) GetService(family uint16) (services.Service, bool) {
	s, ok := sm.services[family]
	return s, ok
}
