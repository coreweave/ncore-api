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
	// GetNodePayloads reads all payloads for mac_address and returns them as a list.
	GetNodePayloads(ctx context.Context, macAddress string) ([]*NodePayload, error)

	// GetSubnetDefaultPayload accepts an ip address string and checks if payloads.subnet_default_payloads table
	// contains a payload_id for the corresponding cidr
	// Returns a Payload
	GetSubnetDefaultPayload(ctx context.Context, ipAddress string) (*Payload, error)

	// GetAvailablePayloads returns a list of available payloads
	GetAvailablePayloads(ctx context.Context) []string

	// AddNodePayload adds a NodePayloadDb entry for mac_address
	AddNodePayload(ctx context.Context, config *NodePayloadDb) ([]*NodePayload, error)

	// UpdateNodePayload updates the PayloadId for mac_address.
	UpdateNodePayload(ctx context.Context, config *NodePayloadDb) ([]*NodePayload, error)

	// DeleteNodePayload deletes the PayloadId for mac_address/payload tuple.
	DeleteNodePayload(ctx context.Context, config *NodePayloadDb) ([]*NodePayload, error)

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
