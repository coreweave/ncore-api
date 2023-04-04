package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	sync "sync"
	"time"

	"github.com/coreweave/ncore-api/pkg/ipxe"
	"github.com/coreweave/ncore-api/pkg/payloads"
	"github.com/coreweave/ncore-api/pkg/systems"
)

// Server for the API.
type Server struct {
	HTTPAddress string
	Payloads    *payloads.Service
	Ipxe        *ipxe.Service
    Systems     *systems.Service
	http   *httpServer
	stopFn sync.Once
}

func middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	})
}

// Run starts the HTTP server.
func (s *Server) Run(ctx context.Context) (err error) {
	var ec = make(chan error, 1)
	ctx, cancel := context.WithCancel(ctx)
	s.http = &httpServer{
		ipxe:     s.Ipxe,
		payloads: s.Payloads,
		systems:  s.Systems,
	}
	go func() {
		err := s.http.Run(ctx, s.HTTPAddress)
		if err != nil {
			err = fmt.Errorf("HTTP server error: %w", err)
		}
		ec <- err
	}()

	// Wait for the services to exit.
	var es []string
	for i := 0; i < cap(ec); i++ {
		if err := <-ec; err != nil {
			es = append(es, err.Error())
			if ctx.Err() == nil {
				s.Shutdown(context.Background())
			}
		}
	}
	if len(es) > 0 {
		err = errors.New(strings.Join(es, ", "))
	}
	cancel()
	return err
}

// Shutdown HTTP server.
func (s *Server) Shutdown(ctx context.Context) {
	s.stopFn.Do(func() {
		s.http.Shutdown(ctx)
	})
}

type httpServer struct {
	ipxe       *ipxe.Service
	payloads   *payloads.Service
	systems    *systems.Service
	middleware func(http.Handler) http.Handler
	http       *http.Server
}

// Run HTTP server.
func (s *httpServer) Run(ctx context.Context, address string) error {
	handler := NewHTTPServer(s.ipxe, s.payloads, s.systems)

	if s.middleware != nil {
		log.Printf("Using middleware")
		handler = s.middleware(handler)
	}

	s.http = &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("HTTP server listening at %s\n", address)
	if err := s.http.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown HTTP server.
func (s *httpServer) Shutdown(ctx context.Context) {
	log.Println("shutting down HTTP server")
	if s.http != nil {
		if err := s.http.Shutdown(ctx); err != nil {
			log.Println("graceful shutdown of HTTP server failed")
		}
	}
}
