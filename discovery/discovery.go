// Discovery is responsible for watching registered service nodes.
// It is tightly connected with Broker and is started along with it (i.e. per one remote service).
package discovery

import (
	"log"
)

type Discovery struct {
	service string
	nodes   []Node

	newNodes chan []Node
	nextNode chan string

	nodesManager NodesManager
}

type Node struct {
	Uri        string
	alive      bool
	registered bool
}

type NodesManager interface {
	ServiceNodes(service string) ([]string, error)
}

func NewDiscovery(service string, nextNode chan string, nodesManager NodesManager) *Discovery {
	return &Discovery{
		service:      service,
		newNodes:     make(chan []Node),
		nextNode:     nextNode,
		nodesManager: nodesManager,
	}
}

func (d *Discovery) Start() {
	go d.watchNodes()
	go d.run()
}

func (d *Discovery) run() {
	d.nodes = <-d.newNodes
	nextNode := d.getNextNode()

	for {
		select {
		case newNodes := <-d.newNodes:
			log.Println("Received nodes update:", newNodes)
			d.nodes = newNodes
			nextNode = d.getNextNode()
		default:
			select {
			case d.nextNode <- nextNode:
				nextNode = d.getNextNode()
			default:
			}
		}
	}
}

func (d *Discovery) watchNodes() {
	for {
		nodes, err := d.nodesManager.ServiceNodes(d.service)
		if err != nil {
			log.Println("Nodes manager error:", err.Error())
			continue
		}

		var newNodes []Node
		for _, node := range nodes {
			newNodes = append(newNodes, Node{
				Uri: node,
			})
		}

		d.newNodes <- newNodes
	}
}

func (d *Discovery) getNextNode() string {
	if len(d.nodes) == 0 {
		return ""
	} else {
		node := d.nodes[0]

		// Cycle the nodes
		if len(d.nodes) > 1 {
			d.nodes = append(d.nodes[1:], node)
		}

		return node.Uri
	}
}
