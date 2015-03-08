// Broker is a local broker responsible for routing messages to and from remote services.
// There is one broker started per remote service.
package broker

import (
	"errors"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"log"

	"github.com/m110/cort/resources/consul"
)

type Message int

const (
	CONNECT Message = iota
	DISCONNECT
	PING
	PONG
)

type Broker struct {
	service string
	running bool

	remoteSocket *zmq.Socket
	localSocket  *zmq.Socket

	nodeCommand  chan NodeMessage
	nodeResponse chan NodeMessage
	nextNode     chan string
}

type NodeMessage struct {
	Message Message
	Uri     string
}

var brokers = map[string]*Broker{}

func Start(service string) error {
	_, ok := brokers[service]
	if !ok {
		log.Printf("Starting %s broker", service)

		d := newBroker(service)
		brokers[service] = d

		err := d.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func newBroker(service string) *Broker {
	discovery := &Broker{
		service:      service,
		nodeCommand:  make(chan NodeMessage),
		nodeResponse: make(chan NodeMessage),
		nextNode:     make(chan string),
	}

	return discovery
}

func (b *Broker) Start() error {
	var err error

	b.remoteSocket, err = zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		return err
	}

	b.localSocket, err = zmq.NewSocket(zmq.ROUTER)
	if err == nil {
		err = b.localSocket.Bind("inproc://" + b.service)
	}
	if err != nil {
		return err
	}

	nodesManager, err := consul.NewConsulProxy()
	if err != nil {
		return err
	}

	discovery := NewDiscovery(b.service, nodesManager, b.nodeCommand, b.nodeResponse, b.nextNode)
	discovery.Start()

	b.running = true
	go b.serve()

	return nil
}

func (b *Broker) serve() {
	poller := zmq.NewPoller()
	poller.Add(b.remoteSocket, zmq.POLLIN)
	poller.Add(b.localSocket, zmq.POLLIN)

	for b.running {
		select {
		case message := <-b.nodeCommand:
			err := b.handleNodeCommand(message)
			if err != nil {
				log.Println("Node command error:", err)
			}
		default:
		}

		polled, err := poller.Poll(100)
		if err != nil {
			log.Println("ZMQ poll failed:", err)
			continue
		}

		if len(polled) > 0 {
			for _, p := range polled {
				switch socket := p.Socket; socket {
				case b.remoteSocket:
					err := b.handleRemoteSocket()
					if err != nil {
						log.Println("Remote socket error:", err)
					}
				case b.localSocket:
					err := b.handleLocalSocket()
					if err != nil {
						log.Println("Local socket error:", err)
					}
				}
			}
		}
	}

	b.cleanUp()
}

func (b *Broker) handleNodeCommand(message NodeMessage) error {
	switch message.Message {
	case CONNECT:
		log.Println("Connecting to", message.Uri)
		b.remoteSocket.Connect(message.Uri)
		b.sendRemote(message.Uri, "PING")
	case DISCONNECT:
		log.Println("Disconnecting from", message.Uri)
		b.remoteSocket.Disconnect(message.Uri)
	case PING:
		log.Println("Sending PING to", message.Uri)
		b.sendRemote(message.Uri, "PING")
	default:
		return fmt.Errorf("Unknown node message: %d", message.Message)
	}

	return nil
}

func (b *Broker) handleRemoteSocket() error {
	response, err := b.remoteSocket.RecvMessage(0)
	if err != nil {
		return err
	}

	uri, message := response[0], response[1:]
	request := message[len(message)-1]

	log.Printf("Received message from %s (%s): %s\n", b.service, uri, message)

	if request == "PONG" {
		b.nodeResponse <- NodeMessage{PONG, uri}
	} else {
		err = b.sendLocal(message...)
	}

	return err
}

func (b *Broker) handleLocalSocket() error {
	var err error
	request, err := b.localSocket.RecvMessage(0)
	if err != nil {
		return err
	}

	uri := <-b.nextNode
	if uri == "" {
		err = errors.New("No nodes available")
		b.sendError(b.localSocket, request, err)
		return err
	}

	log.Printf("Routing message to %s (%s): %s\n", b.service, uri, request)
	err = b.sendRemote(uri, request...)

	return err
}

func (b *Broker) sendRemote(uri string, frames ...string) error {
	frames = append([]string{uri}, frames...)
	_, err := b.remoteSocket.SendMessage(frames)
	return err
}

func (b *Broker) sendLocal(frames ...string) error {
	_, err := b.localSocket.SendMessage(frames)
	return err
}

func (b *Broker) sendError(socket *zmq.Socket, msg []string, err error) error {
	msg[len(msg)-1] = err.Error()
	_, err = socket.SendMessage(msg)
	return err
}

func (b *Broker) cleanUp() {
	b.remoteSocket.Close()
	b.localSocket.Close()
}

func (b *Broker) Stop() {
	log.Printf("Stopping %s broker", b.service)
	delete(brokers, b.service)

	// TODO Stop serve goroutine peacefully
	b.running = false
}
