package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/store"
)

// SignatureInput requests a mechanism signature under a schema+vocab version.
type SignatureInput struct {
	DBPath        string
	MechanismID   string
	VocabVersion  string
	SchemaVersion string
	JSONOutput    bool
}

// CompareInput requests a comparison of two mechanism signatures.
type CompareInput struct {
	DBPath         string
	MechanismAID   string
	MechanismBID   string
	VocabVersion   string
	SchemaVersion  string
	WeightsVersion string
	// ClassifyVersion selects the classifier contract (classify/v1 default;
	// classify/v2 reproduces the production ASSESSMENT rule). Diagnostic
	// parity: an operator must be able to reproduce an assessment verdict
	// ad hoc without rerunning an experiment.
	ClassifyVersion string
	NoWrite         bool
	JSONOutput      bool
}

// SignatureMechanism builds (or returns the existing) signature for a mechanism.
func (a *App) SignatureMechanism(ctx context.Context, input SignatureInput) (SignatureResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return SignatureResponse{}, err
	}
	defer repoStore.Close()

	record, created, err := a.buildAndPersistSignature(ctx, repoStore, input.MechanismID, input.VocabVersion, input.SchemaVersion)
	if err != nil {
		return SignatureResponse{}, err
	}

	status := "existing"
	if created {
		status = "created"
	}
	return SignatureResponse{
		OK:        true,
		Command:   "mechanism signature",
		Store:     dbPath,
		Status:    status,
		Signature: signatureView(record),
	}, nil
}

// buildAndPersistSignature is the shared path for signature + compare. It loads
// the mechanism, resolves the vocabulary version, builds and persists the
// signature (idempotent), and returns the persisted record.
func (a *App) buildAndPersistSignature(ctx context.Context, repoStore problemStore, mechanismID, vocabVersion, schemaVersion string) (store.SignatureRecord, bool, error) {
	if vocabVersion == "" {
		vocabVersion = canon.VocabularyMechanismV1
	}
	if schemaVersion == "" {
		schemaVersion = canon.SchemaMechanismV1
	}
	if schemaVersion != canon.SchemaMechanismV1 {
		return store.SignatureRecord{}, false, fmt.Errorf("unsupported schema version %q", schemaVersion)
	}

	detail, err := repoStore.GetMechanismDetail(ctx, mechanismID)
	if err != nil {
		return store.SignatureRecord{}, false, err
	}

	vocab, err := a.loadVocabulary(ctx, repoStore, vocabVersion)
	if err != nil {
		return store.SignatureRecord{}, false, err
	}

	input := mechanismInputFromDetail(detail)

	// Merge operator-adjudicated interpretation claims (v33). The status is
	// FIXED here to inferred — an interpretation is a GeneratedInterpretation
	// hypothesis and can never enter as (or be promoted to) an explicit
	// source-backed claim. The provenance ref (adjudication-ledger entry) is
	// carried in the support locator so the claim stays traceable without
	// fabricating a source snapshot.
	interp, err := repoStore.ListInterpretationClaims(ctx, mechanismID)
	if err != nil {
		return store.SignatureRecord{}, false, err
	}
	for _, ic := range interp {
		input.Claims = append(input.Claims, canon.MechanismClaimInput{
			FieldKind:      domain.FieldKind(ic.FieldKind),
			SurfaceLabel:   ic.SurfaceLabel,
			Status:         domain.ClaimInferred,
			SupportLocator: "interpretation:" + ic.ProvenanceRef,
		})
	}

	sig := canon.BuildSignature(input, vocab)

	// Attach a run for provenance.
	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   detail.Approach.ProblemID,
		Operation:   "mechanism signature",
		Status:      domain.RunStatusRunning,
		InputRef:    "mechanism:" + mechanismID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return store.SignatureRecord{}, false, err
	}

	record := signatureRecord(sig, mechanismID, run.ID, now)
	result, err := repoStore.PersistSignature(ctx, record)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return store.SignatureRecord{}, false, err
	}
	if ferr := a.finalizeRun(ctx, repoStore, run.ID, nil); ferr != nil {
		return store.SignatureRecord{}, false, ferr
	}
	return result.Record, result.Created, nil
}

