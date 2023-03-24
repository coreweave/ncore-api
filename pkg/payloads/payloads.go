package payloads

import (
	"context"
	"time"
)

// NodePayload with directory for mac_address.
type NodePayload struct {
	PayloadId        string
	PayloadDirectory string
	MacAddress       string
}

// NodePayload payloads.node_payloads entry for mac_address.
type NodePayloadDb struct {
	PayloadId        string
	MacAddress       string
}

// Payload for payload_id.
type Payload struct {
	PayloadId        string
	PayloadDirectory string
	CreatedAt        time.Time
	ModifiedAt       time.Time
}

// PayloadSchema for payload_id.
type PayloadSchema struct {
	Id              string
	PayloadSchemaId string
	ParameterName   string
	CreatedAt       time.Time
	ModifiedAt      time.Time
}

type PayloadParameter struct {
	Parameter string
	Value     string
}

type PayloadParameters struct {
	Id         string
	PayloadId  string
	Parameters map[string]interface{}
	CreatedAt  time.Time
	ModifiedAt time.Time
}

// GetNodePayload returns a NodePayload for mac_address.
func (s *Service) GetNodePayload(ctx context.Context, macAddress string) (*NodePayload, error) {
	if macAddress == "" {
		return nil, ValidationError{"missing payload macAddress"}
	}
	return s.db.GetNodePayload(ctx, macAddress)
}

// AddDefaultNodePayload adds a db entry with defaults for mac_address.
func (s *Service) AddDefaultNodePayload(ctx context.Context, macAddress string) (*NodePayloadDb, error) {
	if macAddress == "" {
		return nil, ValidationError{"missing payload macAddress"}
	}
	var npd = &NodePayloadDb{
		PayloadId: s.payloadsDefaultPayloadId,
		MacAddress: macAddress,
	}
	return s.db.AddDefaultNodePayload(ctx, npd)
}

// GetAvailablePayloads returns a list of available payloads
func (s *Service) GetAvailablePayloads(ctx context.Context) ([]string) {
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
