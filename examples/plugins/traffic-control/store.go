package main

import (
	"sync"
)

// State describes the rough state of the container
type State int

const (
	// Created means that the container was just created, but it
	// is not yet running.
	Created State = iota
	// Running means that the container is running, so we can get
	// the PID of the process.
	Running
	// Stopped means that the container was stopped.
	Stopped
	// Destroyed means that the container was destroyed.
	Destroyed
)

// Container holds some data needed for traffic control.
type Container struct {
	State State
	PID   int
}

// Store holds all the registered containers.
type Store struct {
	lock       sync.Mutex
	containers map[string]Container
}

// NewStore creates a new store, duh.
func NewStore() *Store {
	return &Store{
		containers: map[string]Container{},
	}
}

// Container gets the information about a container with a given ID.
func (s *Store) Container(containerID string) (Container, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	container, found := s.containers[containerID]
	return container, found
}

// SetContainer overwrites the information about the container with a
// given ID.
func (s *Store) SetContainer(containerID string, container Container) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.containers[containerID] = container
}

// DeleteContainer removes the container with a given ID from the
// store.
func (s *Store) DeleteContainer(containerID string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.containers, containerID)
}

// ForEach calls a given callback function on each stored
// container. Do not call other Store functions in the callback, a
// deadlock will ensue.
func (s *Store) ForEach(callback func(ID string, c Container)) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for containerID, container := range s.containers {
		callback(containerID, container)
	}
}