// CompareMechanisms builds both signatures (idempotent) and compares them.
func (a *App) CompareMechanisms(ctx context.Context, input CompareInput) (CompareResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return CompareResponse{}, err
	}
	defer repoStore.Close()

	recordA, _, err := a.buildAndPersistSignature(ctx, repoStore, input.MechanismAID, input.VocabVersion, input.SchemaVersion)
	if err != nil {
		return CompareResponse{}, err
	}
	recordB, _, err := a.buildAndPersistSignature(ctx, repoStore, input.MechanismBID, input.VocabVersion, input.SchemaVersion)
	if err != nil {
		return CompareResponse{}, err
	}

	sigA := signatureFromRecord(recordA)
	sigB := signatureFromRecord(recordB)

	profile, err := canon.ProfileForClassifyVersion(input.ClassifyVersion)
	if err != nil {
		return CompareResponse{}, err
	}
	if input.WeightsVersion != "" && input.WeightsVersion != canon.WeightsMechanismV1 {
		return CompareResponse{}, fmt.Errorf("unknown weights version %q", input.WeightsVersion)
	}
	cmp := canon.CompareWithProfile(sigA, sigB, profile)

	resp := CompareResponse{
		OK:           true,
		Command:      "mechanism compare",
		Store:        dbPath,
		SignatureAID: recordA.ID,
		SignatureBID: recordB.ID,
		FingerprintA: recordA.Fingerprint,
		FingerprintB: recordB.Fingerprint,
		Comparison:   comparisonView(cmp),
	}

	if !input.NoWrite {
		now := a.now()
		detail, derr := repoStore.GetMechanismDetail(ctx, input.MechanismAID)
		if derr != nil {
			return CompareResponse{}, derr
		}
		run, rerr := repoStore.CreateRun(ctx, domain.NewRun{
			ID:          domain.NewRunID(now),
			ProblemID:   detail.Approach.ProblemID,
			Operation:   "mechanism compare",
			Status:      domain.RunStatusRunning,
			InputRef:    "compare:" + input.MechanismAID + ":" + input.MechanismBID,
			ToolName:    "newf",
			ToolVersion: a.version,
			StartedAt:   now,
			CompletedAt: now,
		})
		if rerr != nil {
			return CompareResponse{}, rerr
		}
		comparisonID := domain.NewComparisonRunID(now)
		rec := store.ComparisonRecord{
			ID:              comparisonID,
			SignatureAID:    recordA.ID,
			SignatureBID:    recordB.ID,
			WeightsVersion:  cmp.WeightsVersion,
			ClassifyVersion: cmp.ClassifyVersion,
			Classification:  string(cmp.Classification),
			RunID:           run.ID,
			CreatedAt:       now.Format(timeLayout),
		}
		for _, f := range cmp.Fields {
			rec.Fields = append(rec.Fields, store.ComparisonFieldResultRow{
				FieldKind:    string(f.FieldKind),
				OverlapCount: f.OverlapCount,
				UnionCount:   f.UnionCount,
				Jaccard:      f.Jaccard,
				Ordinal:      string(f.Ordinal),
				Incomparable: f.Incomparable,
			})
		}
		if err := repoStore.PersistComparison(ctx, rec); err != nil {
			a.failRun(ctx, repoStore, run.ID, err)
			return CompareResponse{}, err
		}
		if err := a.finalizeRun(ctx, repoStore, run.ID, nil); err != nil {
			return CompareResponse{}, err
		}
		resp.ComparisonRunID = comparisonID
	}

	return resp, nil
}

