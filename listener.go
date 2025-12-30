package main

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	spellchecker "github.com/f1monkey/spellchecker/v3"
	"github.com/ziedyousfi/axidev-io-go/keyboard"
)

const wordTimeout = 5 * time.Second

type CurrentWord struct {
	Word          string
	App           *App
	StartTime     time.Time
	LastEventTime time.Time
	Checker       *spellchecker.Spellchecker
	Sender        *keyboard.Sender
	IsCorrecting  bool
	mu            sync.RWMutex
}

func NewCurrentWord(app *App, checker *spellchecker.Spellchecker, sender *keyboard.Sender) *CurrentWord {
	return &CurrentWord{
		App:     app,
		Checker: checker,
		Sender:  sender,
	}
}

func (w *CurrentWord) Callback(event keyboard.KeyEvent) {
	w.mu.Lock()
	if w.IsCorrecting {
		w.mu.Unlock()
		return
	}

	if !event.IsPress() {
		w.mu.Unlock()
		return
	}

	// Check for timeout
	if !w.LastEventTime.IsZero() && time.Since(w.LastEventTime) > wordTimeout {
		fmt.Printf("Timeout reached (%.2fs), clearing word: %s\n", time.Since(w.LastEventTime).Seconds(), w.Word)
		w.Word = ""
		w.StartTime = time.Time{}
	}
	w.LastEventTime = time.Now()

	fmt.Printf("Callback received rune: %q\n", event.Rune())
	fmt.Printf("Current word before processing: %s\n", w.Word)

	r := event.Rune()
	if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
		if w.Word != "" {
			wordLower := strings.ToLower(w.Word)
			isCorrect := w.Checker.IsCorrect(wordLower)

			fmt.Printf("\n=== Word: %s ===\n", w.Word)
			fmt.Printf("Spelling: ")
			if isCorrect {
				fmt.Println("✓ CORRECT")
			} else {
				fmt.Println("✗ INCORRECT")
				result := w.Checker.Suggest(wordLower, 3)
				if len(result.Suggestions) > 0 {
					words := make([]string, len(result.Suggestions))
					for i, match := range result.Suggestions {
						words[i] = match.Value
					}
					fmt.Printf("Suggestions: %v\n", words)

					if w.Sender != nil {
						correction := result.Suggestions[0].Value
						fmt.Printf("Auto-correcting to: %s\n", correction)
						w.mu.Unlock()
						w.correctWord(correction)
						w.mu.Lock()
						w.Word = ""
						w.StartTime = time.Time{}
						w.mu.Unlock()
						w.updateDisplay()
						return
					}
				}
			}
			fmt.Println()

			w.Word = ""
			w.StartTime = time.Time{}
		}
	} else if r != 0 {
		if w.Word == "" {
			w.StartTime = time.Now()
		}
		w.Word += string(r)
	}
	w.mu.Unlock()

	w.updateDisplay()
}

func (w *CurrentWord) updateDisplay() {
	if w.App == nil {
		return
	}

	w.mu.RLock()
	word := w.Word
	isCorrecting := w.IsCorrecting
	w.mu.RUnlock()

	fmt.Printf("Updating display for word: %s (correcting: %v)\n", word, isCorrecting)

	var displayText string
	var state string

	if isCorrecting {
		displayText = "Correcting..."
		state = "correcting"
	} else if word == "" {
		displayText = "Waiting..."
		state = "waiting"
	} else {
		wordLower := strings.ToLower(word)
		if w.Checker.IsCorrect(wordLower) {
			displayText = word + " ✓"
			state = "correct"
		} else {
			result := w.Checker.Suggest(wordLower, 1)
			if len(result.Suggestions) > 0 {
				displayText = fmt.Sprintf("%s → %s", word, result.Suggestions[0].Value)
				state = "suggestion"
			} else {
				displayText = word + " ?"
				state = "incorrect"
			}
		}
	}

	w.App.UpdateDisplay(displayText, state)
}

func (w *CurrentWord) correctWord(correction string) {
	if w.Sender == nil {
		return
	}

	fmt.Println("--- Starting auto-correction ---")
	w.mu.Lock()
	w.IsCorrecting = true
	w.mu.Unlock()
	w.updateDisplay()

	defer func() {
		// Give a small buffer for events to settle
		time.AfterFunc(200*time.Millisecond, func() {
			w.mu.Lock()
			w.IsCorrecting = false
			w.mu.Unlock()
			fmt.Println("--- Auto-correction finished, input unlocked ---")
			w.updateDisplay()
		})
	}()

	leftKey := keyboard.StringToKey("Left")
	backspaceKey := keyboard.StringToKey("Backspace")

	// Select the word

	// Alt + Shift + Left Arrow for macOS
	if runtime.GOOS == "darwin" {
		if err := w.Sender.Combo(keyboard.ModAlt|keyboard.ModShift, leftKey); err != nil {
			fmt.Printf("Error selecting word: %v\n", err)
			return
		}
	} else {
		// Ctrl + Shift + Left Arrow for Windows and Linux
		if err := w.Sender.Combo(keyboard.ModCtrl|keyboard.ModShift, leftKey); err != nil {
			fmt.Printf("Error selecting word: %v\n", err)
			return
		}
	}

	if err := w.Sender.Tap(backspaceKey); err != nil {
		fmt.Printf("Error deleting word: %v\n", err)
		return
	}

	if err := w.Sender.TypeText(correction + " "); err != nil {
		fmt.Printf("Error typing correction: %v\n", err)
		return
	}

	w.Sender.Flush()
}
