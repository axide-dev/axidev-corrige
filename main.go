package main

import (
	"embed"
	"fmt"
	"log"

	"github.com/ziedyousfi/axidev-corrige/internal/app"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	axidevio "github.com/ziedyousfi/axidev-io-go"
)

//go:embed frontend/*
var assets embed.FS

func main() {
	axidevio.SetLogLevel(axidevio.LogLevelWarn)
	fmt.Println("Listening for keyboard events...")

	// Create app instance with default config
	application, err := app.New(app.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	// Create Wails application
	err = wails.Run(&options.App{
		Title:  "Axidev Corrige",
		Width:  400,
		Height: 100,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  application.Startup,
		OnShutdown: application.Shutdown,
		Bind: []interface{}{
			application,
		},
	})

	if err != nil {
		log.Fatal("Error:", err)
	}
}