// mechanismInputFromDetail maps a store.ApproachDetail into the store-free
// canon.MechanismInput. Support locators are matched to claims by field path
// where available; when absent the claim carries no support locator.
func mechanismInputFromDetail(detail store.ApproachDetail) canon.MechanismInput {
	// Index support by field path for provenance lookup.
	supportByField := map[string]domain.SourceSupport{}
	for _, s := range detail.Support {
		supportByField[s.FieldPath] = s
	}

	in := canon.MechanismInput{
		MechanismID: detail.Mechanism.ID,
		Posture: canon.Posture{
			Locality:     detail.Mechanism.Locality,
			Construction: detail.Mechanism.ConstructionMode,
			Uncertainty:  detail.Mechanism.UncertaintyMode,
		},
		OutcomeClass: detail.Outcome.Class,
		// Posture/outcome provenance is preserved from #7 dotted-path support,
		// defaulting to unknown (never explicit) when no support row exists.
		PostureProvenance: canon.PostureProvenance{
			Locality:     claimStatusForPaths(supportByField, domain.PostureSupportPaths("locality")),
			Construction: claimStatusForPaths(supportByField, domain.PostureSupportPaths("construction")),
			Uncertainty:  claimStatusForPaths(supportByField, domain.PostureSupportPaths("uncertainty")),
		},
		OutcomeProvenance: claimStatusForPaths(supportByField, domain.OutcomeSupportPaths()),
	}

	// Justified completeness declarations (v25/v26): only ACCEPTED declarations
	// acquire evaluation authority; declared_only rows are auditable claims
	// that keep the conservative unobserved default.
	if len(detail.FieldCompleteness) > 0 {
		for _, fc := range detail.FieldCompleteness {
			if fc.Admission != domain.CompletenessAccepted {
				continue
			}
			kind, err := domain.AttributeFieldKind(fc.Kind)
			if err != nil {
				continue
			}
			if in.DeclaredCompleteness == nil {
				in.DeclaredCompleteness = map[domain.FieldKind]domain.FieldCompleteness{}
			}
			in.DeclaredCompleteness[kind] = fc.Completeness
		}
	}

	for _, attr := range detail.Attributes {
		fieldKind, err := domain.AttributeFieldKind(attr.Kind)
		if err != nil {
			continue
		}
		paths := domain.AttributeSupportPaths(attr.Kind)
		claim := canon.MechanismClaimInput{
			FieldKind:    fieldKind,
			SurfaceLabel: attr.Value,
			Status:       claimStatusForPaths(supportByField, paths),
		}
		if sup, ok := lookupSupport(supportByField, paths); ok {
			claim.SupportSnapshotID = sup.SnapshotID
			claim.SupportLocator = sup.Locator
			claim.Confidence = sup.Confidence
		}
		in.Claims = append(in.Claims, claim)
	}

	// Boundary provenance is split across two distinct #7 fields:
	//   * outcome.boundary_statement — a single statement-level boundary that
	//     carries its own support (outcome.boundary_statement); and
	//   * enumerated outcome.boundary_conditions (detail.Boundaries) — which
	//     have NO per-condition support path in #7.
	// The statement support must attach only to the statement-derived boundary,
	// never be broadcast across the enumerated conditions (that would conflate
	// and over-attribute provenance). Enumerated conditions carry unknown until
	// a per-condition support path exists.
	if statement := detail.Outcome.BoundaryStatement; statement != "" {
		bin := canon.MechanismBoundaryInput{
			SurfaceLabel: statement,
			Relation:     "stops_at",
			Status:       claimStatusForPaths(supportByField, []string{domain.SupportPathBoundaryStatement}),
		}
		if sup, ok := lookupSupport(supportByField, []string{domain.SupportPathBoundaryStatement}); ok {
			bin.SupportSnapshotID = sup.SnapshotID
			bin.SupportLocator = sup.Locator
		}
		in.Boundaries = append(in.Boundaries, bin)
	}
	for _, b := range detail.Boundaries {
		in.Boundaries = append(in.Boundaries, canon.MechanismBoundaryInput{
			SurfaceLabel: b.Condition,
			Relation:     "stops_at",
			Status:       domain.ClaimUnknown,
		})
	}
	return in
}

