package consul

import (
	"github.com/hashicorp/consul/api"
	"log"
)

type ConsulProxy struct {
	client *api.Client

	lastIndex struct {
		service uint64
	}
}

func NewConsulProxy() (*ConsulProxy, error) {
	var err error

	consulProxy := &ConsulProxy{}

	consulProxy.client, err = api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}

	return consulProxy, nil
}

func (c *ConsulProxy) ServiceNodes(service string) ([]string, error) {
	tags := ""
	opts := &api.QueryOptions{}

	if c.lastIndex.service > 0 {
		opts.WaitIndex = c.lastIndex.service
	}

	nodes, meta, err := c.client.Catalog().Service(service, tags, opts)
	if err != nil {
		return nil, err
	}

	c.lastIndex.service = meta.LastIndex

	var newNodes []string
	for _, node := range nodes {
		newNodes = append(newNodes, node.ServiceID)
	}

	return newNodes, nil
}

func (c *ConsulProxy) ServiceRegister(id, name, address string, port int, tags []string) error {
	log.Println("Registering service:", id)

	return c.client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Address: address,
		Port:    port,
		Tags:    tags,
	})
}

func (c *ConsulProxy) ServiceDeregister(id string) error {
	log.Println("Deregistering service:", id)

	return c.client.Agent().ServiceDeregister(id)
}
