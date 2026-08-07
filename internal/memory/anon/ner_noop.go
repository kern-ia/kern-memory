//go:build !onnx

package anon

import "github.com/YoLaub/PresidioGo/nlp"

// nlpEngine is nil in the default build (no -tags onnx) — see ner_onnx.go for the real
// implementation.
func nlpEngine() nlp.Engine { return nil }
