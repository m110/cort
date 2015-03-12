package rpc

import (
	"errors"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"log"
	"strconv"
)

type Server struct {
	id      string
	address string
	port    int
	workers int
	running bool

	remoteSocket  *zmq.Socket
	workersSocket *zmq.Socket

	remoteEndpoint  string
	workersEndpoint string
}

func NewServer(address string, port int, workers int) *Server {
	id := fmt.Sprintf("tcp://%s:%d", address, port)

	server := &Server{
		id:              id,
		address:         address,
		port:            port,
		workers:         workers,
		remoteEndpoint:  fmt.Sprintf("tcp://%s:%d", address, port),
		workersEndpoint: "inproc://" + id,
	}

	return server
}

func (s *Server) Start() error {
	var err error

	log.Println("Starting the server")

	s.remoteSocket, err = zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		return err
	}

	err = s.remoteSocket.SetIdentity(s.remoteEndpoint)
	if err != nil {
		return err
	}

	s.remoteSocket.Bind(s.remoteEndpoint)
	if err != nil {
		return err
	}

	s.workersSocket, err = zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		return err
	}

	err = s.workersSocket.Bind(s.workersEndpoint)
	if err != nil {
		return err
	}

	err = s.startWorkers()
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
			log.Println("ZMQ poll failed:", err)
			continue
		}

		if len(polled) > 0 {
			err = s.handleSockets(polled)
			if err != nil {
				log.Println("handleSockets error:", err)
			}
		}
	}

	return nil
}

func (s *Server) Stop() error {
	log.Println("Stopping the server")

	if !s.running {
		return errors.New("Server is already stopped")
	}

	s.running = false
	return nil
}

func (s *Server) startWorkers() error {
	if s.running {
		return errors.New("Server is already running")
	}

	for i := 1; i <= s.workers; i++ {
		worker, err := NewWorker(strconv.Itoa(i), s.workersEndpoint)
		if err != nil {
			return err
		}

		go worker.Start()
	}

	return nil
}

func (s *Server) handleSockets(polled []zmq.Polled) error {
	for _, p := range polled {
		switch socket := p.Socket; socket {
		case s.remoteSocket:
			err := s.handleRemoteSocket()
			if err != nil {
				return err
			}
		case s.workersSocket:
			err := s.handleWorkersSocket()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Server) handleRemoteSocket() error {
	message, err := s.remoteSocket.RecvMessage(0)
	if err != nil {
		return err
	}

	uri, request := message[0], message[len(message)-1]

	// Should not be changed by server
	envelope := message[1 : len(message)-1]

	log.Println("Received request:", request)

	if request == "PING" {
		s.remoteSocket.SendMessage(uri, "PONG")
	} else {
		// TODO Receive message and route it to one of available workers
		s.remoteSocket.SendMessage(uri, envelope, "Not Implemented")
	}

	return nil
}

func (s *Server) handleWorkersSocket() error {
	// TODO Receive message and route it back to the remote client
	return nil
}

func (s *Server) Id() string {
	return s.id
}

func (s *Server) Address() string {
	return s.address
}

func (s *Server) Port() int {
	return s.port
}
