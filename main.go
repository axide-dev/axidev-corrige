package main

import (
	"embed"
	"fmt"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	axidevio "github.com/ziedyousfi/axidev-io-go"
)

//go:embed frontend/*
var assets embed.FS

func main() {
	axidevio.SetLogLevel(axidevio.LogLevelWarn)
	fmt.Println("Listening for keyboard events... (Press Space to clear word)")

	// Initialize spellchecker
	sc, err := NewFrenchSpellchecker()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Loaded %d French words into dictionary\n", len(frenchWords))

	// Create app instance
	app := NewApp(sc)

	// Create Wails application
	err = wails.Run(&options.App{
		Title:  "Axidev Corrige",
		Width:  400,
		Height: 100,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal("Error:", err)
	}
}
