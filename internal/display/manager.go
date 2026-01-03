package display

import (
	"context"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Update holds text and state for UI updates
type Update struct {
	Text  string
	State string
}

// State constants for display states
const (
	StateWaiting    = "waiting"
	StateListening  = "listening"
	StateCorrect    = "correct"
	StateIncorrect  = "incorrect"
	StateSuggestion = "suggestion"
	StateCorrecting = "correcting"
)

// Manager handles UI display updates
type Manager struct {
	ctx        context.Context
	updateChan chan Update
	mu         sync.RWMutex
	running    bool
}

// NewManager creates a new display manager
func NewManager() *Manager {
	return &Manager{
		updateChan: make(chan Update, 100),
	}
}

// Start begins processing display updates
func (m *Manager) Start(ctx context.Context) {
	m.mu.Lock()
	m.ctx = ctx
	m.running = true
	m.mu.Unlock()

	go m.processUpdates()
}

// Stop stops the display manager
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		m.running = false
		close(m.updateChan)
	}
}

// processUpdates handles UI updates in a goroutine
func (m *Manager) processUpdates() {
	for update := range m.updateChan {
		m.mu.RLock()
		ctx := m.ctx
		m.mu.RUnlock()

		if ctx != nil {
			runtime.EventsEmit(ctx, "updateText", map[string]string{
				"text":  update.Text,
				"state": update.State,
			})
		}
	}
}

// Send sends an update to the display (non-blocking)
func (m *Manager) Send(text, state string) {
	m.mu.RLock()
	running := m.running
	m.mu.RUnlock()

	if !running {
		return
	}

	select {
	case m.updateChan <- Update{Text: text, State: state}:
	default:
		// Channel full, drop update
	}
}

// SendUpdate sends an Update struct to the display
func (m *Manager) SendUpdate(update Update) {
	m.Send(update.Text, update.State)
}

// Waiting sends a waiting state update
func (m *Manager) Waiting() {
	m.Send("Waiting...", StateWaiting)
}

// Correcting sends a correcting state update
func (m *Manager) Correcting() {
	m.Send("Correcting...", StateCorrecting)
}
