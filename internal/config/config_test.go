package config

import (
	"testing"
)

func TestNewPocketBaseConfig(t *testing.T) {
	config := NewPocketBaseConfig()
	if config == nil {
		t.Fatal("Expected non-nil PocketBaseConfig")
	}
	if config.App == nil {
		t.Fatal("Expected non-nil PocketBase app")
	}
}
