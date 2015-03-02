package discovery

import (
	zmq "github.com/pebbe/zmq4"
	"log"

	"github.com/m110/cort/resources/consul"
)

type Discovery struct {
	service string
	running bool

	remoteSocket *zmq.Socket
	localSocket  *zmq.Socket

	nextNode chan Node
}

var discovering = map[string]*Discovery{}

func Start(service string) error {
	_, ok := discovering[service]
	if !ok {
		log.Printf("Starting %s discovery", service)

		d := newDiscovery(service)
		discovering[service] = d

		err := d.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func newDiscovery(service string) *Discovery {
	discovery := &Discovery{
		service: service,
	}

	return discovery
}

func (d *Discovery) Start() error {
	var err error

	d.remoteSocket, err = zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		return err
	}

	d.localSocket, err = zmq.NewSocket(zmq.ROUTER)
	if err == nil {
		err = d.localSocket.Bind("inproc://" + d.service)
	}
	if err != nil {
		return err
	}

	d.nextNode = make(chan Node)

	nodesManager, err := consul.NewConsulProxy()
	if err != nil {
		return err
	}

	watcher := NewWatcher(d.service, d.nextNode, nodesManager)
	watcher.Start()

	d.running = true
	go d.serve()

	return nil
}

func (d *Discovery) serve() {
	poller := zmq.NewPoller()
	poller.Add(d.remoteSocket, zmq.POLLIN)
	poller.Add(d.localSocket, zmq.POLLIN)

	for d.running {
		polled, err := poller.Poll(100)
		if err != nil {
			log.Println("ZMQ poll failed:", err)
			continue
		}

		if len(polled) > 0 {
			for _, p := range polled {
				switch socket := p.Socket; socket {
				case d.remoteSocket:
					err := d.handleRemoteSocket()
					if err != nil {
						log.Println("Remote socket error:", err)
					}
				case d.localSocket:
					err := d.handleLocalSocket()
					if err != nil {
						log.Println("Local socket error:", err)
					}
				}
			}
		}
	}

	d.cleanUp()
}

func (d *Discovery) handleRemoteSocket() error {
	msg, err := d.remoteSocket.RecvMessage(0)
	if err != nil {
		return err
	}

	_, err = d.localSocket.SendMessage(msg)

	return err
}

func (d *Discovery) handleLocalSocket() error {
	msg, err := d.localSocket.RecvMessage(0)
	if err != nil {
		return err
	}

	node := <-d.nextNode
	if node.Err != nil {
		d.sendError(d.localSocket, msg, node.Err)
		return node.Err
	}

	log.Printf("Routing message to %s (%s): %s\n", d.service, node.Uri, msg[1:])
	_, err = d.remoteSocket.SendMessage(msg)

	return err
}

func (d *Discovery) sendError(socket *zmq.Socket, msg []string, err error) error {
	errMsg := append(msg[0:len(msg)-1], err.Error())
	_, err = socket.SendMessage(errMsg)
	return err
}

func (d *Discovery) cleanUp() {
	d.remoteSocket.Close()
	d.localSocket.Close()
}

func (d *Discovery) Stop() {
	log.Printf("Stopping %s discovery", d.service)
	delete(discovering, d.service)

	// TODO Stop serve goroutine peacefully
	d.running = false
}
