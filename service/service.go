package service

import (
	"github.com/m110/cort/rpc"
)

type Service struct {
	name   string
	server *rpc.Server
}

func NewService(name string, uri string) *Service {
	service := &Service{
		name:   name,
		server: rpc.NewServer(uri),
	}

	return service
}

func (s *Service) Start() error {
	var err error

	err = s.Register()
	if err != nil {
		return err
	}

	err = s.server.Start()
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Stop() error {
	var err error

	err = s.Deregister()
	if err != nil {
		return err
	}

	err = s.server.Stop()
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Register() error {
	// TODO Register the service to be discovered by other services
	return nil
}

func (s *Service) Deregister() error {
	// TODO Deregister the service
	return nil
}
