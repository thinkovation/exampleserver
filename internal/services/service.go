package services

import "context"

// Service represents a background service that can be started and stopped
type Service interface {
	Start(context.Context) error
}
