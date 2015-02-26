package rpc

import (
	"code.google.com/p/go-uuid/uuid"
	zmq "github.com/pebbe/zmq4"
)

type Server struct {
	id            string
	uri           string
	running       bool
	remoteSocket  *zmq.Socket
	workersSocket *zmq.Socket
}

func NewServer(uri string) *Server {
	server := &Server{
		id:  uuid.New(),
		uri: uri,
	}

	return server
}

func (s *Server) Start() error {
	var err error

	s.remoteSocket, err = zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		return err
	}

	s.remoteSocket.Bind(s.uri)
	if err != nil {
		return err
	}

	s.workersSocket, err = zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		return err
	}

	err = s.workersSocket.Bind("inproc://" + s.id)
	if err != nil {
		return err
	}

	poller := zmq.NewPoller()
	poller.Add(s.remoteSocket, zmq.POLLIN)
	poller.Add(s.workersSocket, zmq.POLLIN)

	s.running = true

	for s.running {
		polled, err := poller.Poll(100)
		if err != nil {
			// TODO Log a warning
		}

		if len(polled) > 0 {
			s.handleSockets(polled)
		}
	}

	return nil
}

func (s *Server) Stop() {
	if s.running {
		s.running = false
	}
}

func (s *Server) handleSockets(polled []zmq.Polled) {
	for _, p := range polled {
		switch socket := p.Socket; socket {
		case s.remoteSocket:
			s.handleRemoteSocket()
		case s.workersSocket:
			s.handleWorkersSocket()
		}
	}
}

func (s *Server) handleRemoteSocket() {
	// TODO Receive message and route it to one of available workers
}

func (s *Server) handleWorkersSocket() {
	// TODO Receive message and route it back to the remote client
}
