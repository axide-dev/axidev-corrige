package main

import (
	"fmt"
	"log"
	"os"

	"gioui.org/app"
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

	// Create overlay window
	overlay := NewOverlayWindow()

	// Create current word tracker
	cw := NewCurrentWord(overlay, sc, sender)

	// Initialize keyboard listener
	listener, err := typrio.NewListener()
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	// Start keyboard listener in goroutine
	go func() {
		err = listener.Start(cw.Callback)
		if err != nil {
			log.Printf("Listener error: %v", err)
		}
	}()

	// Run overlay in goroutine (Gio requirement)
	go func() {
		if err := overlay.Run(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	// Gio main loop (must be on main thread)
	app.Main()
}