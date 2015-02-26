package rpc

import (
	zmq "github.com/pebbe/zmq4"

	"github.com/m110/cort/discovery"
)

type Client struct {
	service string
	socket  *zmq.Socket
}

func NewClient(service string) *Client {
	client := &Client{
		service: service,
	}

	return client
}

func (c *Client) Connect() error {
	var err error

	discovery.Start(c.service)

	c.socket, err = zmq.NewSocket(zmq.REQ)
	if err != nil {
		return err
	}

	c.socket.Connect("inproc://" + c.service)

	return nil
}

func (c *Client) Call(request string) (string, error) {
	_, err := c.socket.SendMessage(request)
	if err != nil {
		return "", err
	}

	response, err := c.socket.RecvMessage(0)
	if err != nil {
		return "", err
	}

	return response[0], nil
}
