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
	if config.Adapter == nil {
		t.Fatal("Expected non-nil PocketBaseAdapter")
	}
}

func TestPocketBaseAdapter(t *testing.T) {
	config := NewPocketBaseConfig()
	adapter := config.Adapter

	if adapter.GetAuthTokenSecret() == "" {
		t.Error("Expected non-empty auth token secret")
	}

	// Note: We can't fully test FindAuthRecordByToken without a running PocketBase instance
	// But we can at least check that the method exists and doesn't panic
	_, err := adapter.FindAuthRecordByToken("test-token", "test-secret")
	if err == nil {
		t.Error("Expected an error for invalid token, got nil")
	}
}
