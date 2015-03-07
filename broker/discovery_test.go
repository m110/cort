package broker

import (
	"errors"
	"io/ioutil"
	"log"
	"testing"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

type StubNodesManager struct {
	returned bool
}

func (s *StubNodesManager) ServiceNodes(service string) ([]string, error) {
	if !s.returned {
		s.returned = true
		return []string{"a", "b", "c"}, nil
	} else {
		return nil, errors.New("Timed out")
	}
}

func TestDiscovery(t *testing.T) {
	nextNode := make(chan string)

	expected := []string{"a", "b", "c", "a", "b", "c", "a"}

	discovery := NewDiscovery("AnyService", &StubNodesManager{}, nil, nil, nextNode)
	discovery.Start()

	for i, want := range expected {
		got := <-nextNode
		if got != want {
			t.Errorf("Wanted %s at %d, got %s", want, i, got)
		}
	}
}
