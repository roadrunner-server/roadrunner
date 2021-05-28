package storage

import (
	"sync"

	"github.com/spiral/roadrunner/v2/pkg/bst"
)

type Storage struct {
	sync.RWMutex
	BST bst.Storage
}

func NewStorage() *Storage {
	return &Storage{
		BST: bst.NewBST(),
	}
}

func (s *Storage) InsertMany(connID string, topics []string) {
	s.Lock()
	defer s.Unlock()

	for i := 0; i < len(topics); i++ {
		s.BST.Insert(connID, topics[i])
	}
}

func (s *Storage) Insert(connID string, topic string) {
	s.Lock()
	defer s.Unlock()

	s.BST.Insert(connID, topic)
}

func (s *Storage) RemoveMany(connID string, topics []string) {
	s.Lock()
	defer s.Unlock()

	for i := 0; i < len(topics); i++ {
		s.BST.Remove(connID, topics[i])
	}
}

func (s *Storage) Remove(connID string, topic string) {
	s.Lock()
	defer s.Unlock()

	s.BST.Remove(connID, topic)
}

// GetByPtrTS Thread safe get
func (s *Storage) GetByPtrTS(topics []string, res map[string]struct{}) {
	s.Lock()
	defer s.Unlock()

	for i := 0; i < len(topics); i++ {
		d := s.BST.Get(topics[i])
		if len(d) > 0 {
			for ii := range d {
				res[ii] = struct{}{}
			}
		}
	}
}

func (s *Storage) GetByPtr(topics []string, res map[string]struct{}) {
	s.RLock()
	defer s.RUnlock()

	for i := 0; i < len(topics); i++ {
		d := s.BST.Get(topics[i])
		if len(d) > 0 {
			for ii := range d {
				res[ii] = struct{}{}
			}
		}
	}
}
