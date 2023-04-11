package systems

import (
	"context"
	//"time"
)

type SystemStatusDb struct {
	MacAddress string
	NodeId     string
	IpAddress  string
}

func (s *Service) UpdateSystemStatus(ctx context.Context, stat *SystemStatusDb) (*SystemStatusDb, error) {
	if stat.MacAddress == "" {
		return nil, ValidationError{"missing system macAddress"}
	}
	return s.db.UpdateSystemStatus(ctx, stat)
}

// ValidationError is returned when there is an invalid parameter received.
type ValidationError struct {
	s string
}

func (e ValidationError) Error() string {
	return e.s
}
