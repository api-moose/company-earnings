package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

type AuthProvider interface {
	FindAuthRecordByToken(token, secret string) (*models.Record, error)
	GetAuthTokenSecret() string
}

type PocketBaseConfig struct {
	App     *pocketbase.PocketBase
	Adapter AuthProvider
}

type PocketBaseAdapter struct {
	App *pocketbase.PocketBase
}

func (a *PocketBaseAdapter) FindAuthRecordByToken(token, secret string) (*models.Record, error) {
	// Implement the actual logic to find the auth record by token
	// This is a placeholder implementation, adjust according to your needs
	return a.App.Dao().FindAuthRecordByToken(token, secret)
}

func (a *PocketBaseAdapter) GetAuthTokenSecret() string {
	// Implement the actual logic to get the auth token secret
	// This is a placeholder implementation, adjust according to your needs
	return a.App.Settings().RecordAuthToken.Secret
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

	adapter := &PocketBaseAdapter{App: app}

	return &PocketBaseConfig{
		App:     app,
		Adapter: adapter,
	}
}

func (c *PocketBaseConfig) Start() error {
	if err := c.App.Start(); err != nil {
		return err
	}
	return nil
}
