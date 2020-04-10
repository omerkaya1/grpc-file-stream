package domain

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/omerkaya1/grpc-file-stream/internal"
)

type (
	// Storage .
	Storage interface {
		Create(string, io.Reader) error
	}
	// Store .
	Store struct {
		location string
		log      internal.SimpleLogger
	}
)

func NewStore(location string, log internal.SimpleLogger) (*Store, error) {
	if _, err := os.Stat(location); os.IsNotExist(err) {
		return nil, fmt.Errorf("could not resolve path '%s', error: %s", location, err)
	}
	return &Store{
		location: location,
		log:      log,
	}, nil
}

func (s *Store) Create(name string, r io.Reader) error {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	s.log.Print("[debug] started writing a file: ", name)
	s.log.Print("[debug] file size: ", len(buf)/1024^2, " MB")
	if err := ioutil.WriteFile(s.location+name, buf, os.ModePerm); err != nil {
		s.log.Print("[result] failure [message] ", err.Error())
	}
	return nil
}
