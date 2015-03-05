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
	id := fmt.Sprintf("%s:%d", address, port)

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
			s.handleSockets(polled)
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

func (s *Server) Id() string {
	return s.id
}

func (s *Server) Address() string {
	return s.address
}

func (s *Server) Port() int {
	return s.port
}
