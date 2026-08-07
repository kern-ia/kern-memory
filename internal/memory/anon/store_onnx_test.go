//go:build onnx

package anon

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/yoann/kern-memory/internal/memory"
)

// Real integration test against the ONNX NER engine — skipped unless
// KERN_ANON_NER_MODEL_DIR points at a real downloaded model (see
// Kern-Anon/scripts/download-model-macos.sh).
func requireNerModel(t *testing.T) {
	t.Helper()
	if os.Getenv("KERN_ANON_NER_MODEL_DIR") == "" {
		t.Skip("KERN_ANON_NER_MODEL_DIR not set, skipping real NER integration test")
	}
}

func TestWriteMasksAPersonNameButKeepsOrganizationWhenNerIsConfigured(t *testing.T) {
	requireNerModel(t)

	inner := &fakeStore{}
	s := Wrap(inner)

	_, err := s.Write(context.Background(), memory.Memory{
		Text: "La banque Alpha Crédit, contactée par Jean Dupont, accepte les prêts en SCI.",
	})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	got := inner.written[0].Text
	if strings.Contains(got, "Jean Dupont") {
		t.Errorf("text still contains the raw name: %q", got)
	}
	if !strings.Contains(got, "<PERSONNE_1>") {
		t.Errorf("text does not contain a PERSONNE token: %q", got)
	}
	// The whole point of this memory is the bank's name — masking it would break RAG.
	if !strings.Contains(got, "Alpha Crédit") {
		t.Errorf("text should keep the organization name (it IS the content), got %q", got)
	}
}
