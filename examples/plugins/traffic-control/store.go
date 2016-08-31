package main

import (
	"sync"
)

// State is the container internal state
type State int

const (
	// Created state
	Created State = iota
	// Running state
	Running
	// Stopped state
	Stopped
	// Destroyed state
	Destroyed
)

// Container data structure
type Container struct {
	State State
	PID   int
}

// Store data structure
type Store struct {
	lock       sync.Mutex
	containers map[string]Container
}

// NewStore instantiates a new Store
func NewStore() *Store {
	return &Store{
		containers: map[string]Container{},
	}
}

// Container returns a container form its ID
func (s *Store) Container(containerID string) (Container, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	container, found := s.containers[containerID]
	return container, found
}

// SetContainer sets a container into the store
func (s *Store) SetContainer(containerID string, container Container) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.containers[containerID] = container
}

// DeleteContainer deletes a container from the store
func (s *Store) DeleteContainer(containerID string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.containers, containerID)
}

// ForEach execute a function on each container in the store
func (s *Store) ForEach(callback func(ID string, c Container)) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for containerID, container := range s.containers {
		callback(containerID, container)
	}
}
