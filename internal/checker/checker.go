package checker

import (
	"embed"
	"strings"

	spellchecker "github.com/f1monkey/spellchecker/v3"
)

//go:embed francais.txt
var frenchDictFS embed.FS

// Checker wraps the spellchecker with additional functionality
type Checker struct {
	sc        *spellchecker.Spellchecker
	wordCount int
}

// Suggestion represents a spelling suggestion
type Suggestion struct {
	Value string
	Score float64
}

// Result holds spell check results
type Result struct {
	Original    string
	IsCorrect   bool
	Suggestions []Suggestion
}

// NewFrenchChecker creates a checker with the French dictionary
func NewFrenchChecker() (*Checker, error) {
	words, err := loadFrenchWords()
	if err != nil {
		return nil, err
	}

	// Initialize spellchecker with French alphabet
	alphabet := "abcdefghijklmnopqrstuvwxyzàâäæçéèêëïîôùûüÿœ"
	sc, err := spellchecker.New(alphabet)
	if err != nil {
		return nil, err
	}

	sc.AddMany(words)

	return &Checker{
		sc:        sc,
		wordCount: len(words),
	}, nil
}

func loadFrenchWords() ([]string, error) {
	data, err := frenchDictFS.ReadFile("francais.txt")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	words := make([]string, 0, len(lines))
	for _, line := range lines {
		word := strings.TrimSpace(line)
		if word != "" {
			words = append(words, word)
		}
	}
	return words, nil
}

// WordCount returns the number of words in the dictionary
func (c *Checker) WordCount() int {
	return c.wordCount
}

// IsCorrect checks if a word is spelled correctly
func (c *Checker) IsCorrect(word string) bool {
	return c.sc.IsCorrect(strings.ToLower(word))
}

// Check performs a full spell check on a word
func (c *Checker) Check(word string, maxSuggestions int) Result {
	wordLower := strings.ToLower(word)
	isCorrect := c.sc.IsCorrect(wordLower)

	result := Result{
		Original:  word,
		IsCorrect: isCorrect,
	}

	if !isCorrect && maxSuggestions > 0 {
		scResult := c.sc.Suggest(wordLower, maxSuggestions)
		result.Suggestions = make([]Suggestion, len(scResult.Suggestions))
		for i, s := range scResult.Suggestions {
			result.Suggestions[i] = Suggestion{
				Value: s.Value,
				Score: s.Score,
			}
		}
	}

	return result
}

// Suggest returns spelling suggestions for a word
func (c *Checker) Suggest(word string, maxSuggestions int) []Suggestion {
	result := c.Check(word, maxSuggestions)
	return result.Suggestions
}

// BestSuggestion returns the best spelling suggestion, or empty string if none
func (c *Checker) BestSuggestion(word string) string {
	suggestions := c.Suggest(word, 1)
	if len(suggestions) > 0 {
		return suggestions[0].Value
	}
	return ""
}
