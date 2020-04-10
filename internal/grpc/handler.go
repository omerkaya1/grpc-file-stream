package grpc

import (
	"bytes"
	"io"
	"os"
	"os/signal"

	"github.com/omerkaya1/grpc-file-stream/internal/grpc/api"
)

func (s *Server) StoreFile(r api.FileStreamerService_StoreFileServer) error {
	// Service shutdown can be implemented through a channel, but I'm too lazy
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	var err error
	nb := bytes.NewBuffer([]byte{})
	var filename string

READ:
	for {
		select {
		case <-stop:
			s.log.Print("interrupt signal received: aborting process")
			break READ
		case <-r.Context().Done():
			s.log.Printf("context done: %s", r.Context().Err())
			err = r.Context().Err()
			break READ
		default:
			// Receive a stream
			stream, err := r.Recv()
			// TODO: fix file naming issue
			filename = "test_file.mov"
			if err == io.EOF {
				if err = s.store.Create(filename, nb); err != nil {
					return r.SendAndClose(composeResponse(err))
				}
				return r.SendAndClose(composeResponse(err))
			}
			if err != nil {
				return r.SendAndClose(composeResponse(err))
			}
			// Write to buffer
			// TODO: try to use io.Pipe in the next iteration
			if _, err := nb.Write(stream.GetContent()[:stream.GetReadLimit()]); err != nil {
				if err != nil {
					return r.SendAndClose(composeResponse(err))
				}
			}
		}
	}
	return r.SendAndClose(composeResponse(err))
}

func composeResponse(err error) *api.Response {
	switch err {
	case io.EOF, nil:
		return &api.Response{
			Message: "Successfully finished upload",
			Code:    0,
		}
	default:
		return &api.Response{
			Message: "An error occurred: " + err.Error(),
			Code:    1,
		}
	}
}
