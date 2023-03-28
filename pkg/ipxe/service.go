package ipxe

import (
	"context"
	"log"

	"github.com/coreweave/ncore-api/pkg/s3"
)

// NewService creates an API service.
func NewService(
	db DB,
	s3Svc s3.S3Svc,
	s3Presigner s3.HttpPresigner,
	ipxeTemplateFile string,
	ipxeDefaultImage string,
	ipxeDefaultImageTag string,
	ipxeDefaultImageType string,
	ipxeDefaultBucket string,
) *Service {
	log.Printf("Starting Ipxe service")
	return &Service{
		db:                   db,
		s3Svc:                s3Svc,
		s3Presigner:          s3Presigner,
		ipxeTemplateFile:     ipxeTemplateFile,
		ipxeDefaultImage:     ipxeDefaultImage,
		ipxeDefaultImageTag:  ipxeDefaultImageTag,
		ipxeDefaultImageType: ipxeDefaultImageType,
		ipxeDefaultBucket:    ipxeDefaultBucket,
	}
}

// Service for the API.
type Service struct {
	db                   DB
	s3Svc                s3.S3Svc
	s3Presigner          s3.HttpPresigner
	ipxeTemplateFile     string
	ipxeDefaultImage     string
	ipxeDefaultImageTag  string
	ipxeDefaultImageType string
	ipxeDefaultBucket    string
}

// DB layer.
//
//go:generate mockgen --build_flags=--mod=mod -package ipxe -destination mock_ipxe_db_test.go . DB
type DB interface {
	// GetIpxe returns an IpxeConfig for a macAddress.
	GetIpxeDbConfig(ctx context.Context, macAddress string) (*IpxeDbConfig, error)
	CreateNodeIpxeConfig(ctx context.Context, config *IpxeNodeDbConfig) (*IpxeNodeDbConfig, error)
	CreateIpxeImage(ctx context.Context, config *IpxeDbConfig) (*IpxeConfig, error)
	DeleteIpxeImage(ctx context.Context, config *IpxeDbDeleteConfig) (*IpxeDbConfig, error)
}

// ValidationError is returned when there is an invalid parameter received.
type ValidationError struct {
	s string
}

func (e ValidationError) Error() string {
	return e.s
}
