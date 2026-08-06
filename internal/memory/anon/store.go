// Package anon is a memory.Store decorator that pseudonymizes text before it reaches the
// wrapped layer — "mémoriser du pseudonymisé" (ROADMAP EPIC-13 transverse). Unlike
// Kern-Orch's courtage-extraction pipeline (which masks, lets a model reason, then
// demasks the same text within one run), memory is long-lived and queried across many
// future calls: there is no round trip here, no token map to keep — the masked form is
// what stays at rest, permanently. Query passes through unchanged; results are already
// safe because Write already made them safe.
package anon

import (
	"context"
	"fmt"

	"github.com/YoLaub/PresidioGo/analyzer"
	"github.com/YoLaub/PresidioGo/anonymizer"
	"github.com/YoLaub/PresidioGo/registry"

	"github.com/yoann/kern-memory/internal/memory"
)

// piiTokenLabels mirrors Kern-Orch's internal/cmd/courtage_anon.go — same recognizer set
// (kern-anon's fr/generic registry), same short prefixes. Kept independent (not shared
// code) because these are two separate Go modules; duplicating ~15 lines is cheaper than a
// cross-repo internal dependency for this.
var piiTokenLabels = map[string]string{
	"FR_NIR":           "NIR",
	"FR_SIREN":         "SIREN",
	"FR_SIRET":         "SIRET",
	"FR_LICENSE_PLATE": "PLAQUE",
	"FR_PHONE_NUMBER":  "TELEPHONE",
	"EMAIL_ADDRESS":    "EMAIL",
	"IBAN_CODE":        "IBAN",
	"CREDIT_CARD":      "CARTE",
	"CRYPTO":           "CRYPTO",
	"MAC_ADDRESS":      "MAC",
	"IP_ADDRESS":       "IP",
	"URL":              "URL",
}

// Store wraps an inner memory.Store, masking PII in Memory.Text on Write.
type Store struct {
	inner memory.Store
}

// Wrap returns a Store that pseudonymizes before delegating to inner.
func Wrap(inner memory.Store) *Store {
	return &Store{inner: inner}
}

// Write masks m.Text before delegating; every other field is untouched.
func (s *Store) Write(ctx context.Context, m memory.Memory) (memory.Memory, error) {
	masked, err := mask(ctx, m.Text)
	if err != nil {
		return memory.Memory{}, fmt.Errorf("anon: mask: %w", err)
	}
	m.Text = masked
	out, err := s.inner.Write(ctx, m)
	if err != nil {
		return memory.Memory{}, fmt.Errorf("anon: write: %w", err)
	}
	return out, nil
}

// Query passes through unchanged — recalled text is already safe, masked at write time.
func (s *Store) Query(ctx context.Context, q memory.Query) ([]memory.Recall, error) {
	out, err := s.inner.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("anon: query: %w", err)
	}
	return out, nil
}

func mask(ctx context.Context, text string) (string, error) {
	eng, err := analyzer.New(analyzer.WithRegistry(registry.Default("fr")))
	if err != nil {
		return "", fmt.Errorf("analyzer: %w", err)
	}
	results, err := eng.Analyze(ctx, text, analyzer.Language("fr"))
	if err != nil {
		return "", fmt.Errorf("analyze: %w", err)
	}

	counters := make(map[string]int)
	mk := func(label string) anonymizer.Operator {
		return anonymizer.Custom("mask_token", func(v string) (string, error) {
			counters[label]++
			return fmt.Sprintf("<%s_%d>", label, counters[label]), nil
		})
	}
	ops := make(map[string]anonymizer.Operator, len(piiTokenLabels)+1)
	for entityType, label := range piiTokenLabels {
		ops[entityType] = mk(label)
	}
	ops["DEFAULT"] = mk("PII")

	result, err := anonymizer.New().Anonymize(text, results, ops)
	if err != nil {
		return "", fmt.Errorf("anonymize: %w", err)
	}
	return result.Text, nil
}