// lookupSupport resolves a support row by trying accepted field paths in order
// (canonical first, historical aliases after). It returns the first match. The
// alias tolerance exists so authored provenance is not silently demoted to
// unknown by field-path spelling drift between the #7 writer (which stores the
// verbatim JSON-key path, e.g. mechanism.operators / mechanism.uncertainty_mode
// / outcome.boundary_statement) and the #9 reader.
func lookupSupport(byField map[string]domain.SourceSupport, paths []string) (domain.SourceSupport, bool) {
	for _, p := range paths {
		if sup, ok := byField[p]; ok {
			return sup, true
		}
	}
	return domain.SourceSupport{}, false
}

// claimStatusForPaths maps the resolved support (across accepted paths) to a
// ClaimStatus, preserving unknown when no row is present so an unprovenanced
// field is never promoted to explicit.
func claimStatusForPaths(byField map[string]domain.SourceSupport, paths []string) domain.ClaimStatus {
	sup, ok := lookupSupport(byField, paths)
	if !ok {
		return domain.ClaimUnknown
	}
	return claimStatusFromSupport(sup)
}

// claimStatusFromSupport maps a resolved #7 SupportKind to a ClaimStatus.
// Absence of a support row (handled by callers via claimStatusForPaths) is NOT
// treated as explicit: #7 does not require a support row for every populated
// field, so a missing row means the provenance is unknown, not source-backed.
// Defaulting to explicit would silently promote an unprovenanced value into an
// explicit source-backed claim, violating the epistemic-status hard constraint.
func claimStatusFromSupport(sup domain.SourceSupport) domain.ClaimStatus {
	switch sup.SupportKind {
	case domain.SupportExplicit:
		return domain.ClaimExplicit
	case domain.SupportInferred:
		return domain.ClaimInferred
	case domain.SupportUnsupported:
		return domain.ClaimUnsupported
	default:
		return domain.ClaimUnknown
	}
}

// signatureRecord flattens a canon.MechanismSignature into store rows.
func signatureRecord(sig canon.MechanismSignature, mechanismID, runID string, now time.Time) store.SignatureRecord {
	rec := store.SignatureRecord{
		ID:                domain.NewMechanismSignatureID(now),
		MechanismID:       mechanismID,
		SchemaVersion:     sig.SchemaVersion,
		VocabularyVersion: sig.VocabularyVersion,
		Fingerprint:       canon.Fingerprint(sig),
		RunID:             runID,
		CreatedAt:         now.Format(timeLayout),
		OutcomeClass:      string(sig.OutcomeClass),
		OutcomeStatus:     string(sig.OutcomeProvenance),
		Posture: map[string]string{
			"locality":     string(sig.Posture.Locality),
			"construction": string(sig.Posture.Construction),
			"uncertainty":  string(sig.Posture.Uncertainty),
		},
		PostureStatus: map[string]string{
			"locality":     string(sig.PostureProvenance.Locality),
			"construction": string(sig.PostureProvenance.Construction),
			"uncertainty":  string(sig.PostureProvenance.Uncertainty),
		},
	}
	appendClaims := func(claims []canon.FieldClaim) {
		for i, c := range claims {
			rec.FieldClaims = append(rec.FieldClaims, store.SignatureFieldClaimRow{
				FieldKind:          string(c.FieldKind),
				SurfaceLabel:       c.SurfaceLabel,
				ResolutionState:    string(c.State),
				CanonicalID:        string(c.CanonicalID),
				ClaimStatus:        string(c.Status),
				SupportSnapshotID:  c.SupportSnapshotID,
				SupportLocator:     c.SupportLocator,
				Confidence:         c.Confidence,
				ClassifierContract: c.ClassifierContract,
				Ordinal:            i,
			})
		}
	}
	appendClaims(sig.Representations)
	appendClaims(sig.Operators)
	appendClaims(sig.Assumptions)
	appendClaims(sig.Preserves)
	appendClaims(sig.Breaks)
	appendClaims(sig.AuxiliaryObjects)
	for i, b := range sig.Boundaries {
		rec.Boundaries = append(rec.Boundaries, store.SignatureBoundaryRow{
			SurfaceLabel:      b.SurfaceLabel,
			ResolutionState:   string(b.State),
			CanonicalID:       string(b.CanonicalID),
			Relation:          b.Relation,
			ClaimStatus:       string(b.Status),
			SupportSnapshotID: b.SupportSnapshotID,
			SupportLocator:    b.SupportLocator,
			Ordinal:           i,
		})
	}
	return rec
}

