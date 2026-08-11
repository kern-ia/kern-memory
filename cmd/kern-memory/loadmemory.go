package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/yoann/kern-memory/internal/config"
	"github.com/yoann/kern-memory/internal/memory"
)

// loadFile is the editable content format (epic-2 issue 1): a batch of memories plus the
// graph edges between them, meant to be hand-edited and re-loaded with `load-memory` —
// mirroring seed's directness (cmd/kern-memory/main.go's runSeed), not a new architecture.
type loadFile struct {
	Memories []memoryRecord `json:"memories"`
	Edges    []edgeRecord   `json:"edges"`
}

// memoryRecord is one `memories` entry. Kind is memory.Kind's own JSON values ("okf" or
// "vector") — an empty Kind is valid and, per Router.Write, defaults to the vector layer.
type memoryRecord struct {
	ID       string            `json:"id"`
	Kind     memory.Kind       `json:"kind"`
	Text     string            `json:"text"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
}

// edgeRecord is one `edges` entry, matching the graph layer's FromKind/FromID/ToKind/ToID/
// Relation fields exactly (docs/planning/specs/05-graph-interface-contract.md). ID is
// optional — the same memory.Memory.ID field graph.Store already honors when set (see
// graph.Store.Write's TestWriteKeepsAGivenID) — but note the file format's own example
// omits it, which is what makes re-running the loader on an unchanged file accumulate
// duplicate edges rather than upsert (internal/memory/graph/store_test.go's
// TestRepeatedWriteWithNoIDCreatesASeparateEdgeNotAnUpsert documents this gap).
type edgeRecord struct {
	ID       string `json:"id"`
	FromKind string `json:"from_kind"`
	FromID   string `json:"from_id"`
	ToKind   string `json:"to_kind"`
	ToID     string `json:"to_id"`
	Relation string `json:"relation"`
}

// parseLoadFile parses data as a loadFile. It only validates that the JSON itself is
// well-formed — per-field validation (e.g. graph's Relation length cap) is each store's own
// job, reused as-is rather than duplicated here.
func parseLoadFile(data []byte) (loadFile, error) {
	var f loadFile
	if err := json.Unmarshal(data, &f); err != nil {
		return loadFile{}, fmt.Errorf("load-memory: parse: %w", err)
	}
	return f, nil
}

// writeAll writes every memory then every edge in f through router, in file order. A memory
// entry is written with its own Kind (Router.Write routes it); an edge entry is written with
// Kind: memory.KindGraph so it lands in the graph layer, per the issue's scope.
func (f loadFile) writeAll(ctx context.Context, router memory.Store) error {
	for _, m := range f.Memories {
		mm := memory.Memory{ID: m.ID, Kind: m.Kind, Text: m.Text, Tags: m.Tags, Metadata: m.Metadata}
		if _, err := router.Write(ctx, mm); err != nil {
			return fmt.Errorf("load-memory: write memory %q: %w", m.ID, err)
		}
	}
	for _, e := range f.Edges {
		em := memory.Memory{
			ID:       e.ID,
			Kind:     memory.KindGraph,
			FromKind: e.FromKind,
			FromID:   e.FromID,
			ToKind:   e.ToKind,
			ToID:     e.ToID,
			Relation: e.Relation,
		}
		if _, err := router.Write(ctx, em); err != nil {
			return fmt.Errorf("load-memory: write edge %s->%s: %w", e.FromID, e.ToID, err)
		}
	}
	return nil
}

// runLoadMemory implements `kern-memory load-memory <file>`: parse the file, open the same
// Router runServe/openMemory already builds, write everything through it. Direct store
// calls, no HTTP round trip — same category as runSeed, no HTTP write path for bulk content
// either, by design.
func runLoadMemory(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: kern-memory load-memory <file>")
	}

	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("load-memory: read file: %w", err)
	}
	f, err := parseLoadFile(data)
	if err != nil {
		return err
	}

	cfg := config.Load()
	mem, closeMem, err := openMemory(cfg)
	if err != nil {
		return err
	}
	defer closeMem()

	if err := f.writeAll(context.Background(), mem); err != nil {
		return err
	}
	fmt.Printf("loaded %d memories and %d edges\n", len(f.Memories), len(f.Edges))
	return nil
}
