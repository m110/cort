package broker

import (
	"io/ioutil"
	"log"
	"reflect"
	"testing"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

type StubNodesManager struct {
	returned bool
}

func TestNextNode(t *testing.T) {
	expected := []string{"a", "b", "c", "a", "b", "c", "a"}

	discovery := NewDiscovery("AnyService", nil, nil, nil, nil)
	discovery.nodesCycle = []string{"a", "b", "c"}

	for i, want := range expected {
		got := discovery.getNextNode()
		if got != want {
			t.Errorf("Wanted %s at %d, got %s", want, i, got)
		}
	}
}

func TestUpdateNodesCycle(t *testing.T) {
	discovery := NewDiscovery("AnyService", nil, nil, nil, nil)

	discovery.nodes["key_1"] = &Node{Uri: "key_1", Alive: true}
	discovery.nodes["key_2"] = &Node{Uri: "key_2", Alive: false}
	discovery.nodes["key_3"] = &Node{Uri: "key_3", Alive: true}

	expected := []string{"key_1", "key_3"}
	discovery.updateNodesCycle()

	if !reflect.DeepEqual(discovery.nodesCycle, expected) {
		t.Errorf("nodesCycle contains %s instead of %s", discovery.nodesCycle, expected)
	}
}

func TestRemovedNodes(t *testing.T) {
	discovery := NewDiscovery("AnyService", nil, nil, nil, nil)

	discovery.nodes["key_1"] = &Node{}
	discovery.nodes["key_2"] = &Node{}

	removed := discovery.removedNodes([]string{"key_1", "key_3"})
	expected := []string{"key_2"}

	if !reflect.DeepEqual(removed, expected) {
		t.Errorf("Removed nodes returned %s instead of %s", removed, expected)
	}
}
