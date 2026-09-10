package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/instagrim-dev/newf/internal/normalize"
)

// FixtureProviderName is the recorded provider identity for the deterministic
// fixture normalizer.
const (
	FixtureProviderName    = "fixture"
	FixtureProviderVersion = "v1"
	FixtureModelName       = "deterministic-fixture"
)

// fixtureBlockStart and fixtureBlockEnd delimit the embedded normalization
// contract inside a fixture source. Keeping the expected output alongside the
// prose lets a fixture demonstrate that differing surface wording can still map
// to comparable normalized mechanism fields, while remaining fully offline and
// reproducible.
const (
	fixtureBlockStart = "<!-- newf-normalize"
	fixtureBlockEnd   = "newf-normalize -->"
)

// ErrProviderTransport marks a simulated provider transport failure, distinct
// from a semantic skip or schema violation.
var ErrProviderTransport = errors.New("provider transport failure")

// FixtureNormalizer is a deterministic, network-free Normalizer. It derives its
// structured output from an embedded contract block in the snapshot content, so
// CLI and store behavior is reproducible in CI without model access.
type FixtureNormalizer struct{}

// NewFixtureNormalizer constructs the deterministic fixture normalizer.
func NewFixtureNormalizer() *FixtureNormalizer {
	return &FixtureNormalizer{}
}

// Normalize returns a deterministic result for the request. Non-text content is
// reported as a typed skip; text without an embedded contract is a typed
// no-approaches skip; malformed embedded JSON is a schema violation.
func (f *FixtureNormalizer) Normalize(_ context.Context, req normalize.Request) (NormalizeResponse, error) {
	metadata := Metadata{
		ProviderName:    FixtureProviderName,
		ProviderVersion: FixtureProviderVersion,
		ModelName:       FixtureModelName,
		SchemaVersion:   normalize.SchemaVersion,
	}
	requestPayload := fmt.Sprintf("snapshot=%s media_type=%s sha256=%s", req.SnapshotID, req.MediaType, hashContent(req.Content))

	if !isProbablyText(req.MediaType, req.Content) {
		result := normalize.Result{
			SchemaVersion: normalize.SchemaVersion,
			Skipped:       true,
			SkipReason:    normalize.SkipUnsupportedContentRepresentation,
		}
		return NormalizeResponse{
			Result:          result,
			Metadata:        metadata,
			RequestPayload:  requestPayload,
			ResponsePayload: mustJSON(result),
		}, nil
	}

	block, found := extractBlock(req.Content)
	if !found {
		result := normalize.Result{
			SchemaVersion: normalize.SchemaVersion,
			Skipped:       true,
			SkipReason:    normalize.SkipNoApproachesFound,
		}
		return NormalizeResponse{
			Result:          result,
			Metadata:        metadata,
			RequestPayload:  requestPayload,
			ResponsePayload: mustJSON(result),
		}, nil
	}

	var result normalize.Result
	if err := json.Unmarshal([]byte(block), &result); err != nil {
		return NormalizeResponse{}, normalize.ErrSchemaViolation
	}
	if result.SchemaVersion == "" {
		result.SchemaVersion = normalize.SchemaVersion
	}

	return NormalizeResponse{
		Result:          result,
		Metadata:        metadata,
		RequestPayload:  requestPayload,
		ResponsePayload: mustJSON(result),
	}, nil
}

func extractBlock(content string) (string, bool) {
	start := strings.Index(content, fixtureBlockStart)
	if start < 0 {
		return "", false
	}
	rest := content[start+len(fixtureBlockStart):]
	end := strings.Index(rest, fixtureBlockEnd)
	if end < 0 {
		return "", false
	}
	return strings.TrimSpace(rest[:end]), true
}

func isProbablyText(mediaType, content string) bool {
	lower := strings.ToLower(strings.TrimSpace(mediaType))
	switch {
	case strings.HasPrefix(lower, "text/"),
		lower == "application/json",
		lower == "application/yaml",
		lower == "application/x-yaml":
		// still verify bytes below
	case strings.Contains(lower, "pdf"),
		strings.HasPrefix(lower, "image/"),
		strings.HasPrefix(lower, "audio/"),
		strings.HasPrefix(lower, "video/"),
		lower == "application/octet-stream":
		return false
	}
	if content == "" {
		return false
	}
	if !utf8.ValidString(content) {
		return false
	}
	if strings.IndexByte(content, 0x00) >= 0 {
		return false
	}
	return true
}

func hashContent(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func mustJSON(result normalize.Result) string {
	raw, err := json.Marshal(result)
	if err != nil {
		return "{}"
	}
	return string(raw)
}
