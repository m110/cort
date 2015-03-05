package rpc

import (
	zmq "github.com/pebbe/zmq4"
	"log"
)

type Worker struct {
	id       string
	endpoint string
	running  bool

	socket *zmq.Socket
}

func NewWorker(id, endpoint string) (*Worker, error) {
	worker := &Worker{
		id:       id,
		endpoint: endpoint,
	}

	var err error
	worker.socket, err = zmq.NewSocket(zmq.REQ)
	if err != nil {
		return nil, err
	}

	err = worker.socket.Connect(endpoint)
	if err != nil {
		return nil, err
	}

	return worker, nil
}

func (w *Worker) Start() {
	w.running = true

	w.socket.SendMessage("READY")

	log.Printf("Worker-%s ready\n", w.id)

	for w.running {
		message, err := w.socket.RecvMessage(0)
		if err != nil {
			log.Println("ZMQ recv failed:", err)
			continue
		}

		request := message[len(message)-1]
		response, err := w.processRequest(request)
		if err != nil {
			log.Println("Error while processing request:", err)
			// TODO Send proper error message
			w.socket.SendMessage(err)
			continue
		}

		w.socket.SendMessage(response)
	}
}

func (w *Worker) stop() {
	w.running = false
}

func (w *Worker) processRequest(request string) (string, error) {
	// TODO Process the request
	return request, nil
}
