// Package provider defines replaceable, role-specific model-native operator
// interfaces and the deterministic fixtures used to exercise them offline.
//
// Domain records never depend on this package's concrete adapters, and this
// package never leaks vendor SDK types into the domain. Each concrete provider
// records the operator role it performed for provenance.
package provider

import (
	"context"

	"github.com/instagrim-dev/newf/internal/normalize"
)

// Metadata identifies the provider/model/config behind one invocation. It is
// persisted (minus any secrets) so a normalization revision is replayable.
type Metadata struct {
	ProviderName    string
	ProviderVersion string
	ModelName       string
	SchemaVersion   string
}

// NormalizeResponse pairs the structured normalization output with the provider
// metadata and the raw request/response payloads retained for audit/replay.
type NormalizeResponse struct {
	Result          normalize.Result
	Metadata        Metadata
	RequestPayload  string
	ResponsePayload string
}

// Normalizer is the replaceable provider-facing interface for the normalize
// operator. Transport failures must be returned as errors; a snapshot the
// provider understands but cannot normalize is reported via a skipped Result.
type Normalizer interface {
	Normalize(ctx context.Context, req normalize.Request) (NormalizeResponse, error)
}
