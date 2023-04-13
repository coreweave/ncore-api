package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coreweave/ncore-api/pkg/api"
	"github.com/coreweave/ncore-api/pkg/database"
	"github.com/coreweave/ncore-api/pkg/ipxe"
	"github.com/coreweave/ncore-api/pkg/nodes"
	"github.com/coreweave/ncore-api/pkg/payloads"
	"github.com/coreweave/ncore-api/pkg/postgres"
	"github.com/coreweave/ncore-api/pkg/s3"
)

type pgConfig struct {
	pguser     string
	pgpassword string
	pghost     string
	pgport     string
	pgdatabase string
}

func newPGConfigFromEnv(envPrefix string) *pgConfig {
	appendPrefix := func(s string) string { return envPrefix + "_" + s }
	return &pgConfig{
		pguser:     os.Getenv(appendPrefix("PGUSER")),
		pgpassword: os.Getenv(appendPrefix("PGPASSWORD")),
		pghost:     os.Getenv(appendPrefix("PGHOST")),
		pgport:     os.Getenv(appendPrefix("PGPORT")),
		pgdatabase: os.Getenv(appendPrefix("PGDATABASE")),
	}
}

func (p *pgConfig) connString() string {
	formatString := `user=%s
password=%s
host=%s
port=%s
dbname=%s`

	return fmt.Sprintf(formatString,
		p.pguser,
		p.pgpassword,
		p.pghost,
		p.pgport,
		p.pgdatabase,
	)
}

func main() {
	var (
		httpAddr,
		s3Host,
		ipxeTemplateFile,
		ipxeDefaultImage,
		ipxeDefaultImageTag,
		ipxeDefaultImageType,
		ipxeDefaultBucket,
		payloadsDefaultPayloadId,
		payloadsDefaultPayloadDirectory string
	)

	flag.StringVar(&httpAddr, "http", "localhost:8080", "HTTP service address to listen for incoming requests on")
	flag.StringVar(&s3Host, "s3.host", "https://accel-object.ord1.coreweave.com", "S3 Storage endpoint")
	flag.StringVar(&ipxeTemplateFile, "ipxe.template", "pkg/ipxe/templates/template_https.ipxe", "Relative path to ipxe template file")
	flag.StringVar(&ipxeDefaultImage, "ipxe.default.image", "default", "Default image used when database is unavailable or no entry found for macAddress")
	flag.StringVar(&ipxeDefaultImageTag, "ipxe.default.imageTag", "default", "Default image_tag entry added for node when no entry found for macAddress")
	flag.StringVar(&ipxeDefaultImageType, "ipxe.default.imageType", "default", "Default image_type entry added for node when no entry found for macAddress")
	flag.StringVar(&ipxeDefaultBucket, "ipxe.default.bucket", "default", "Default image used when database is unavailable or no entry found for macAddress")
	flag.StringVar(&payloadsDefaultPayloadId, "payloads.default.payloadId", "default", "Default PayloadId assigned when no entry found for macAddress")
	flag.StringVar(&payloadsDefaultPayloadDirectory, "payloads.default.payloadDirectory", "default", "Default PayloadDirectory assigned when no entry found for macAddress")

	flag.Parse()
	pgxLogLevel, err := database.LogLevelFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	payloadsPGConfig := newPGConfigFromEnv("PAYLOADS")
	pgPoolPayloads, err := database.NewPGXPool(context.Background(), payloadsPGConfig.connString(), &database.PGXStdLogger{}, pgxLogLevel)
	if err != nil {
		log.Fatal(err)
	}
	defer pgPoolPayloads.Close()

	ipxePGConfig := newPGConfigFromEnv("IPXE")
	pgPoolIpxe, err := database.NewPGXPool(context.Background(), ipxePGConfig.connString(), &database.PGXStdLogger{}, pgxLogLevel)
	if err != nil {
		log.Fatal(err)
	}
	defer pgPoolIpxe.Close()

	nodesPGConfig := newPGConfigFromEnv("NODES")
	pgPoolNodes, err := database.NewPGXPool(context.Background(), nodesPGConfig.connString(), &database.PGXStdLogger{}, pgxLogLevel)
	if err != nil {
		log.Fatal(err)
	}
	defer pgPoolNodes.Close()

	log.Printf("s3Host: %v", s3Host)
	s3Svc := s3.NewClient(s3Host)
	presignClient := s3.NewPresigner(*s3Svc)

	s := &api.Server{
		Payloads: payloads.NewService(
			&postgres.DB{
				Postgres: pgPoolPayloads,
			},
			payloadsDefaultPayloadId,
			payloadsDefaultPayloadDirectory,
		),
		Ipxe: ipxe.NewService(
			&postgres.DB{
				Postgres: pgPoolIpxe,
			},
			*s3Svc,
			presignClient,
			ipxeTemplateFile,
			ipxeDefaultImage,
			ipxeDefaultImageTag,
			ipxeDefaultImageType,
			ipxeDefaultBucket,
		),
		Nodes: nodes.NewService(
			&postgres.DB{
				Postgres: pgPoolNodes,
			},
		),
		HTTPAddress: httpAddr,
	}
	ec := make(chan error, 1)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	go func() {
		ec <- s.Run(context.Background())
	}()

	// Waits for an internal error that shutdowns the server.
	// Otherwise, wait for a SIGINT or SIGTERM and tries to shutdown the server gracefully.
	// After a shutdown signal, HTTP requests taking longer than the specified grace period are forcibly closed.
	select {
	case err = <-ec:
	case <-ctx.Done():
		haltCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		s.Shutdown(haltCtx)
		stop()
		err = <-ec
	}
	if err != nil {
		log.Fatal(err)
	}
}
