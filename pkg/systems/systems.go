package systems

import (
	"context"
	//"time"
)

type SystemId struct {
	MacAddress  string
	Id			string
}

func (s *Service) UpdateSystemStatus(ctx context.Context, macAddress string) (*SystemId, error) {
	if macAddress == "" {
		return nil, ValidationError{"missing system macAddress"}
	}
	return s.db.UpdateSystemStatus(ctx, macAddress)
}

// ValidationError is returned when there is an invalid parameter received.
type ValidationError struct {
	s string
}

func (e ValidationError) Error() string {
	return e.s
}