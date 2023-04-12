package nodes

import (
	"context"
)

type Node struct {
	MacAddress string
	NodeId     string
	IpAddress  string
}

// GetNodePayload returns a NodePayload for mac_address.
func (s *Service) UpdateNodeStats(ctx context.Context, n *Node) (*Node, error) {
	if n.MacAddress == "" {
		return nil, ValidationError{"missing payload macAddress"}
	}
	return s.db.UpdateNodeStats(ctx, n)
}
