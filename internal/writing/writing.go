package writing

import (
	"strings"
	"sync"
	"time"
)

// Word represents a single word being typed
type Word struct {
	Text      string
	StartTime time.Time
}

// IsEmpty returns true if the word has no text
func (w Word) IsEmpty() bool {
	return w.Text == ""
}

// Writing represents the current writing session with multiple words
type Writing struct {
	Words         []Word
	CurrentWord   Word
	LastEventTime time.Time
	Timeout       time.Duration
	mu            sync.RWMutex
}

// Config holds configuration for the Writing buffer
type Config struct {
	Timeout time.Duration
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Timeout: 5 * time.Second,
	}
}

// NewWriting creates a new Writing buffer
func NewWriting(cfg Config) *Writing {
	return &Writing{
		Words:   make([]Word, 0),
		Timeout: cfg.Timeout,
	}
}

// AddChar adds a character to the current word
func (w *Writing) AddChar(r rune) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()

	// Check for timeout
	if w.checkTimeoutLocked(now) {
		w.clearLocked()
	}

	w.LastEventTime = now

	// Start new word if current is empty
	if w.CurrentWord.IsEmpty() {
		w.CurrentWord.StartTime = now
	}
	w.CurrentWord.Text += string(r)
}

// CompleteWord marks the current word as complete and adds it to the list
func (w *Writing) CompleteWord() *Word {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.CurrentWord.IsEmpty() {
		return nil
	}

	word := w.CurrentWord
	w.Words = append(w.Words, word)
	w.CurrentWord = Word{}

	return &word
}

// ReplaceLastWord replaces the last completed word with a correction
func (w *Writing) ReplaceLastWord(correction string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.Words) > 0 {
		w.Words[len(w.Words)-1].Text = correction
	}
}

// Clear resets the entire writing buffer
func (w *Writing) Clear() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.clearLocked()
}

func (w *Writing) clearLocked() {
	w.Words = make([]Word, 0)
	w.CurrentWord = Word{}
	w.LastEventTime = time.Time{}
}

// checkTimeoutLocked checks if timeout has occurred (must hold lock)
func (w *Writing) checkTimeoutLocked(now time.Time) bool {
	if w.LastEventTime.IsZero() {
		return false
	}
	return now.Sub(w.LastEventTime) > w.Timeout
}

// CheckTimeout checks if the writing has timed out and clears if so
func (w *Writing) CheckTimeout() bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.checkTimeoutLocked(time.Now()) {
		w.clearLocked()
		return true
	}
	return false
}

// GetCurrentWord returns a copy of the current word being typed
func (w *Writing) GetCurrentWord() Word {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.CurrentWord
}

// GetWords returns a copy of all completed words
func (w *Writing) GetWords() []Word {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make([]Word, len(w.Words))
	copy(result, w.Words)
	return result
}

// GetLastWord returns the last completed word, or nil if none
func (w *Writing) GetLastWord() *Word {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if len(w.Words) == 0 {
		return nil
	}
	word := w.Words[len(w.Words)-1]
	return &word
}

// IsEmpty returns true if there are no words and no current word
func (w *Writing) IsEmpty() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.Words) == 0 && w.CurrentWord.IsEmpty()
}

// GetFullText returns all words joined with spaces, including current word
func (w *Writing) GetFullText() string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	parts := make([]string, 0, len(w.Words)+1)
	for _, word := range w.Words {
		parts = append(parts, word.Text)
	}
	if !w.CurrentWord.IsEmpty() {
		parts = append(parts, w.CurrentWord.Text)
	}
	return strings.Join(parts, " ")
}

// WordCount returns the total number of words (completed + current if any)
func (w *Writing) WordCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	count := len(w.Words)
	if !w.CurrentWord.IsEmpty() {
		count++
	}
	return count
}

// RemoveLastWord removes and returns the last completed word
func (w *Writing) RemoveLastWord() *Word {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.Words) == 0 {
		return nil
	}

	lastIdx := len(w.Words) - 1
	word := w.Words[lastIdx]
	w.Words = w.Words[:lastIdx]
	return &word
}
