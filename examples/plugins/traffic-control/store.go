package main

import (
	"sync"
)

type State int

const (
	Created State = iota
	Running
	Stopped
	Destroyed
)

type Container struct {
	State State
	PID   int
}

type Store struct {
	lock       sync.Mutex
	containers map[string]Container
}

func NewStore() *Store {
	return &Store{
		containers: map[string]Container{},
	}
}

func (s *Store) Container(containerID string) (Container, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	container, found := s.containers[containerID]
	return container, found
}

func (s *Store) SetContainer(containerID string, container Container) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.containers[containerID] = container
}

func (s *Store) DeleteContainer(containerID string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.containers, containerID)
}

func (s *Store) ForEach(callback func(ID string, c Container)) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for containerID, container := range s.containers {
		callback(containerID, container)
	}
}
