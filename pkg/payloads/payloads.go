package payloads

import (
	"context"
)

// NodePayload with directory for mac_address.
type NodePayload struct {
	PayloadId        string `json:"payload_id"`
	PayloadDirectory string `json:"payload_directory"`
	MacAddress       string `json:"mac_address"`
}

// NodePayload payloads.node_payloads entry for mac_address.
type NodePayloadDb struct {
	PayloadId  string `json:"payload_id"`
	MacAddress string `json:"mac_address"`
}

// Payload for payload_id.
type Payload struct {
	PayloadId        string
	PayloadDirectory string
}

// GetNodePayload returns a NodePayload for mac_address.
func (s *Service) GetNodePayload(ctx context.Context, macAddress string) (*NodePayload, error) {
	if macAddress == "" {
		return nil, ValidationError{"missing payload macAddress"}
	}
	return s.db.GetNodePayload(ctx, macAddress)
}

// GetSubnetDefaultPayload accepts an ip address string and checks if payloads.subnet_default_payloads table
// contains a payload_id for the corresponding cidr
// Returns a Payload
func (s *Service) GetSubnetDefaultPayload(ctx context.Context, ipAddress string) (*Payload, error) {
	if ipAddress == "" {
		return nil, ValidationError{"missing payload ipAddress"}
	}
	return s.db.GetSubnetDefaultPayload(ctx, ipAddress)
}

// AddNodePayload adds a NodePayload entry.
// Returns a NodePayload
func (s *Service) AddNodePayload(ctx context.Context, nodePayloadDb *NodePayloadDb) error {
	if nodePayloadDb.PayloadId == "" {
		return ValidationError{"missing nodePayloadDb PayloadId"}
	}
	if nodePayloadDb.MacAddress == "" {
		return ValidationError{"missing nodePayloadDb payload MacAddress"}
	}

	return s.db.AddNodePayload(ctx, nodePayloadDb)
}

// GetDefaultPayload returns the default Payload from flags.
func (s *Service) GetDefaultPayload(ctx context.Context) *Payload {
	var p = &Payload{
		PayloadId:        s.payloadsDefaultPayloadId,
		PayloadDirectory: s.payloadsDefaultPayloadDirectory,
	}
	return p
}

// GetAvailablePayloads returns a list of available payloads
func (s *Service) GetAvailablePayloads(ctx context.Context) []string {
	return s.db.GetAvailablePayloads(ctx)
}

// UpdateNodePayload updates the PayloadId for mac_address.
func (s *Service) UpdateNodePayload(ctx context.Context, config *NodePayloadDb) (*NodePayloadDb, error) {
	return s.db.UpdateNodePayload(ctx, config)
}

// GetPayloadParameters returns a PayloadSchema for PayloadId.
func (s *Service) GetPayloadParameters(ctx context.Context, payloadId string) (interface{}, error) {
	if payloadId == "" {
		return nil, ValidationError{"missing payloadId"}
	}
	return s.db.GetPayloadParameters(ctx, payloadId)
}
