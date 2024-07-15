package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFirebaseApp(t *testing.T) {
	app, err := NewFirebaseApp("path/to/serviceAccountKey.json")
	assert.NoError(t, err)
	assert.NotNil(t, app)
}
