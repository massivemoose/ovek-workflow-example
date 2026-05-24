package config

import (
	"strings"
	"time"
)

type Config struct {
	Port             string
	PocketBaseURL    string
	SuperuserEmail   string
	SuperuserPass    string
	SuperuserToken   string
	WorkflowRunID    string
	RequestTimeout   time.Duration
	CollectionEnsure bool
}

type LookupFunc func(string) string

func LoadApp(lookup LookupFunc) Config {
	return Config{
		Port:             env(lookup, "PORT", "8080"),
		PocketBaseURL:    strings.TrimRight(env(lookup, "POCKETBASE_URL", "http://127.0.0.1:8090"), "/"),
		SuperuserEmail:   strings.TrimSpace(lookup("PB_SUPERUSER_EMAIL")),
		SuperuserPass:    lookup("PB_SUPERUSER_PASSWORD"),
		SuperuserToken:   strings.TrimSpace(lookup("PB_SUPERUSER_TOKEN")),
		RequestTimeout:   10 * time.Second,
		CollectionEnsure: true,
	}
}

func LoadWorkflow(lookup LookupFunc) Config {
	cfg := LoadApp(lookup)
	cfg.Port = ""
	cfg.WorkflowRunID = strings.TrimSpace(lookup("OVEK_WORKFLOW_RUN_ID"))
	return cfg
}

func env(lookup LookupFunc, name string, fallback string) string {
	value := strings.TrimSpace(lookup(name))
	if value == "" {
		return fallback
	}
	return value
}
