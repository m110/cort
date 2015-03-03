package service

import (
	"github.com/m110/cort/resources/consul"
	"github.com/m110/cort/rpc"
)

type Service struct {
	name   string
	server *rpc.Server
	consul *consul.ConsulProxy
}

func NewService(name string, address string, port int) *Service {
	service := &Service{
		name: name,
	}

	service.server = rpc.NewServer(address, port)

	return service
}

func (s *Service) Start() error {
	var err error

	s.consul, err = consul.NewConsulProxy()
	if err != nil {
		return err
	}

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
	return s.consul.ServiceRegister(
		s.server.Id(),
		s.name,
		s.server.Address(),
		s.server.Port(),
		[]string{},
	)
}

func (s *Service) Deregister() error {
	return s.consul.ServiceDeregister(s.server.Id())
}
