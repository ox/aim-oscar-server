package main

type ServiceManager struct {
	services map[uint16]Service
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		services: make(map[uint16]Service),
	}
}

func (sm *ServiceManager) RegisterService(family uint16, service Service) {
	sm.services[family] = service
}

func (sm *ServiceManager) GetService(family uint16) (Service, bool) {
	s, ok := sm.services[family]
	return s, ok
}
