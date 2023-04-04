package systems

import (
	"context"
	"log"
)

// NewService creates an API service.
func NewService(db DB) *Service {
	log.Printf("Starting Systems service")
	return &Service{db: db}
}

// Service for the API.
type Service struct {
  db DB
}

type DB interface {
	// status check in to db from system
	UpdateSystemStatus(ctx context.Context, macAddress string) (*SystemId, error)
}