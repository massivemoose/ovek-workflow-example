package config

import (
	"testing"
	"time"
)

func TestLoadAppConfig(t *testing.T) {
	env := map[string]string{
		"PORT":                  "9090",
		"POCKETBASE_URL":        "http://pocketbase.local/",
		"PB_SUPERUSER_EMAIL":    "admin@example.com",
		"PB_SUPERUSER_PASSWORD": "secret",
	}

	cfg := LoadApp(func(key string) string { return env[key] })

	if cfg.Port != "9090" {
		t.Fatalf("Port = %q, want 9090", cfg.Port)
	}
	if cfg.PocketBaseURL != "http://pocketbase.local" {
		t.Fatalf("PocketBaseURL = %q, want trimmed URL", cfg.PocketBaseURL)
	}
	if cfg.SuperuserEmail != "admin@example.com" {
		t.Fatalf("SuperuserEmail = %q, want admin@example.com", cfg.SuperuserEmail)
	}
	if cfg.SuperuserPass != "secret" {
		t.Fatalf("SuperuserPass = %q, want secret", cfg.SuperuserPass)
	}
	if cfg.RequestTimeout != 10*time.Second {
		t.Fatalf("RequestTimeout = %s, want 10s", cfg.RequestTimeout)
	}
	if !cfg.CollectionEnsure {
		t.Fatal("CollectionEnsure = false, want true")
	}
}

func TestLoadAppConfigDefaults(t *testing.T) {
	cfg := LoadApp(func(string) string { return "" })

	if cfg.Port != "8080" {
		t.Fatalf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.PocketBaseURL != "http://127.0.0.1:8090" {
		t.Fatalf("PocketBaseURL = %q, want local PocketBase URL", cfg.PocketBaseURL)
	}
}

func TestLoadWorkflowConfig(t *testing.T) {
	env := map[string]string{
		"POCKETBASE_URL":        "https://pb.example.test/",
		"PB_SUPERUSER_EMAIL":    "admin@example.com",
		"PB_SUPERUSER_PASSWORD": "secret",
		"OVEK_WORKFLOW_RUN_ID":  "run_1234567890",
	}

	cfg := LoadWorkflow(func(key string) string { return env[key] })

	if cfg.PocketBaseURL != "https://pb.example.test" {
		t.Fatalf("PocketBaseURL = %q, want trimmed URL", cfg.PocketBaseURL)
	}
	if cfg.WorkflowRunID != "run_1234567890" {
		t.Fatalf("WorkflowRunID = %q, want run_1234567890", cfg.WorkflowRunID)
	}
	if cfg.RequestTimeout != 10*time.Second {
		t.Fatalf("RequestTimeout = %s, want 10s", cfg.RequestTimeout)
	}
}
