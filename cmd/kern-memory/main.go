// Command kern-memory is the daemon and its seed CLI, kern-ui's only source for C8's
// document storage slice.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/yoann/kern-memory/internal/config"
	"github.com/yoann/kern-memory/internal/httpapi"
	"github.com/yoann/kern-memory/internal/memory"
	"github.com/yoann/kern-memory/internal/memory/anon"
	"github.com/yoann/kern-memory/internal/memory/okf"
	"github.com/yoann/kern-memory/internal/memory/vector"
	"github.com/yoann/kern-memory/internal/store"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "serve":
		err = runServe()
	case "seed":
		err = runSeed(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "kern-memory:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: kern-memory serve | kern-memory seed <title> <body-file>")
}

func runServe() error {
	cfg := config.Load()
	if err := config.CheckExposure(cfg.Addr, cfg.Token); err != nil {
		return err
	}

	s, err := store.Open(cfg.DB)
	if err != nil {
		return err
	}
	defer s.Close()

	mem, closeMem, err := openMemory(cfg)
	if err != nil {
		return err
	}
	defer closeMem()

	slog.Info("kern-memory: listening", "addr", cfg.Addr, "pseudonymize", cfg.Pseudonymize)
	return http.ListenAndServe(cfg.Addr, httpapi.NewCombinedRouter(s, mem, cfg.Token))
}

// openMemory wires EPIC-13 phase 1's Router (internal/memory) from the two layers plus
// the optional kern-anon transverse — see internal/memory/anon's doc for why Write-time
// masking has no round trip here, unlike courtage-extraction's.
func openMemory(cfg config.Config) (memory.Store, func(), error) {
	okfStore, err := okf.Open(cfg.OKFDB)
	if err != nil {
		return nil, nil, err
	}
	vectorStore, err := vector.Open(cfg.VectorDB, cfg.OllamaModel, cfg.OllamaURL)
	if err != nil {
		okfStore.Close()
		return nil, nil, err
	}

	router := &memory.Router{OKF: okfStore, Vector: vectorStore}
	closeFn := func() {
		okfStore.Close()
		vectorStore.Close()
	}

	if !cfg.Pseudonymize {
		return router, closeFn, nil
	}
	return anon.Wrap(router), closeFn, nil
}

func runSeed(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: kern-memory seed <title> <body-file>")
	}
	title, bodyPath := args[0], args[1]

	body, err := os.ReadFile(bodyPath)
	if err != nil {
		return fmt.Errorf("read body file: %w", err)
	}

	cfg := config.Load()
	s, err := store.Open(cfg.DB)
	if err != nil {
		return err
	}
	defer s.Close()

	id := newID()
	if err := s.Seed(context.Background(), id, title, string(body)); err != nil {
		return err
	}
	fmt.Println(id)
	return nil
}

func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
