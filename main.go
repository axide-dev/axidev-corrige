package main

import (
	"fmt"
	"image/color"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"

	spellchecker "github.com/f1monkey/spellchecker/v3"
	typrio "github.com/ziedyousfi/typr-io-go"
)

// French words list for dictionary
var frenchWords = []string{
	"bonjour", "merci", "salut", "bonsoir", "oui", "non", "peut", "être", "avoir", "faire",
	"aller", "venir", "dire", "pouvoir", "voir", "savoir", "vouloir", "temps", "année", "jour",
	"homme", "femme", "enfant", "chose", "vie", "monde", "pays", "maison", "ville", "rue",
	"monsieur", "madame", "grand", "petit", "bon", "mauvais", "beau", "nouveau", "vieux", "jeune",
	"français", "français", "anglais", "espagnol", "allemand", "italien", "européen", "américain",
	"travail", "école", "livre", "ordinateur", "téléphone", "voiture", "table", "chaise", "porte", "fenêtre",
	"manger", "boire", "dormir", "parler", "écrire", "lire", "écouter", "regarder", "aimer", "détester",
	"heureux", "triste", "content", "fâché", "surpris", "fatigué", "malade", "sain", "fort", "faible",
	"aujourd", "demain", "hier", "maintenant", "toujours", "jamais", "souvent", "parfois", "rarement",
	"ici", "là", "partout", "nulle", "part", "quelque", "part", "ailleurs", "dedans", "dehors",
	"comment", "pourquoi", "quand", "combien", "lequel", "laquelle", "lesquels", "lesquelles",
	"avec", "sans", "pour", "contre", "chez", "vers", "dans", "sur", "sous", "devant",
	"derrière", "entre", "parmi", "durant", "pendant", "avant", "après", "depuis", "jusqu",
	"café", "thé", "eau", "vin", "pain", "fromage", "viande", "poisson", "légume", "fruit",
	"pomme", "orange", "banane", "tomate", "carotte", "salade", "poulet", "boeuf", "porc",
	"rouge", "bleu", "vert", "jaune", "noir", "blanc", "gris", "rose", "violet", "orange",
	"un", "deux", "trois", "quatre", "cinq", "six", "sept", "huit", "neuf", "dix",
	"vingt", "trente", "quarante", "cinquante", "soixante", "cent", "mille", "million",
	"père", "mère", "fils", "fille", "frère", "soeur", "oncle", "tante", "cousin", "cousine",
	"ami", "amie", "copain", "copine", "voisin", "voisine", "collègue", "patron", "client",
	"janvier", "février", "mars", "avril", "mai", "juin", "juillet", "août", "septembre", "octobre",
	"novembre", "décembre", "lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi", "dimanche",
	"matin", "midi", "après", "soir", "nuit", "heure", "minute", "seconde", "semaine", "mois",
}

type CurrentWord struct {
	Word      string
	Text      *canvas.Text
	StartTime time.Time
	Checker   *spellchecker.Spellchecker
}

func main() {
	fmt.Println("Listening for keyboard events... (Press Space to clear word)")

	// Initialize spellchecker with French alphabet
	sc, err := spellchecker.New(
		"abcdefghijklmnopqrstuvwxyzàâäæçéèêëïîôùûüÿœ", // French alphabet including accented characters
	)
	if err != nil {
		log.Fatal(err)
	}

	// Add French words to dictionary
	sc.AddMany(frenchWords)
	fmt.Printf("Loaded %d French words into dictionary\n", len(frenchWords))

	myApp := app.New()

	var window fyne.Window
	if drv, ok := myApp.Driver().(desktop.Driver); ok {
		window = drv.CreateSplashWindow()
	} else {
		window = myApp.NewWindow("Prototypage")
	}

	window.SetFixedSize(true)
	window.SetPadded(false)

	text := canvas.NewText("Waiting...", color.White)
	text.TextSize = 14
	content := container.NewCenter(text)

	window.SetContent(content)
	window.Resize(fyne.NewSize(400, 100))
	window.CenterOnScreen()

	cw := &CurrentWord{
		Text:    text,
		Checker: sc,
	}

	listener, err := typrio.NewListener()
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	go func() {
		err = listener.Start(cw.cb)
		if err != nil {
			log.Printf("Listener error: %v", err)
		}
	}()

	window.ShowAndRun()
}

func (w *CurrentWord) cb(event typrio.KeyEvent) {
	if !event.IsPress() {
		return
	}

	r := event.Rune()
	if r == ' ' {
		// Word completed - calculate speed and check spelling
		if w.Word != "" {
			duration := time.Since(w.StartTime)
			chars := len(w.Word)

			// Calculate metrics
			seconds := duration.Seconds()
			charsPerSecond := float64(chars) / seconds
			// WPM calculation: assuming average word length of 5 characters
			wpm := (charsPerSecond * 60) / 5

			// Check spelling
			wordLower := strings.ToLower(w.Word)
			isCorrect := w.Checker.IsCorrect(wordLower)

			fmt.Printf("\n=== Word: %s ===\n", w.Word)
			fmt.Printf("Duration: %.2f seconds\n", seconds)
			fmt.Printf("Speed: %.2f chars/sec, %.2f WPM\n", charsPerSecond, wpm)
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