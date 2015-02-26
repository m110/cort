package discovery

import (
	zmq "github.com/pebbe/zmq4"
	"log"
)

var discovering = map[string]bool{}

func Start(service string) {
	if !discovering[service] {
		log.Printf("Starting %s discovery", service)
		discovering[service] = true
		go startDiscovery(service)
	}
}

func startDiscovery(service string) {
	var err error

	remoteSocket, err := zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		log.Println(err)
		stopDiscovery(service)
		return
	}

	localSocket, err := zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		log.Println(err)
		stopDiscovery(service)
		return
	}

	poller := zmq.NewPoller()
	poller.Add(remoteSocket, zmq.POLLIN)
	poller.Add(localSocket, zmq.POLLIN)

	for {
		polled, err := poller.Poll(100)
		if err != nil {
			// TODO Log a warning
		}

		if len(polled) > 0 {
			// TODO Handle messages
		}
	}
}

func stopDiscovery(service string) {
	log.Printf("Stopping %s discovery", service)
	delete(discovering, service)

	// TODO Stop discovery goroutine peacefully
}
