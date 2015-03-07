// Discovery is responsible for watching registered service nodes.
// It is tightly connected with Broker and is started along with it (i.e. per one remote service).
package broker

import (
	"fmt"
	"log"
)

type Discovery struct {
	service    string
	nodes      map[string]*Node
	nodesCycle []string

	nodesManager NodesManager

	// Channels for communication with Broker
	nodeCommand  chan<- NodeMessage
	nodeResponse <-chan NodeMessage
	nextNode     chan<- string

	newNodes chan []string
}

type Node struct {
	Uri        string
	Alive      bool
	Registered bool
}

type NodesManager interface {
	ServiceNodes(service string) ([]string, error)
}

func NewDiscovery(service string, nodesManager NodesManager, nodeCommand, nodeResponse chan NodeMessage, nextNode chan string) *Discovery {
	return &Discovery{
		service:      service,
		nodes:        make(map[string]*Node),
		nodesManager: nodesManager,
		nodeCommand:  nodeCommand,
		nodeResponse: nodeResponse,
		nextNode:     nextNode,
		newNodes:     make(chan []string),
	}
}

func (d *Discovery) Start() {
	go d.watchNodes()
	go d.run()
}

func (d *Discovery) run() {
	nextNode := ""

	for {
		select {
		case newNodes := <-d.newNodes:
			log.Println("Received nodes update:", newNodes)

			err := d.processNewNodes(newNodes)
			if err != nil {
				log.Println("Error while new nodes processing:", err)
			}

			nextNode = d.getNextNode()
		default:
			select {
			case d.nextNode <- nextNode:
				nextNode = d.getNextNode()
			case message := <-d.nodeResponse:
				err := d.handleNodeResponse(message)
				if err != nil {
					log.Println("Node response error:", err)
				}
			default:
			}
		}
	}
}

func (d *Discovery) processNewNodes(newNodes []string) error {
	for _, uri := range newNodes {
		node, ok := d.nodes[uri]

		if ok {
			if !node.Registered {
				log.Println("Setting node", uri, "as registered")
				node.Registered = true
				d.nodeCommand <- NodeMessage{CONNECT, uri}
			}
		} else {
			newNode := &Node{
				Uri:        uri,
				Registered: true,
				Alive:      false,
			}

			log.Println("Discovered new node:", uri)
			d.nodes[uri] = newNode
			d.nodeCommand <- NodeMessage{CONNECT, uri}
		}
	}

	for _, uri := range d.removedNodes(newNodes) {
		log.Println("Deregistering node", uri)
		d.nodes[uri].Registered = false
		d.nodes[uri].Alive = false
		d.nodeCommand <- NodeMessage{DISCONNECT, uri}
	}

	return nil
}

func (d *Discovery) handleNodeResponse(message NodeMessage) error {
	log.Println("Received response from broker:", message)
	switch message.Message {
	case PONG:
		_, ok := d.nodes[message.Uri]
		if ok {
			log.Println("Setting node", message.Uri, "as alive")
			d.nodes[message.Uri].Alive = true
		}
	default:
		return fmt.Errorf("Unknown node message: %d", message.Message)
	}

	return nil
}

func (d *Discovery) watchNodes() {
	for {
		nodes, err := d.nodesManager.ServiceNodes(d.service)
		if err != nil {
			log.Println("Nodes manager error:", err.Error())
			continue
		}

		d.newNodes <- nodes
	}
}

func (d *Discovery) getNextNode() string {
	if len(d.nodesCycle) == 0 {
		return ""
	} else {
		node := d.nodesCycle[0]

		// Cycle the nodes
		if len(d.nodesCycle) > 1 {
			d.nodesCycle = append(d.nodesCycle[1:], node)
		}

		return node
	}
}

func (d *Discovery) removedNodes(uris []string) []string {
	var set map[string]struct{} = make(map[string]struct{})
	for _, uri := range uris {
		set[uri] = struct{}{}
	}

	var removed []string
	for key, _ := range d.nodes {
		_, ok := set[key]
		if !ok {
			removed = append(removed, key)
		}
	}

	return removed
}
