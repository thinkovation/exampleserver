package services

import (
	"context"
	"log"
	"sync"
)

// Manager handles multiple background services
type Manager struct {
	services []Service
	wg       sync.WaitGroup
}

func NewManager() *Manager {
	return &Manager{
		services: make([]Service, 0),
	}
}

// AddService adds a service to be managed
func (m *Manager) AddService(service Service) {
	m.services = append(m.services, service)
}

// Start starts all services
func (m *Manager) Start(ctx context.Context) {
	for _, service := range m.services {
		m.wg.Add(1)
		go func(s Service) {
			defer m.wg.Done()
			if err := s.Start(ctx); err != nil && err != context.Canceled {
				log.Printf("service error: %v", err)
			}
		}(service)
	}
}

// Wait waits for all services to complete
func (m *Manager) Wait() {
	m.wg.Wait()
}
