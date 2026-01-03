package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/axide-dev/axidev-corrige/internal/checker"
	"github.com/axide-dev/axidev-corrige/internal/display"
	"github.com/axide-dev/axidev-corrige/internal/input"
	"github.com/axide-dev/axidev-corrige/internal/state"
	"github.com/axide-dev/axidev-corrige/internal/writing"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/ziedyousfi/axidev-io-go/keyboard"
)

// Config holds application configuration
type Config struct {
	WordTimeout time.Duration
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		WordTimeout: 5 * time.Second,
	}
}

// App is the main application orchestrator
type App struct {
	ctx     context.Context
	config  Config
	state   *state.Machine
	writing *writing.Writing
	checker *checker.Checker
	input   *input.Handler
	display *display.Manager
}

// New creates a new App instance
func New(cfg Config) (*App, error) {
	// Initialize the spell checker
	chk, err := checker.NewFrenchChecker()
	if err != nil {
		return nil, fmt.Errorf("failed to create checker: %w", err)
	}

	fmt.Printf("Loaded %d French words into dictionary\n", chk.WordCount())

	app := &App{
		config:  cfg,
		state:   state.NewMachine(),
		writing: writing.NewWriting(writing.Config{Timeout: cfg.WordTimeout}),
		checker: chk,
		display: display.NewManager(),
	}

	// Register state transition handler
	app.state.OnTransition(app.onStateTransition)

	return app, nil
}

// Startup is called when the Wails app starts
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Set window always on top
	runtime.WindowSetAlwaysOnTop(ctx, true)

	// Start display manager
	a.display.Start(ctx)

	// Initialize input handler
	handler, err := input.NewHandler(input.Config{
		OnEvent: a.handleKeyEvent,
	})
	if err != nil {
		log.Printf("Failed to create input handler: %v", err)
		return
	}
	a.input = handler

	// Request permissions if needed
	if handler.NeedsPermissions() {
		if !handler.RequestPermissions() {
			fmt.Println("Warning: Permissions not granted, auto-correction may not work")
		}
	}

	// Start keyboard listener
	go func() {
		if err := a.input.Start(); err != nil {
			log.Printf("Input handler error: %v", err)
		}
	}()

	// Transition to idle state
	a.state.Transition(state.Idle)
	a.updateDisplay()
}

// Shutdown is called when the app closes
func (a *App) Shutdown(ctx context.Context) {
	a.display.Stop()
	if a.input != nil {
		a.input.Close()
	}
}

// handleKeyEvent processes keyboard events
func (a *App) handleKeyEvent(event keyboard.KeyEvent) {
	// Only process key presses
	if !event.IsPress() {
		return
	}

	// Check if we can accept input
	if !a.state.CanAcceptInput() {
		return
	}

	// Check for timeout
	if a.writing.CheckTimeout() {
		fmt.Println("Timeout reached, cleared writing buffer")
		a.state.Transition(state.Idle)
	}

	r := event.Rune()

	// Handle word separators
	if input.IsWordSeparator(r) {
		a.handleWordComplete()
		return
	}

	// Handle printable characters
	if input.IsPrintable(r) {
		a.handleCharacter(r)
	}
}

// handleCharacter processes a single character
func (a *App) handleCharacter(r rune) {
	// Transition to listening if idle
	if a.state.Is(state.Idle) {
		a.state.Transition(state.Listening)
	}

	a.writing.AddChar(r)
	a.updateDisplay()

	fmt.Printf("Added char '%c', current word: %s\n", r, a.writing.GetCurrentWord().Text)
}

// handleWordComplete processes word completion
func (a *App) handleWordComplete() {
	word := a.writing.CompleteWord()
	if word == nil {
		return
	}

	fmt.Printf("\n=== Word completed: %s ===\n", word.Text)

	// Check spelling
	result := a.checker.Check(word.Text, 3)

	if result.IsCorrect {
		fmt.Println("Spelling: ✓ CORRECT")
	} else {
		fmt.Println("Spelling: ✗ INCORRECT")

		if len(result.Suggestions) > 0 {
			words := make([]string, len(result.Suggestions))
			score := make([]float64, len(result.Suggestions))
			for i, s := range result.Suggestions {
				words[i] = s.Value
				score[i] = s.Score
			}
			fmt.Printf("Suggestions: %v\n", words)
			fmt.Printf("Scores: %v\n", score)

			// Check score of best suggestion
			if result.Suggestions[0].Score < 0.8 {
				fmt.Println("Best suggestion score too low, skipping auto-correction")
				fmt.Println()
				a.updateDisplay()
				return
			}

			// Perform auto-correction
			if a.input != nil && a.input.CanSend() {
				correction := result.Suggestions[0].Value
				a.performCorrection(word.Text, correction)
			}
		}
	}
	fmt.Println()

	// Update display
	a.updateDisplay()

	// Transition back to idle if writing buffer is empty
	if a.writing.IsEmpty() {
		a.state.Transition(state.Idle)
	}
}

// performCorrection corrects a misspelled word
func (a *App) performCorrection(original, correction string) {
	fmt.Printf("Auto-correcting '%s' to '%s'\n", original, correction)

	// Transition to correcting state
	a.state.Transition(state.Correcting)
	a.display.Correcting()

	// Perform the correction
	if err := a.input.ReplaceWord(correction); err != nil {
		fmt.Printf("Correction failed: %v\n", err)
	}

	// Update the word in writing buffer
	a.writing.ReplaceLastWord(correction)

	// Delay before transitioning back
	time.AfterFunc(input.CorrectionDelay(), func() {
		if a.writing.IsEmpty() {
			a.state.Transition(state.Idle)
		} else {
			a.state.Transition(state.Listening)
		}
		a.updateDisplay()
		fmt.Println("--- Auto-correction finished ---")
	})
}

// updateDisplay updates the UI based on current state
func (a *App) updateDisplay() {
	var text string
	var displayState string

	switch a.state.Current() {
	case state.Correcting:
		text = "Correcting..."
		displayState = display.StateCorrecting

	case state.Idle:
		text = "Waiting..."
		displayState = display.StateWaiting

	case state.Listening:
		word := a.writing.GetCurrentWord()
		if word.IsEmpty() {
			// Show last completed word if any
			if last := a.writing.GetLastWord(); last != nil {
				text = last.Text + " ✓"
				displayState = display.StateCorrect
			} else {
				text = "Listening..."
				displayState = display.StateListening
			}
		} else {
			result := a.checker.Check(word.Text, 1)
			if result.IsCorrect {
				text = word.Text + " ✓"
				displayState = display.StateCorrect
			} else if len(result.Suggestions) > 0 {
				text = fmt.Sprintf("%s → %s", word.Text, result.Suggestions[0].Value)
				displayState = display.StateSuggestion
			} else {
				text = word.Text + " ?"
				displayState = display.StateIncorrect
			}
		}

	case state.Paused:
		text = "Paused"
		displayState = display.StateWaiting
	}

	a.display.Send(text, displayState)
}

// onStateTransition handles state change events
func (a *App) onStateTransition(from, to state.State) {
	fmt.Printf("State: %s → %s\n", from, to)
}

// GetState returns the current application state (for UI binding)
func (a *App) GetState() string {
	return a.state.Current().String()
}

// GetWriting returns the current writing text (for UI binding)
func (a *App) GetWriting() string {
	return a.writing.GetFullText()
}

// GetWordCount returns the number of words in the buffer (for UI binding)
func (a *App) GetWordCount() int {
	return a.writing.WordCount()
}
