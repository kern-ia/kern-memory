package config

import "testing"

func TestLoadDefaultsGraphDBWhenUnset(t *testing.T) {
	t.Setenv(EnvGraphDB, "")

	cfg := Load()

	if cfg.GraphDB != "kern-memory-graph.db" {
		t.Errorf("GraphDB = %q, want default %q", cfg.GraphDB, "kern-memory-graph.db")
	}
}

func TestLoadReadsGraphDBOverrideFromEnv(t *testing.T) {
	t.Setenv(EnvGraphDB, "/tmp/custom-graph.db")

	cfg := Load()

	if cfg.GraphDB != "/tmp/custom-graph.db" {
		t.Errorf("GraphDB = %q, want override %q", cfg.GraphDB, "/tmp/custom-graph.db")
	}
}
