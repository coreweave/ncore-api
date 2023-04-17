package nodes

import (
	"context"
)

type Node struct {
	MacAddress string
	Hostname   string
	IpAddress  string
}

// Update nodes stats by provided data payload.
func (s *Service) UpdateNodeStats(ctx context.Context, n *Node) (*Node, error) {
	if n.MacAddress == "" {
		return nil, ValidationError{"Missing macAddress"}
	}
	return s.db.UpdateNodeStats(ctx, n)
}
