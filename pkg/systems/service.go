package systems

import (
	"context"
	"log"
)

type Node struct {
	MacAddress string
	NodeId     string
	IpAddress  string
}

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
	UpdateSystemStatus(ctx context.Context, node *SystemStatusDb) (*SystemStatusDb, error)
}
