package discovery

import (
	"errors"
	"log"
)

type Watcher struct {
	service string
	nodes   []Node

	newNodes chan []Node
	nextNode chan Node

	nodesManager NodesManager
}

type Node struct {
	Uri string
	Err error
}

var emptyNode = Node{
	Err: errors.New("No nodes available"),
}

type NodesManager interface {
	NewNodes(service string) ([]string, error)
}

func NewWatcher(service string, nextNode chan Node, nodesManager NodesManager) *Watcher {
	return &Watcher{
		service:      service,
		newNodes:     make(chan []Node),
		nextNode:     nextNode,
		nodesManager: nodesManager,
	}
}

func (w *Watcher) Start() {
	go w.watchNodes()
	go w.run()
}

func (w *Watcher) run() {
	nextNode := emptyNode

	for {
		select {
		case newNodes := <-w.newNodes:
			log.Println("Received nodes update:", newNodes)
			w.nodes = newNodes
			nextNode = w.getNextNode()
		default:
			select {
			case w.nextNode <- nextNode:
				nextNode = w.getNextNode()
			default:
			}
		}
	}
}

func (w *Watcher) watchNodes() {
	for {
		nodes, err := w.nodesManager.NewNodes(w.service)
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

		w.newNodes <- newNodes
	}
}

func (w *Watcher) getNextNode() Node {
	if len(w.nodes) == 0 {
		return emptyNode
	} else {
		node := w.nodes[0]

		// Cycle the nodes
		if len(w.nodes) > 1 {
			w.nodes = append(w.nodes[1:], node)
		}

		return node
	}
}
