package grpc

import (
	"context"
	"log"
	"net"

	"github.com/omerkaya1/grpc-file-stream/internal"
	"github.com/omerkaya1/grpc-file-stream/internal/domain"
	"github.com/omerkaya1/grpc-file-stream/internal/grpc/api"
	"google.golang.org/grpc"
)

// Server is an object that holds everything related to the file-streamer server
type Server struct {
	store domain.Storage
	log   internal.MetaLogger
	cfg   internal.Config
}

// NewServer returns a new instance of the file-streamer server
func NewServer(s domain.Storage, l internal.MetaLogger, cfg internal.Config) *Server {
	return &Server{
		store: s,
		cfg:   cfg,
		log:   l,
	}
}

// Run runs the file-streamer server
func (s *Server) Run(ctx context.Context) error {
	server := grpc.NewServer()
	l, err := net.Listen("tcp", s.cfg.Host+":"+s.cfg.Port)
	if err != nil {
		log.Fatalf("%s", err)
	}

	api.RegisterFileStreamerServiceServer(server, s)

	go func(c context.Context) {
		<-c.Done()
		log.Print("Context interrupt received")
		if c.Err() != nil && c.Err() != context.Canceled {
			s.log.Printf("Context error: %s", c.Err())
		}
		server.GracefulStop()
		s.log.Print("Graceful shutdown performed. Bye!")
		return
	}(ctx)

	s.log.Printf("Server initialisation is completed. Server address: %s:%s", s.cfg.Host, s.cfg.Port)
	if err := server.Serve(l); err != nil {
		return err
	}
	return nil
}
