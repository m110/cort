// Broker is a local broker responsible for routing messages to and from remote services.
// There is one broker started per remote service.
package broker

import (
	zmq "github.com/pebbe/zmq4"
	"log"

	"errors"
	"github.com/m110/cort/resources/consul"
)

type Broker struct {
	service string
	running bool

	remoteSocket *zmq.Socket
	localSocket  *zmq.Socket

	nextNode chan string
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
		service:  service,
		nextNode: make(chan string),
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

	discovery := NewDiscovery(b.service, b.nextNode, nodesManager)
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

func (b *Broker) handleRemoteSocket() error {
	msg, err := b.remoteSocket.RecvMessage(0)
	if err != nil {
		return err
	}

	_, err = b.localSocket.SendMessage(msg)

	return err
}

func (b *Broker) handleLocalSocket() error {
	var err error
	msg, err := b.localSocket.RecvMessage(0)
	if err != nil {
		return err
	}

	node := <-b.nextNode
	if node == "" {
		err = errors.New("No nodes available")
		b.sendError(b.localSocket, msg, err)
		return err
	}

	log.Printf("Routing message to %s (%s): %s\n", b.service, node, msg[1:])
	_, err = b.remoteSocket.SendMessage(node, msg)

	return err
}

func (b *Broker) sendError(socket *zmq.Socket, msg []string, err error) error {
	errMsg := append(msg[0:len(msg)-1], err.Error())
	_, err = socket.SendMessage(errMsg)
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
