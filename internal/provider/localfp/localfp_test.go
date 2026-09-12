package localfp

import (
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/provider"
)

func base() SemanticConfig {
	return SemanticConfig{
		WeightsSHA256:         "ca590b42a4e55ae58af480d3d06ae02766b5708ac3587f860a096ce58e438e91",
		Quantization:          "Q8_0",
		RuntimeName:           "llama.cpp",
		RuntimeVersion:        "b1234",
		PromptTemplateVersion: "miner/v1",
		Temperature:           0.2,
		TopP:                  0.95,
		TopK:                  40,
		Seed:                  7,
		ContextSize:           8192,
	}
}

// The motivating case: two quantizations of the same model must NOT share a
// reuse key. This is the collision that would let a Q5 request silently receive
// a revision that Q4 produced.
func TestQuantizationsDoNotCollide(t *testing.T) {
	q4 := base()
	q4.Quantization = "Q4_K_M"
	q4.WeightsSHA256 = strings.Repeat("a", 64)
	q5 := base()
	q5.Quantization = "Q5_K_M"
	q5.WeightsSHA256 = strings.Repeat("b", 64)

	f4, err := q4.Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	f5, err := q5.Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	if f4 == f5 {
		t.Fatal("different quantizations must not share a config fingerprint")
	}

	// And the property that actually matters: distinct miner_version reuse keys.
	v4 := provider.MinerIdentity{ContractVersion: "invariant/v1", ProviderName: "local",
		ModelName: "deepseek-prover-v2-7b", ConfigFingerprint: f4}.Version()
	v5 := provider.MinerIdentity{ContractVersion: "invariant/v1", ProviderName: "local",
		ModelName: "deepseek-prover-v2-7b", ConfigFingerprint: f5}.Version()
	if v4 == v5 {
		t.Fatal("different quantizations must produce different miner_version reuse keys")
	}
}

// Every semantic field must be identity-bearing: changing any one alone must move
// the fingerprint. A field silently excluded from the hash is a latent collision.
func TestEverySemanticFieldIsIdentityBearing(t *testing.T) {
	orig, err := base().Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	mutations := map[string]func(*SemanticConfig){
		"weights":     func(c *SemanticConfig) { c.WeightsSHA256 = strings.Repeat("f", 64) },
		"quant":       func(c *SemanticConfig) { c.Quantization = "Q4_K_M" },
		"runtime":     func(c *SemanticConfig) { c.RuntimeName = "vllm-metal" },
		"runtime_ver": func(c *SemanticConfig) { c.RuntimeVersion = "b9999" },
		"template":    func(c *SemanticConfig) { c.PromptTemplateVersion = "miner/v2" },
		"chat_tmpl":   func(c *SemanticConfig) { c.ChatTemplateSHA256 = "deadbeef" },
		"grammar":     func(c *SemanticConfig) { c.GrammarRevision = "gbnf/v3" },
		"temperature": func(c *SemanticConfig) { c.Temperature = 0.9 },
		"top_p":       func(c *SemanticConfig) { c.TopP = 0.5 },
		"top_k":       func(c *SemanticConfig) { c.TopK = 10 },
		"min_p":       func(c *SemanticConfig) { c.MinP = 0.05 },
		"seed":        func(c *SemanticConfig) { c.Seed = 8 },
		"n_ctx":       func(c *SemanticConfig) { c.ContextSize = 4096 },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			c := base()
			mutate(&c)
			got, err := c.Fingerprint()
			if err != nil {
				t.Fatal(err)
			}
			if got == orig {
				t.Fatalf("changing %s must change the fingerprint", name)
			}
		})
	}
}

// The other half of the contract: execution knobs must NOT move the reuse key,
// or every batch-size change fragments the cache.
func TestExecutionEnvironmentIsNotIdentityBearing(t *testing.T) {
	f1, err := base().Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	// Same semantic config; wildly different execution settings.
	e1 := ExecutionEnvironment{BatchSize: 512, GPULayers: 99, Threads: 8}
	e2 := ExecutionEnvironment{BatchSize: 64, GPULayers: 40, Threads: 2, FlashAttention: true}
	d1, _ := e1.Digest()
	d2, _ := e2.Digest()
	if d1 == d2 {
		t.Fatal("different execution environments should produce different digests")
	}
	f2, err := base().Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	if f1 != f2 {
		t.Fatal("the semantic fingerprint must be independent of execution settings")
	}
}

func TestFingerprintIsDeterministic(t *testing.T) {
	a, err := base().Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 50; i++ {
		b, err := base().Fingerprint()
		if err != nil {
			t.Fatal(err)
		}
		if a != b {
			t.Fatal("fingerprint must be deterministic across calls")
		}
	}
}

// An incomplete fingerprint must ERROR, not hash blanks: two adapters that both
// omit the artifact hash would otherwise agree on a single reuse key.
func TestIncompleteFingerprintIsRejected(t *testing.T) {
	cases := map[string]func(*SemanticConfig){
		"no weights":  func(c *SemanticConfig) { c.WeightsSHA256 = "" },
		"no quant":    func(c *SemanticConfig) { c.Quantization = "" },
		"no runtime":  func(c *SemanticConfig) { c.RuntimeName = "" },
		"no rt ver":   func(c *SemanticConfig) { c.RuntimeVersion = "" },
		"no template": func(c *SemanticConfig) { c.PromptTemplateVersion = "" },
		"no n_ctx":    func(c *SemanticConfig) { c.ContextSize = 0 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			c := base()
			mutate(&c)
			if _, err := c.Fingerprint(); err == nil {
				t.Fatal("an incomplete semantic fingerprint must be rejected")
			}
		})
	}
}

func TestDescribeSurfacesQuantization(t *testing.T) {
	c := base()
	c.Quantization = "Q4_K_M"
	if !strings.Contains(c.Describe(), "Q4_K_M") {
		t.Fatalf("Describe must surface quantization for diagnosis: %s", c.Describe())
	}
}
