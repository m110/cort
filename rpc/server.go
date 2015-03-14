package rpc

import (
	"errors"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"log"
	"strconv"
)

const (
	workerReadyMessage = "READY"
)

type Server struct {
	id      string
	address string
	port    int
	workers int
	running bool

	workersQueue []string

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

// handleRemoteSocket receives message from remote service and routes it to first worker available.
// Frames received from remote service:
//
//     | envelope | request |
//
// Frames routed to worker:
//
//     | worker_id | (empty) | envelope | request |
func (s *Server) handleRemoteSocket() error {
	message, err := s.remoteSocket.RecvMessage(0)
	if err != nil {
		return err
	}

	log.Println("Received from remote service:", message)

	envelope := message[:len(message)-1]
	request := message[len(message)-1]

	worker_id := s.nextWorker()

	if worker_id == "" {
		_, err = s.remoteSocket.SendMessage(envelope, "No workers available")
		return err
	}

	_, err = s.workersSocket.SendMessage(worker_id, "", envelope, request)
	return err
}

// handleWorkersSocket receives message from worker and routes it to remote service.
// Frames received from worker:
//
//     | worker_id | (empty) | remote_uri | envelope | response |
//
// Frames routed to remote service:
//
//     | remote_id | envelope | response |

func (s *Server) handleWorkersSocket() error {
	message, err := s.workersSocket.RecvMessage(0)
	if err != nil {
		return err
	}

	log.Println("Received from worker:", message)

	worker_id := message[0]

	empty_frame := message[1]
	if empty_frame != "" {
		return fmt.Errorf("Empty frame contains invalid data: %s", empty_frame)
	}

	s.workersQueue = append(s.workersQueue, worker_id)

	response := message[len(message)-1]
	if response == workerReadyMessage {
		log.Println("Worker", worker_id, "ready")
		return nil
	}

	remote_id := message[2]
	envelope := message[3 : len(message)-1]
	_, err = s.remoteSocket.SendMessage(remote_id, envelope, response)
	return err
}

func (s *Server) nextWorker() string {
	if len(s.workersQueue) == 0 {
		return ""
	}

	next := s.workersQueue[0]
	s.workersQueue = s.workersQueue[1:]

	return next
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
