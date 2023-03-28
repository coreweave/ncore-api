package payloads

import (
	"context"
	"log"
)

// NewService creates an API service.
func NewService(
	db DB,
	payloadsDefaultPayloadId string,
	payloadsDefaultPayloadDirectory string,
) *Service {
	log.Printf("Starting Payloads service")
	return &Service{
		db:                              db,
		payloadsDefaultPayloadId:        payloadsDefaultPayloadId,
		payloadsDefaultPayloadDirectory: payloadsDefaultPayloadDirectory,
	}
}

// Service for the API.
type Service struct {
	db                              DB
	payloadsDefaultPayloadId        string
	payloadsDefaultPayloadDirectory string
}

// DB layer.
//
//go:generate mockgen --build_flags=--mod=mod -package payloads -destination mock_payloads_db_test.go . DB
type DB interface {
	// GetPayload returns a payload for a node.
	GetNodePayload(ctx context.Context, macAddress string) (*NodePayload, error)

	// GetAvailablePayloads returns a list of available payloads
	GetAvailablePayloads(ctx context.Context) []string

	// AddDefaultNodePayload adds a db entry with defaults for mac_address.
	AddDefaultNodePayload(ctx context.Context, config *NodePayloadDb) (*NodePayloadDb, error)

	// UpdateNodePayload updates the PayloadId for mac_address.
	UpdateNodePayload(ctx context.Context, config *NodePayloadDb) (*NodePayloadDb, error)

	// GetPayload returns a payload for a node.
	GetPayloadParameters(ctx context.Context, payloadId string) (interface{}, error)
}

// ValidationError is returned when there is an invalid parameter received.
type ValidationError struct {
	s string
}

func (e ValidationError) Error() string {
	return e.s
}
