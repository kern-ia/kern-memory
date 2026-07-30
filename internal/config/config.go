// Package config reads kern-memory's environment variables.
package config

import "os"

const (
	EnvAddr  = "KERN_MEMORY_ADDR"
	EnvToken = "KERN_MEMORY_TOKEN"
	EnvDB    = "KERN_MEMORY_DB"
)

// Config is what `serve` needs, read once at startup.
type Config struct {
	Addr  string
	Token string
	DB    string
}

// Load reads Config from the environment, applying the same local-development defaults as
// kern-orch: a loopback address and a database file beside the binary.
func Load() Config {
	return Config{
		Addr:  getOr(EnvAddr, "127.0.0.1:7080"),
		Token: os.Getenv(EnvToken),
		DB:    getOr(EnvDB, "kern-memory.db"),
	}
}

func getOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
