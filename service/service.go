package service

import (
	"github.com/hashicorp/consul/api"
	"log"

	"github.com/m110/cort/rpc"
)

type Service struct {
	name   string
	server *rpc.Server
	consul *api.Client
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

	s.consul, err = api.NewClient(api.DefaultConfig())
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
	log.Println("Registering service with ID:", s.server.Id())

	agent := s.consul.Agent()
	err := agent.ServiceRegister(&api.AgentServiceRegistration{
		ID:      s.server.Id(),
		Name:    s.name,
		Tags:    []string{},
		Address: s.server.Address(),
		Port:    s.server.Port(),
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Deregister() error {
	log.Println("Deregistering service")

	agent := s.consul.Agent()
	err := agent.ServiceDeregister(s.server.Id())

	if err != nil {
		return err
	}

	return nil
}
