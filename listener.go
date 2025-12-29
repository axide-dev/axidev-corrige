package main

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	spellchecker "github.com/f1monkey/spellchecker/v3"
	typrio "github.com/ziedyousfi/typr-io-go"
)

// CurrentWord tracks the currently typed word and its metrics
type CurrentWord struct {
	Word      string
	Text      *canvas.Text
	StartTime time.Time
	Checker   *spellchecker.Spellchecker
	Sender    *typrio.Sender
}

// NewCurrentWord creates a new CurrentWord instance
func NewCurrentWord(text *canvas.Text, checker *spellchecker.Spellchecker, sender *typrio.Sender) *CurrentWord {
	return &CurrentWord{
		Text:    text,
		Checker: checker,
		Sender:  sender,
	}
}

// Callback handles keyboard events for the typing application
func (w *CurrentWord) Callback(event typrio.KeyEvent) {
	if !event.IsPress() {
		return
	}

	r := event.Rune()
	if r == ' ' {
		// Word completed - calculate speed and check spelling
		if w.Word != "" {
			// Check spelling
			wordLower := strings.ToLower(w.Word)
			isCorrect := w.Checker.IsCorrect(wordLower)

			fmt.Printf("\n=== Word: %s ===\n", w.Word)
			fmt.Printf("Spelling: ")
			if isCorrect {
				fmt.Println("✓ CORRECT")
			} else {
				fmt.Println("✗ INCORRECT")
				// Get suggestions (with max 3 results)
				result := w.Checker.Suggest(wordLower, 3)
				if len(result.Suggestions) > 0 {
					words := make([]string, len(result.Suggestions))
					for i, match := range result.Suggestions {
						words[i] = match.Value
					}
					fmt.Printf("Suggestions: %v\n", words)

					// Auto-correct: select word, delete, and type correction
					if w.Sender != nil {
						correction := result.Suggestions[0].Value
						fmt.Printf("Auto-correcting to: %s\n", correction)
						w.correctWord(correction)
					}
				}
			}
			fmt.Println()

			w.Word = ""
			w.StartTime = time.Time{} // Reset start time
		}
	} else if r != 0 {
		// Start timing on first character
		if w.Word == "" {
			w.StartTime = time.Now()
		}
		w.Word += string(r)
	}

	if w.Text != nil {
		fyne.Do(func() {
			if w.Word == "" {
				w.Text.Text = "Waiting..."
			} else {
				w.Text.Text = w.Word
			}
			w.Text.Refresh()
		})
	}
}

// correctWord selects the previous word, deletes it, and types the correction
func (w *CurrentWord) correctWord(correction string) {
	if w.Sender == nil {
		return
	}

	// Small delay to ensure space key is processed
	time.Sleep(50 * time.Millisecond)

	// On macOS: Option+Shift+Left arrow to select the word (including the space we just typed)
	// We need to select word + space, so we do it twice or use a different approach
	// First, go back over the space
	leftKey := typrio.StringToKey("Left")
	backspaceKey := typrio.StringToKey("Backspace")

	// Select the word: Option+Shift+Left (on macOS, Alt is Option)
	if err := w.Sender.Combo(typrio.ModAlt|typrio.ModShift, leftKey); err != nil {
		fmt.Printf("Error selecting word: %v\n", err)
		return
	}

	time.Sleep(20 * time.Millisecond)

	// Delete the selected word
	if err := w.Sender.Tap(backspaceKey); err != nil {
		fmt.Printf("Error deleting word: %v\n", err)
		return
	}

	time.Sleep(20 * time.Millisecond)

	// Type the correction followed by a space
	if err := w.Sender.TypeText(correction + " "); err != nil {
		fmt.Printf("Error typing correction: %v\n", err)
		return
	}

	w.Sender.Flush()
}
