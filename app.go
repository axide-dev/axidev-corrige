package main

import (
	"context"
	"fmt"
	"log"

	spellchecker "github.com/f1monkey/spellchecker/v3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/ziedyousfi/axidev-io-go/keyboard"
)

// DisplayUpdate holds text and state for UI updates
type DisplayUpdate struct {
	Text  string
	State string
}

// App struct holds the application state
type App struct {
	ctx         context.Context
	Checker     *spellchecker.Spellchecker
	CurrentWord *CurrentWord
	listener    *keyboard.Listener
	sender      *keyboard.Sender
	updateChan  chan DisplayUpdate
}

// NewApp creates a new App instance
func NewApp(checker *spellchecker.Spellchecker) *App {
	return &App{
		Checker:    checker,
		updateChan: make(chan DisplayUpdate, 100),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Start UI update processor goroutine
	go a.processUpdates()

	// Create sender for auto-correction
	sender, err := keyboard.NewSender()
	if err != nil {
		log.Printf("Failed to create sender: %v", err)
	} else {
		a.sender = sender

		// Request accessibility permissions if needed (macOS)
		caps := sender.Capabilities()
		if caps.NeedsAccessibilityPerm {
			fmt.Println("Requesting accessibility permissions...")
			if !sender.RequestPermissions() {
				fmt.Println("Warning: Permissions not granted, auto-correction may not work")
			}
		}
	}

	// Create current word tracker
	a.CurrentWord = NewCurrentWord(a, a.Checker, a.sender)

	// Initialize keyboard listener
	listener, err := keyboard.NewListener()
	if err != nil {
		log.Printf("Failed to create listener: %v", err)
		return
	}
	a.listener = listener

	// Start keyboard listener in goroutine
	go func() {
		err := a.listener.Start(a.CurrentWord.Callback)
		if err != nil {
			log.Printf("Listener error: %v", err)
		}
	}()
}

// processUpdates handles UI updates in a safe goroutine
func (a *App) processUpdates() {
	for update := range a.updateChan {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "updateText", map[string]string{
				"text":  update.Text,
				"state": update.State,
			})
		}
	}
}

// shutdown is called when the app closes
func (a *App) shutdown(ctx context.Context) {
	close(a.updateChan)
	if a.listener != nil {
		a.listener.Close()
	}
	if a.sender != nil {
		a.sender.Close()
	}
}

// UpdateDisplay sends display text to the frontend (thread-safe)
func (a *App) UpdateDisplay(text string, state string) {
	select {
	case a.updateChan <- DisplayUpdate{Text: text, State: state}:
	default:
		// Channel full, drop update
	}
}
