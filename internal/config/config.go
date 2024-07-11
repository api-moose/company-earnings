package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type PocketBaseConfig struct {
	App *pocketbase.PocketBase
}

func NewPocketBaseConfig() *PocketBaseConfig {
	dataDir := filepath.Join(".", "pb_data")
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	app := pocketbase.New()

	// Customize your PocketBase configuration here
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// add your custom routes or other configurations
		return nil
	})

	return &PocketBaseConfig{
		App: app,
	}
}

func (c *PocketBaseConfig) Start() error {
	if err := c.App.Start(); err != nil {
		return err
	}
	return nil
}
