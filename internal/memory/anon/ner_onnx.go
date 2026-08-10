//go:build onnx

package anon

import (
	"fmt"
	"os"
	"sync"

	"github.com/YoLaub/PresidioGo/nlp"
	"github.com/YoLaub/PresidioGo/nlp/onnx"
)

var (
	nlpEngineOnce sync.Once
	nlpEngineVal  nlp.Engine
)

// nlpEngine returns the shared ONNX NER engine, loaded once and reused across every Write
// — mirrors Kern-Orch's internal/cmd/courtage_ner_onnx.go (same reasoning: reloading a
// ~170MB model per call would make writes prohibitively slow). Same "vide = repli sûr"
// pattern as the rest of this project: unset KERN_ANON_NER_MODEL_DIR (or a build without
// -tags onnx at all, see ner_noop.go) means no person-name detection, not a broken write
// path — the regex recognizers (IBAN, email, ...) run regardless.
func nlpEngine() nlp.Engine {
	nlpEngineOnce.Do(func() {
		dir := os.Getenv("KERN_ANON_NER_MODEL_DIR")
		if dir == "" {
			return
		}
		eng := onnx.New(dir)
		if err := eng.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "anon: échec du chargement du moteur NER (%s) : %v\n", dir, err)
			return
		}
		nlpEngineVal = eng
	})
	return nlpEngineVal
}
