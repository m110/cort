package discovery

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

func (s *StubNodesManager) NewNodes(service string) ([]string, error) {
	if !s.returned {
		s.returned = true
		return []string{"a", "b", "c"}, nil
	} else {
		return nil, errors.New("Timed out")
	}
}

func TestWatcher(t *testing.T) {
	nextNode := make(chan Node)

	expected := []string{"a", "b", "c", "a", "b", "c", "a"}

	watcher := NewWatcher("AnyService", nextNode, &StubNodesManager{})
	watcher.Start()

	for i, want := range expected {
		got := <-nextNode
		if got.Uri != want {
			t.Errorf("Wanted %s at %d, got %s", want, i, got.Uri)
		}
	}
}
