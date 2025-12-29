package main

import (
	"fmt"
	"log"

	typrio "github.com/ziedyousfi/typr-io-go"
)

func main() {
	fmt.Println("Listening for keyboard events... (Press Space to clear word)")

	// Initialize spellchecker
	sc, err := NewFrenchSpellchecker()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Loaded %d French words into dictionary\n", len(frenchWords))

	// Create sender for auto-correction
	sender, err := typrio.NewSender()
	if err != nil {
		log.Fatal("Failed to create sender:", err)
	}
	defer sender.Close()

	// Request accessibility permissions if needed (macOS)
	caps := sender.Capabilities()
	if caps.NeedsAccessibilityPerm {
		fmt.Println("Requesting accessibility permissions...")
		if !sender.RequestPermissions() {
			fmt.Println("Warning: Permissions not granted, auto-correction may not work")
		}
	}

	// Create window and get text display
	window, text := CreateTypingWindow()

	// Create current word tracker with sender for auto-correction
	cw := NewCurrentWord(text, sc, sender)

	// Start keyboard listener
	listener, err := typrio.NewListener()
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	go func() {
		err = listener.Start(cw.Callback)
		if err != nil {
			log.Printf("Listener error: %v", err)
		}
	}()

	window.ShowAndRun()
}