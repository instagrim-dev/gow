// Package localfp builds the identity fingerprints a LOCAL model adapter owes the
// reuse key, and separates the two things that are routinely conflated:
//
//	semantic config fingerprint  !=  execution environment fingerprint
//
// ---------------------------------------------------------------------------
// WHY THIS EXISTS
//
// provider.MinerIdentity.Version() already hashes the WHOLE identity struct,
// ConfigFingerprint included, so changing configuration does produce a different
// reuse key. Nothing is broken there. The unmet obligation is on the ADAPTER: the
// struct doc says "a live adapter folds its full configuration here", and if a
// local adapter omits the quantization then Q4_K_M and Q5_K_M collide on one
// reuse key and a Q5 request can silently return a revision produced by Q4.
//
// There is no live adapter in the tree yet -- every provider is a deterministic
// fixture -- which is exactly why this is being fixed now. miner_version sits
// inside UNIQUE(problem_id, failure_space_id, miner_version, predicate_schema,
// min_support) on append-only, trigger-immutable rows: once live rows exist, any
// later change to what the fingerprint covers silently orphans every prior reuse
// key. The decision is free before the first live row and expensive after it.
//
// ---------------------------------------------------------------------------
// THE SPLIT, AND WHY IT IS NOT ONE OPAQUE HASH
//
// SEMANTIC inputs can change WHAT the model produces. They belong in
// ConfigFingerprint and therefore in the reuse key.
//
// EXECUTION inputs are performance and scheduling knobs (batch size, layer
// split, thread count, Metal buffer settings). They are recorded for provenance
// and replay, but folding them into the reuse key fragments the cache: re-running
// the same model and sampling policy with a different batch size would miss a
// perfectly valid prior revision and pay for a fresh mining pass.
//
// The dividing question is what the contract promises. If it promised
// byte-identical stochastic output, batch size WOULD be semantic (batching can
// change floating-point reduction order and thus sampled tokens). This package
// promises the weaker, more useful contract -- same model, same sampling policy
// -- and says so, rather than implying a determinism it cannot deliver.
package localfp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// SemanticConfig is every input that can change WHAT a local model produces. It
// is hashed into provider.MinerIdentity.ConfigFingerprint and thus into the
// durable miner_version reuse key.
//
// WeightsSHA256 is the load-bearing field: it identifies the exact artifact and
// so distinguishes quantizations even when a filename or model tag does not.
// Quantization is ALSO recorded explicitly, because a human reading a fingerprint
// mismatch should not have to resolve a hash to see that Q4 and Q5 differ.
type SemanticConfig struct {
	// --- artifact identity ---
	WeightsSHA256 string `json:"weights_sha256"` // exact artifact hash
	Quantization  string `json:"quantization"`   // e.g. Q4_K_M, Q8_0, MXFP4
	// --- runtime identity (a kernel change can change outputs) ---
	RuntimeName    string `json:"runtime_name"`    // llama.cpp, vllm-metal, ollama
	RuntimeVersion string `json:"runtime_version"` // commit or release identity
	// --- prompt / decoding contract ---
	PromptTemplateVersion string `json:"prompt_template_version"`
	ChatTemplateSHA256    string `json:"chat_template_sha256,omitempty"`
	GrammarRevision       string `json:"grammar_revision,omitempty"` // GBNF / JSON-schema revision
	// --- sampling policy ---
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
	TopK        int     `json:"top_k"`
	MinP        float64 `json:"min_p"`
	Seed        int64   `json:"seed"`
	ContextSize int     `json:"context_size"` // n_ctx: truncation changes what the model sees
}

// ExecutionEnvironment is scheduling and performance configuration. It is
// recorded for provenance and replay but deliberately EXCLUDED from the reuse
// key (see the package doc on cache fragmentation).
type ExecutionEnvironment struct {
	BatchSize      int               `json:"batch_size,omitempty"`  // n_batch
	UBatchSize     int               `json:"ubatch_size,omitempty"` // n_ubatch
	GPULayers      int               `json:"gpu_layers,omitempty"`
	Threads        int               `json:"threads,omitempty"`
	TensorSplit    string            `json:"tensor_split,omitempty"`
	FlashAttention bool              `json:"flash_attention,omitempty"`
	MMap           bool              `json:"mmap,omitempty"`
	HostOS         string            `json:"host_os,omitempty"`
	HostModel      string            `json:"host_model,omitempty"`
	Extra          map[string]string `json:"extra,omitempty"`
}

// requiredSemantic names fields with no safe default. A silently-empty artifact
// hash is the exact failure this package exists to prevent, so an incomplete
// fingerprint is an ERROR rather than a hash of blanks: two different adapters
// that both forget the hash would otherwise agree on one reuse key.
func (s SemanticConfig) validate() error {
	var missing []string
	if strings.TrimSpace(s.WeightsSHA256) == "" {
		missing = append(missing, "weights_sha256")
	}
	if strings.TrimSpace(s.Quantization) == "" {
		missing = append(missing, "quantization")
	}
	if strings.TrimSpace(s.RuntimeName) == "" {
		missing = append(missing, "runtime_name")
	}
	if strings.TrimSpace(s.RuntimeVersion) == "" {
		missing = append(missing, "runtime_version")
	}
	if strings.TrimSpace(s.PromptTemplateVersion) == "" {
		missing = append(missing, "prompt_template_version")
	}
	if s.ContextSize <= 0 {
		missing = append(missing, "context_size")
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("localfp: incomplete semantic fingerprint, missing %s: "+
			"an under-specified fingerprint lets two different configurations share one reuse key",
			strings.Join(missing, ", "))
	}
	return nil
}

// Fingerprint returns the deterministic hash to place in
// provider.MinerIdentity.ConfigFingerprint.
//
// Determinism comes from encoding/json over a struct with fixed field order, so
// the same configuration always yields the same digest across processes. A map
// would not be safe here; Extra lives only in ExecutionEnvironment, which is
// never hashed into the reuse key.
func (s SemanticConfig) Fingerprint() (string, error) {
	if err := s.validate(); err != nil {
		return "", err
	}
	raw, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("localfp: marshal semantic config: %w", err)
	}
	sum := sha256.Sum256(raw)
	// Prefixed so a stored fingerprint is self-describing: a reader can tell which
	// scheme produced it, and a future scheme change is visible instead of silent.
	return "localfp/v1+" + hex.EncodeToString(sum[:]), nil
}

// Digest hashes the execution environment for the provenance record. It is NOT
// part of the reuse key and must never be folded into ConfigFingerprint.
func (e ExecutionEnvironment) Digest() (string, error) {
	raw, err := json.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("localfp: marshal execution environment: %w", err)
	}
	sum := sha256.Sum256(raw)
	return "localexec/v1+" + hex.EncodeToString(sum[:]), nil
}

// Describe renders the semantic configuration as stable, human-readable text for
// diagnostics. When two reuse keys differ unexpectedly, this is what makes the
// cause visible without reversing a hash.
func (s SemanticConfig) Describe() string {
	return fmt.Sprintf(
		"weights=%s quant=%s runtime=%s@%s template=%s grammar=%s temp=%g top_p=%g top_k=%d min_p=%g seed=%d n_ctx=%d",
		short(s.WeightsSHA256), s.Quantization, s.RuntimeName, s.RuntimeVersion,
		s.PromptTemplateVersion, orNone(s.GrammarRevision),
		s.Temperature, s.TopP, s.TopK, s.MinP, s.Seed, s.ContextSize)
}

func short(h string) string {
	if len(h) > 12 {
		return h[:12]
	}
	return h
}

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}