// signatureFromRecord rebuilds a canon.MechanismSignature from persisted rows so
// comparison runs against exactly what was stored.
func signatureFromRecord(rec store.SignatureRecord) canon.MechanismSignature {
	sig := canon.MechanismSignature{
		SchemaVersion:     rec.SchemaVersion,
		VocabularyVersion: rec.VocabularyVersion,
		MechanismID:       rec.MechanismID,
		SignatureID:       rec.ID,
		OutcomeClass:      domain.OutcomeClass(rec.OutcomeClass),
		Posture: canon.Posture{
			Locality:     domain.Locality(rec.Posture["locality"]),
			Construction: domain.ConstructionMode(rec.Posture["construction"]),
			Uncertainty:  domain.UncertaintyMode(rec.Posture["uncertainty"]),
		},
		Representations:  []canon.FieldClaim{},
		Operators:        []canon.FieldClaim{},
		Assumptions:      []canon.FieldClaim{},
		Preserves:        []canon.FieldClaim{},
		Breaks:           []canon.FieldClaim{},
		AuxiliaryObjects: []canon.FieldClaim{},
		Boundaries:       []canon.Boundary{},
	}
	for _, c := range rec.FieldClaims {
		fc := canon.FieldClaim{
			FieldKind:          domain.FieldKind(c.FieldKind),
			SurfaceLabel:       c.SurfaceLabel,
			State:              domain.ResolutionState(c.ResolutionState),
			CanonicalID:        domain.CanonicalID(c.CanonicalID),
			Status:             domain.ClaimStatus(c.ClaimStatus),
			SupportSnapshotID:  c.SupportSnapshotID,
			SupportLocator:     c.SupportLocator,
			Confidence:         c.Confidence,
			ClassifierContract: c.ClassifierContract,
		}
		switch domain.FieldKind(c.FieldKind) {
		case domain.FieldRepresentation:
			sig.Representations = append(sig.Representations, fc)
		case domain.FieldOperator:
			sig.Operators = append(sig.Operators, fc)
		case domain.FieldAssumption:
			sig.Assumptions = append(sig.Assumptions, fc)
		case domain.FieldPreserves:
			sig.Preserves = append(sig.Preserves, fc)
		case domain.FieldBreaks:
			sig.Breaks = append(sig.Breaks, fc)
		case domain.FieldAuxiliaryObject:
			sig.AuxiliaryObjects = append(sig.AuxiliaryObjects, fc)
		}
	}
	for _, b := range rec.Boundaries {
		sig.Boundaries = append(sig.Boundaries, canon.Boundary{
			SurfaceLabel: b.SurfaceLabel,
			State:        domain.ResolutionState(b.ResolutionState),
			CanonicalID:  domain.CanonicalID(b.CanonicalID),
			Relation:     b.Relation,
		})
	}
	// Rehydrate the justified completeness declarations (v25/v26) so a
	// persisted signature evaluates absence exactly as the build path did:
	// only ACCEPTED declarations upgrade a field; declared_only claims and
	// undeclared fields stay unobserved.
	if len(rec.FieldCompleteness) > 0 {
		for _, fc := range rec.FieldCompleteness {
			if fc.Admission != domain.CompletenessAccepted {
				continue
			}
			kind, err := domain.AttributeFieldKind(fc.Kind)
			if err != nil {
				continue
			}
			if sig.SetFieldCompleteness == nil {
				sig.SetFieldCompleteness = map[domain.FieldKind]domain.FieldCompleteness{}
			}
			sig.SetFieldCompleteness[kind] = fc.Completeness
		}
	}
	return sig
}
