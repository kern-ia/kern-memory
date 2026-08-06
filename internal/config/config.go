// Package config reads kern-memory's environment variables.
package config

import "os"

const (
	EnvAddr  = "KERN_MEMORY_ADDR"
	EnvToken = "KERN_MEMORY_TOKEN"
	EnvDB    = "KERN_MEMORY_DB"

	// EPIC-13 phase 1 (internal/memory/*) — a separate DB file per layer, not a shared
	// one: the .okf layer is plain SQL rows, the vector layer is chromem-go's own
	// gob-encoded persistence format, neither reads the other's file.
	EnvOKFDB        = "KERN_MEMORY_OKF_DB"
	EnvVectorDB     = "KERN_MEMORY_VECTOR_DB"
	EnvOllamaModel  = "KERN_MEMORY_OLLAMA_MODEL"
	EnvOllamaURL    = "KERN_MEMORY_OLLAMA_URL"
	EnvPseudonymize = "KERN_MEMORY_PSEUDONYMIZE"
)

// Config is what `serve` needs, read once at startup.
type Config struct {
	Addr  string
	Token string
	DB    string

	OKFDB        string
	VectorDB     string
	OllamaModel  string
	OllamaURL    string
	Pseudonymize bool
}

// Load reads Config from the environment, applying the same local-development defaults as
// kern-orch: a loopback address and database files beside the binary.
func Load() Config {
	return Config{
		Addr:  getOr(EnvAddr, "127.0.0.1:7080"),
		Token: os.Getenv(EnvToken),
		DB:    getOr(EnvDB, "kern-memory.db"),

		OKFDB:       getOr(EnvOKFDB, "kern-memory-okf.db"),
		VectorDB:    getOr(EnvVectorDB, "kern-memory-vector.db"),
		OllamaModel: getOr(EnvOllamaModel, "nomic-embed-text"),
		OllamaURL:   os.Getenv(EnvOllamaURL), // empty = chromem-go's own default
		// Pseudonymize defaults to true — "mémoriser du pseudonymisé" is the safe
		// default (ROADMAP EPIC-13 transverse), not an opt-in a caller has to remember.
		Pseudonymize: os.Getenv(EnvPseudonymize) != "false",
	}
}

func getOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
