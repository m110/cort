package discovery

import (
	zmq "github.com/pebbe/zmq4"
	"log"
)

type Discovery struct {
	service string
	running bool

	remoteSocket *zmq.Socket
	localSocket  *zmq.Socket
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
			// TODO Log a warning
		}

		if len(polled) > 0 {
			// TODO Handle messages
		}
	}

	d.cleanUp()
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
