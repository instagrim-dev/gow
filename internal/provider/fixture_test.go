package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/normalize"
)

const fixtureContent = `# Modular residue-cover

Prose describing an approach.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/modular-residue-cover",
      "label": "modular residue-cover construction",
      "mechanism": {
        "representations": ["congruence classes"],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic"
      },
      "outcome": {"class": "partial_failure", "boundary_statement": "uncovered residue families remain"},
      "support": [{"field_path": "outcome.class", "support_kind": "explicit", "locator": "para:1"}]
    }
  ]
}
newf-normalize -->
`

func TestFixtureNormalizerExtractsEmbeddedResult(t *testing.T) {
	t.Parallel()
	normalizer := NewFixtureNormalizer()
	resp, err := normalizer.Normalize(context.Background(), normalize.Request{
		SnapshotID:    "snap_01K4Y8X6YJJ66Y5QY9G7DNE1H4",
		MediaType:     "text/markdown",
		Content:       fixtureContent,
		SchemaVersion: normalize.SchemaVersion,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if resp.Result.Skipped {
		t.Fatalf("expected non-skipped result, got skip %q", resp.Result.SkipReason)
	}
	if len(resp.Result.Approaches) != 1 {
		t.Fatalf("approaches = %d, want 1", len(resp.Result.Approaches))
	}
	if resp.Metadata.ProviderName != FixtureProviderName {
		t.Fatalf("provider name = %q, want %q", resp.Metadata.ProviderName, FixtureProviderName)
	}
	if resp.RequestPayload == "" || resp.ResponsePayload == "" {
		t.Fatal("expected request/response payloads retained for audit")
	}
	if err := resp.Result.Validate(); err != nil {
		t.Fatalf("provider result failed schema validation: %v", err)
	}
}

func TestFixtureNormalizerSkipsBinaryContent(t *testing.T) {
	t.Parallel()
	normalizer := NewFixtureNormalizer()
	resp, err := normalizer.Normalize(context.Background(), normalize.Request{
		SnapshotID:    "snap_01K4Y8X6YJJ66Y5QY9G7DNE1H5",
		MediaType:     "application/pdf",
		Content:       "%PDF-1.4\nbinary",
		SchemaVersion: normalize.SchemaVersion,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if !resp.Result.Skipped || resp.Result.SkipReason != normalize.SkipUnsupportedContentRepresentation {
		t.Fatalf("expected unsupported-content skip, got %+v", resp.Result)
	}
}

func TestFixtureNormalizerSkipsTextWithoutBlock(t *testing.T) {
	t.Parallel()
	normalizer := NewFixtureNormalizer()
	resp, err := normalizer.Normalize(context.Background(), normalize.Request{
		SnapshotID:    "snap_01K4Y8X6YJJ66Y5QY9G7DNE1H6",
		MediaType:     "text/markdown",
		Content:       "# just prose, no contract block",
		SchemaVersion: normalize.SchemaVersion,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if !resp.Result.Skipped || resp.Result.SkipReason != normalize.SkipNoApproachesFound {
		t.Fatalf("expected no-approaches skip, got %+v", resp.Result)
	}
}

func TestFixtureNormalizerRejectsMalformedBlock(t *testing.T) {
	t.Parallel()
	normalizer := NewFixtureNormalizer()
	content := "text\n<!-- newf-normalize\n{not json}\nnewf-normalize -->\n"
	_, err := normalizer.Normalize(context.Background(), normalize.Request{
		SnapshotID:    "snap_01K4Y8X6YJJ66Y5QY9G7DNE1H7",
		MediaType:     "text/markdown",
		Content:       content,
		SchemaVersion: normalize.SchemaVersion,
	})
	if err == nil || !strings.Contains(err.Error(), "schema violation") {
		t.Fatalf("expected schema violation, got %v", err)
	}
}

func TestFixtureNormalizerDeterministic(t *testing.T) {
	t.Parallel()
	normalizer := NewFixtureNormalizer()
	req := normalize.Request{
		SnapshotID:    "snap_01K4Y8X6YJJ66Y5QY9G7DNE1H4",
		MediaType:     "text/markdown",
		Content:       fixtureContent,
		SchemaVersion: normalize.SchemaVersion,
	}
	first, err := normalizer.Normalize(context.Background(), req)
	if err != nil {
		t.Fatalf("Normalize(first) error = %v", err)
	}
	second, err := normalizer.Normalize(context.Background(), req)
	if err != nil {
		t.Fatalf("Normalize(second) error = %v", err)
	}
	if first.ResponsePayload != second.ResponsePayload {
		t.Fatal("fixture normalizer is not deterministic")
	}
}
