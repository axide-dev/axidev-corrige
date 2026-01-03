package input

import (
	"fmt"
	"runtime"
	"time"

	"github.com/ziedyousfi/axidev-io-go/keyboard"
)

// Handler processes keyboard input events
type Handler struct {
	listener *keyboard.Listener
	sender   *keyboard.Sender
	callback func(event keyboard.KeyEvent)
}

// Config holds input handler configuration
type Config struct {
	OnEvent func(event keyboard.KeyEvent)
}

// NewHandler creates a new input handler
func NewHandler(cfg Config) (*Handler, error) {
	listener, err := keyboard.NewListener()
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	sender, err := keyboard.NewSender()
	if err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to create sender: %w", err)
	}

	return &Handler{
		listener: listener,
		sender:   sender,
		callback: cfg.OnEvent,
	}, nil
}

// Start begins listening for keyboard events
func (h *Handler) Start() error {
	if h.callback == nil {
		return fmt.Errorf("no callback registered")
	}
	return h.listener.Start(h.callback)
}

// Close cleans up resources
func (h *Handler) Close() {
	if h.listener != nil {
		h.listener.Close()
	}
	if h.sender != nil {
		h.sender.Close()
	}
}

// RequestPermissions requests accessibility permissions (macOS)
func (h *Handler) RequestPermissions() bool {
	if h.sender == nil {
		return false
	}

	caps := h.sender.Capabilities()
	if caps.NeedsAccessibilityPerm {
		fmt.Println("Requesting accessibility permissions...")
		return h.sender.RequestPermissions()
	}
	return true
}

// NeedsPermissions returns true if permissions are required
func (h *Handler) NeedsPermissions() bool {
	if h.sender == nil {
		return false
	}
	return h.sender.Capabilities().NeedsAccessibilityPerm
}

// CanSend returns true if the sender is available
func (h *Handler) CanSend() bool {
	return h.sender != nil
}

// ReplaceWord selects and replaces the current word with a correction
func (h *Handler) ReplaceWord(correction string) error {
	if h.sender == nil {
		return fmt.Errorf("sender not available")
	}

	leftKey := keyboard.StringToKey("Left")
	backspaceKey := keyboard.StringToKey("Backspace")

	// Select the word using platform-specific modifier
	if runtime.GOOS == "darwin" {
		// Alt + Shift + Left Arrow for macOS
		if err := h.sender.Combo(keyboard.ModAlt|keyboard.ModShift, leftKey); err != nil {
			return fmt.Errorf("error selecting word: %w", err)
		}
	} else {
		// Ctrl + Shift + Left Arrow for Windows and Linux
		if err := h.sender.Combo(keyboard.ModCtrl|keyboard.ModShift, leftKey); err != nil {
			return fmt.Errorf("error selecting word: %w", err)
		}
	}

	// Delete the selected word
	if err := h.sender.Tap(backspaceKey); err != nil {
		return fmt.Errorf("error deleting word: %w", err)
	}

	// Type the correction with trailing space
	if err := h.sender.TypeText(correction + " "); err != nil {
		return fmt.Errorf("error typing correction: %w", err)
	}

	h.sender.Flush()
	return nil
}

// TypeText types the given text
func (h *Handler) TypeText(text string) error {
	if h.sender == nil {
		return fmt.Errorf("sender not available")
	}
	return h.sender.TypeText(text)
}

// Flush flushes any pending key events
func (h *Handler) Flush() {
	if h.sender != nil {
		h.sender.Flush()
	}
}

// IsWordSeparator returns true if the rune is a word separator
func IsWordSeparator(r rune) bool {
	return r == ' ' || r == '\n' || r == '\t' || r == '\r'
}

// IsPrintable returns true if the rune is a printable character
func IsPrintable(r rune) bool {
	return r != 0
}

// CorrectionDelay returns the recommended delay after a correction
func CorrectionDelay() time.Duration {
	return 200 * time.Millisecond
}
