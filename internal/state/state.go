package state

import (
	"sync"
)

// State represents the current application state
type State int

const (
	// Idle - waiting for input, no active writing
	Idle State = iota
	// Listening - actively collecting keyboard input
	Listening
	// Correcting - performing an auto-correction
	Correcting
	// Paused - input collection temporarily paused
	Paused
)

func (s State) String() string {
	switch s {
	case Idle:
		return "idle"
	case Listening:
		return "listening"
	case Correcting:
		return "correcting"
	case Paused:
		return "paused"
	default:
		return "unknown"
	}
}

// Machine manages application state transitions
type Machine struct {
	current   State
	mu        sync.RWMutex
	listeners []func(from, to State)
}

// NewMachine creates a new state machine starting in Idle state
func NewMachine() *Machine {
	return &Machine{
		current:   Idle,
		listeners: make([]func(from, to State), 0),
	}
}

// Current returns the current state
func (m *Machine) Current() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

// Is checks if the current state matches the given state
func (m *Machine) Is(s State) bool {
	return m.Current() == s
}

// Transition changes the state and notifies listeners
func (m *Machine) Transition(to State) {
	m.mu.Lock()
	from := m.current
	if from == to {
		m.mu.Unlock()
		return
	}
	m.current = to
	listeners := make([]func(from, to State), len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.Unlock()

	// Notify listeners outside the lock
	for _, listener := range listeners {
		listener(from, to)
	}
}

// OnTransition registers a callback for state transitions
func (m *Machine) OnTransition(fn func(from, to State)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, fn)
}

// CanCorrect returns true if correction is allowed in current state
func (m *Machine) CanCorrect() bool {
	s := m.Current()
	return s == Listening || s == Idle
}

// CanAcceptInput returns true if input can be accepted
func (m *Machine) CanAcceptInput() bool {
	s := m.Current()
	return s == Idle || s == Listening
}
