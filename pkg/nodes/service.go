package nodes

import (
	"context"
	"log"
)

func NewService(db DB) *Service {
	log.Printf("Starting Nodes service")
	return &Service{
		db: db,
	}
}

type Service struct {
	db DB
}

type DB interface {
	UpdateNodeStats(ctx context.Context, n *Node) (*Node, error)
}

type ValidationError struct {
	s string
}

func (e ValidationError) Error() string {
	return e.s
}
