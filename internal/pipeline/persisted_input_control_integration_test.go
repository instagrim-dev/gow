package pipeline

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/normalize"
	"github.com/instagrim-dev/newf/internal/provider"
)

// This file is the PERSISTED-INPUT positive control (next milestone after the
// fixture-backed control in positive_control_integration_test.go): the same
// synthetic scenario enters through ORDINARY ingestion and normalization,
// carries JUSTIFIED completeness and vocabulary mappings through storage, and
// is then challenged (CLI-default challenger), generated against (deterministic
// proposal), and assessed — with the exact records asserted at each boundary:
//
//	ingest      -> persisted source snapshots (sha-addressed content)
//	normalize   -> persisted revision + approaches + JUSTIFIED completeness rows
//	signature   -> vocabulary-resolved claims + completeness ROUND-TRIP (v25)
//	cluster     -> three real mechanism families
//	failure-space -> outcome partition (2 failure + 1 partial_success)
//	mine        -> CLI-DEFAULT miner, explicit MinSupport 2 -> ONE candidate
//	challenge   -> CLI-DEFAULT challenger -> three completed negatives -> surviving
//	generate    -> deterministic proposal; persisted violation verdict violates
//	assess      -> decisive structural recovery of the withheld target
//
// The proposal stays deterministic (per the milestone), but its persisted
// resolved claims are ASSERTED to agree with the pinned vocabulary's own
// resolution of their surface labels — the control does not trust a
// provider-supplied `resolved` status or a silently invented alias.

// controlApproach builds one normalize-contract approach for the control
// corpus. When declare is true it carries the JUSTIFIED completeness
// declaration for preserves — the enforceable "all entries of the declared
// payload's preserves list were parsed" scope, not an unqualified authority
// claim.
func controlApproach(id, label string, preserves, operators []string, locality domain.Locality, class domain.OutcomeClass, declare bool) normalize.Approach {
	mech := normalize.Mechanism{
		Preserves:        preserves,
		Operators:        operators,
		Locality:         locality,
		ConstructionMode: domain.ConstructionConstructive,
		UncertaintyMode:  domain.UncertaintyDeterministic,
	}
	if declare {
		// Every set-valued list of the declared payload is exhaustive by
		// construction (including the empty ones): under the corrected
		// missing-data contract (classify/v2) a decisive comparison needs
		// justified completeness on every decisive axis — unobserved empty
		// fields are epistemic gaps, not agreement.
		mech.FieldCompleteness = map[string]string{
			"preserves": "complete", "operator": "complete", "assumption": "complete",
			"breaks": "complete", "auxiliary_object": "complete",
		}
		mech.CompletenessScope = string(domain.ScopeDeclaredPayload)
		mech.CompletenessBasis = "synthetic control corpus: the declared payload's set-valued lists are exhaustive by construction"
	}
	return normalize.Approach{
		LogicalIdentity: id,
		Label:           label,
		Mechanism:       mech,
		Outcome:         normalize.Outcome{Class: class},
		Support: []normalize.FieldSupport{
			{FieldPath: "mechanism.preserves", SupportKind: domain.SupportExplicit},
			{FieldPath: "outcome.class", SupportKind: domain.SupportExplicit},
		},
	}
}

// controlTrainResult is the train-side failure atlas: two mechanistically
// distinct FAILURE approaches sharing preserves "residue locality" (the
// conserved failure structure the miner must find) plus one partial_success
// contrast that does NOT preserve it (so the contrast verdict is a decisive
// violation under the declared completeness).
func controlTrainResult(declare bool) normalize.Result {
	return normalize.Result{
		SchemaVersion: normalize.SchemaVersion,
		Approaches: []normalize.Approach{
			controlApproach("failed-residue-modular", "Residue-local modular decomposition (failed)",
				[]string{"residue locality"}, []string{"modular decomposition"}, domain.LocalityLocal, domain.OutcomeFailure, declare),
			controlApproach("failed-residue-averaging", "Residue-local density averaging (failed)",
				[]string{"residue locality"}, []string{"density averaging"}, domain.LocalityLocal, domain.OutcomeFailure, declare),
			controlApproach("contrast-mean-growth", "Mean-growth bound (partial success)",
				[]string{"mean growth rate"}, nil, domain.LocalityGlobal, domain.OutcomePartialSuccess, declare),
		},
	}
}

// controlTargetResult is the withheld target: the structural advance that
// escaped the failure invariant (it preserves mean growth, not residue
// locality). Its labels resolve through the REAL vocabulary-admission path.
func controlTargetResult() normalize.Result {
	return normalize.Result{
		SchemaVersion: normalize.SchemaVersion,
		Approaches: []normalize.Approach{
			controlApproach("withheld-mean-growth-advance", "Global mean-growth bound (withheld advance)",
				[]string{"mean growth rate"}, nil, domain.LocalityGlobal, domain.OutcomeSuccess, true),
		},
	}
}

// writeControlDoc embeds the normalize contract in a markdown document exactly
// the way the ordinary corpus does, and writes it to disk for ingestion.
func writeControlDoc(t *testing.T, dir, name string, result normalize.Result) string {
	t.Helper()
	raw, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("marshal control payload: %v", err)
	}
	content := "# " + name + "\n\nSynthetic control corpus document.\n\n<!-- newf-normalize\n" + string(raw) + "\nnewf-normalize -->\n"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write control doc: %v", err)
	}
	return path
}

// ingestNormalizeSign drives one problem through the ORDINARY path: ingest the
// document, normalize every snapshot with the default (fixture) provider, and
// compute canonical signatures for every mechanism.
func ingestNormalizeSign(t *testing.T, ctx context.Context, app *App, dbPath, problemID, docPath string) {
	t.Helper()
	if _, err := app.IngestSources(ctx, IngestInput{DBPath: dbPath, ProblemID: problemID, Paths: []string{docPath}}); err != nil {
		t.Fatalf("ingest %s: %v", docPath, err)
	}
	if _, err := app.Normalize(ctx, NormalizeInput{DBPath: dbPath, ProblemID: problemID, All: true}); err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if _, err := app.SignatureAllMechanisms(ctx, SignatureBatchInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("signature batch: %v", err)
	}
}

// TestIntegrationPersistedInputPositiveControl is the persisted-input positive
// control (see file header). Every boundary is asserted on PERSISTED records.
func TestIntegrationPersistedInputPositiveControl(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.normalizers = map[string]provider.Normalizer{provider.FixtureProviderName: provider.NewFixtureNormalizer()}
	// CLI-DEFAULT miner and challenger (deliberately NOT injected).
	prop := pcSignature(pcMeanGrowth, "", true)
	prop.Preserves[0].SurfaceLabel = "mean growth rate" // vocab-consistency is asserted below
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{prop}}

	docDir := t.TempDir()
	trainDoc := writeControlDoc(t, docDir, "control-train.md", controlTrainResult(true))
	targetDoc := writeControlDoc(t, docDir, "control-target.md", controlTargetResult())

	trainInit, err := app.InitProblem(ctx, InitProblemInput{DBPath: dbPath, Statement: "persisted-input control (train)", Slug: "pic-train"})
	if err != nil {
		t.Fatalf("init train: %v", err)
	}
	targetInit, err := app.InitProblem(ctx, InitProblemInput{DBPath: dbPath, Statement: "persisted-input control (target)", Slug: "pic-target", ForceNew: true})
	if err != nil {
		t.Fatalf("init target: %v", err)
	}
	train, target := trainInit.ProblemID, targetInit.ProblemID

	// Boundary 1-3: ingest -> normalize -> signature through the ORDINARY path.
	ingestNormalizeSign(t, ctx, app, dbPath, train, trainDoc)
	ingestNormalizeSign(t, ctx, app, dbPath, target, targetDoc)

	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()

	// Boundary 1: the snapshot is persisted, content-addressed.
	snaps, err := repo.ListSourcesByProblem(ctx, train)
	if err != nil || len(snaps) != 1 {
		t.Fatalf("train must own exactly one persisted source: %v %d", err, len(snaps))
	}

	// Boundary 2: normalization persisted the approaches AND the justified
	// completeness declarations (v25 rows carry the basis verbatim).
	approaches, err := repo.ListApproaches(ctx, train)
	if err != nil {
		t.Fatalf("list approaches: %v", err)
	}
	if len(approaches) != 3 {
		t.Fatalf("train must normalize exactly 3 approaches, got %d", len(approaches))
	}
	for _, item := range approaches {
		detail, err := repo.GetApproachDetail(ctx, item.Approach.ID)
		if err != nil {
			t.Fatalf("approach detail: %v", err)
		}
		// All five decisive-axis declarations persist (the corrected
		// missing-data contract needs every decisive axis justified), with
		// preserves verified explicitly.
		if len(detail.FieldCompleteness) != 5 {
			t.Fatalf("approach %s must persist five justified declarations, got %+v",
				item.Approach.LogicalIdentity, detail.FieldCompleteness)
		}
		sawPreserves := false
		for _, fc := range detail.FieldCompleteness {
			if fc.Completeness != domain.CompletenessComplete || fc.Basis == "" {
				t.Fatalf("approach %s declaration must be complete with a basis: %+v", item.Approach.LogicalIdentity, fc)
			}
			if fc.Kind == domain.AttrPreserves {
				sawPreserves = true
			}
		}
		if !sawPreserves {
			t.Fatalf("approach %s missing the preserves declaration: %+v", item.Approach.LogicalIdentity, detail.FieldCompleteness)
		}
		// Boundary 3: the persisted signature ROUND-TRIPS the declaration and
		// resolves labels through the pinned vocabulary. buildAndPersistSignature
		// is idempotent: it returns the EXISTING persisted record.
		rec, created, err := app.buildAndPersistSignature(ctx, repo, detail.Mechanism.ID, "", "")
		if err != nil {
			t.Fatalf("load signature: %v", err)
		}
		if created {
			t.Fatal("signature must already exist from the ordinary batch pass")
		}
		sig := signatureFromRecord(rec)
		if got := sig.FieldCompleteness(domain.FieldPreserves); got != domain.CompletenessComplete {
			t.Fatalf("rehydrated signature preserves completeness = %q, want complete (v25 round-trip)", got)
		}
		if got := sig.FieldCompleteness(domain.FieldOperator); len(detail.FieldCompleteness) == 1 && got != domain.CompletenessUnobserved {
			t.Fatalf("undeclared operator field must stay unobserved, got %q", got)
		}
		for _, claim := range sig.Preserves {
			if claim.State != domain.ResolutionResolved {
				t.Fatalf("corpus preserves label %q must resolve through the vocabulary, got state %q", claim.SurfaceLabel, claim.State)
			}
		}
	}

	// Boundary 4-5: real clustering + outcome partition.
	if _, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: train}); err != nil {
		t.Fatalf("cluster build: %v", err)
	}
	fsResp, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: train})
	if err != nil {
		t.Fatalf("failure-space build: %v", err)
	}
	if fsResp.FailureSpace.DistinctFamilyCount != 3 {
		t.Fatalf("expected 3 distinct families, got %d", fsResp.FailureSpace.DistinctFamilyCount)
	}
	byClass := map[string]int{}
	for _, o := range fsResp.FailureSpace.Outcomes {
		byClass[o.OutcomeClass] = o.FamilyCount
	}
	if byClass["failure"] != 2 || byClass["partial_success"] != 1 {
		t.Fatalf("outcome partition = %+v, want failure:2 partial_success:1", byClass)
	}

	// Boundary 6: CLI-DEFAULT miner, explicit MinSupport 2 -> exactly one
	// candidate with the known ground-truth predicate, support 2, and a
	// DECISIVE contrast violation (reachable only because the declared
	// completeness survived storage: unobserved contrast would read unknown).
	mined, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: train, MinSupport: 2})
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if mined.Revision.CandidateCount != 1 {
		t.Fatalf("expected exactly one mined candidate, got %d", mined.Revision.CandidateCount)
	}
	cand := mined.Revision.Candidates[0]
	wantFP := invariant.Predicate{
		Schema: invariant.PredicateSchemaV1,
		Root:   invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: pcResidueLocality},
	}.Fingerprint()
	if cand.PredicateFingerprint != wantFP {
		t.Fatalf("candidate fingerprint = %q, want ground-truth %q", cand.PredicateFingerprint, wantFP)
	}
	if cand.DistinctFamilySupport != 2 {
		t.Fatalf("candidate support = %d, want 2", cand.DistinctFamilySupport)
	}
	if cand.ContrastViolatingNum != 1 {
		t.Fatalf("contrast violation must be decisive under declared completeness, got %d", cand.ContrastViolatingNum)
	}

	// Boundary 7: CLI-DEFAULT challenger. All three derived attacks
	// (known-counterexample, success-preserving, bias-critique) complete as
	// negatives on this substrate, so survival is EARNED, not granted.
	chResp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: cand.ID})
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}
	if got := chResp.Reports[0].StateAfter; got != "surviving" {
		t.Fatalf("candidate must survive the default campaign, got %q", got)
	}
	if got := len(chResp.Reports[0].Challenges); got != 3 {
		t.Fatalf("default campaign must record its three derived attacks, got %d", got)
	}

	// Boundary 8-9: experiment against the withheld target. The deterministic
	// proposal preserves ONLY mean_growth_rate (complete), so it (a) verifiably
	// violates the surviving invariant and (b) is mechanism-near the withheld
	// advance whose signature entered through ordinary normalization.
	if _, err := app.DefineExperiment(ctx, ExperimentDefineInput{DBPath: dbPath, ProblemID: train, TargetProblemID: target}); err != nil {
		t.Fatalf("define: %v", err)
	}
	runResp, err := app.RunExperiment(ctx, ExperimentRunInput{DBPath: dbPath, ProblemID: train})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	exp := runResp.Experiment
	if !exp.LeakageCheck.Passed {
		t.Fatalf("clean split must pass the leakage audit: %+v", exp.LeakageCheck)
	}
	b0, b3 := armByName(t, exp, "b0_undirected"), armByName(t, exp, "b3_invariant_guided")
	if b0.ProposalCount != 0 || b0.Recovered {
		t.Fatalf("no-target B0 must be empty by construction: %+v", b0)
	}
	if !b3.Recovered || b3.UnknownCount != 0 || b3.UnassessedCount != 0 {
		t.Fatalf("guided arm must decisively recover the withheld advance: %+v", b3)
	}
	if exp.Conclusion != "structural_recovery" {
		t.Fatalf("conclusion = %q, want structural_recovery", exp.Conclusion)
	}

	// The persisted violation verdict is a VERIFIED break of exactly the
	// surviving invariant (read back, not inferred from recovery).
	gen, err := repo.GetFrontierGeneration(ctx, b3.FrontierGenerationRun)
	if err != nil {
		t.Fatalf("read back generation: %v", err)
	}
	if len(gen.Proposals) != 1 {
		t.Fatalf("expected one persisted proposal, got %d", len(gen.Proposals))
	}
	pRow := gen.Proposals[0]
	if len(pRow.Targets) != 1 || pRow.Targets[0].InvariantID != cand.ID ||
		pRow.Targets[0].Verdict != "violates" || !pRow.ViolatesAnyTarget {
		t.Fatalf("persisted violation verdict must be a verified break of %s: %+v", cand.ID, pRow.Targets)
	}

	// Vocabulary consistency (the proposal-side admission check): every
	// RESOLVED claim in the persisted proposal signature must agree with the
	// pinned vocabulary's own resolution of its surface label — the control
	// does not trust a provider-supplied `resolved` status or invented alias.
	occ, err := repo.ListGenerationOccurrenceContents(ctx, gen.ID)
	if err != nil {
		t.Fatalf("occurrence contents: %v", err)
	}
	oc := occ[pRow.ID]
	if oc.SignatureJSON == "" {
		t.Fatal("proposal must have persisted signature content")
	}
	var persisted canon.MechanismSignature
	if err := json.Unmarshal([]byte(oc.SignatureJSON), &persisted); err != nil {
		t.Fatalf("unmarshal persisted proposal signature: %v", err)
	}
	vocab := canon.MechanismV1()
	for _, claim := range persisted.Preserves {
		if claim.State != domain.ResolutionResolved {
			continue
		}
		res := vocab.Resolve(claim.FieldKind, claim.SurfaceLabel, false)
		if res.State != domain.ResolutionResolved || res.CanonicalID != claim.CanonicalID {
			t.Fatalf("provider-resolved claim %q -> %q disagrees with the pinned vocabulary (%q, %q)",
				claim.SurfaceLabel, claim.CanonicalID, res.State, res.CanonicalID)
		}
	}

	// The withheld target the experiment ACTUALLY scored against carries the
	// declared completeness through GetSignature — the exact reader the
	// experiment uses for its frozen target manifest.
	if len(exp.Targets) != 1 {
		t.Fatalf("expected one frozen target, got %d", len(exp.Targets))
	}
	targetRec, err := repo.GetSignature(ctx, exp.Targets[0].SignatureID)
	if err != nil {
		t.Fatalf("get target signature: %v", err)
	}
	targetSig := signatureFromRecord(targetRec)
	if got := targetSig.FieldCompleteness(domain.FieldPreserves); got != domain.CompletenessComplete {
		t.Fatalf("frozen target signature must carry declared preserves completeness, got %q", got)
	}
}

// TestIntegrationPersistedInputCompletenessDeclarationMatters is the same-path
// completeness mutation: strip ONLY the payload's field_completeness
// declarations and the persisted contrast verdict degrades from a decisive
// violation to unknown (ContrastViolatingNum 1 -> 0) — proving the declaration
// flows payload -> storage -> signature -> predicate evaluation, and its
// absence keeps the conservative default rather than manufacturing a negative.
func TestIntegrationPersistedInputCompletenessDeclarationMatters(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.normalizers = map[string]provider.Normalizer{provider.FixtureProviderName: provider.NewFixtureNormalizer()}

	doc := writeControlDoc(t, t.TempDir(), "control-train-undeclared.md", controlTrainResult(false))
	trainInit, err := app.InitProblem(ctx, InitProblemInput{DBPath: dbPath, Statement: "persisted-input control (undeclared)", Slug: "pic-undeclared"})
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	ingestNormalizeSign(t, ctx, app, dbPath, trainInit.ProblemID, doc)
	if _, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: trainInit.ProblemID}); err != nil {
		t.Fatalf("cluster: %v", err)
	}
	if _, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: trainInit.ProblemID}); err != nil {
		t.Fatalf("failure-space: %v", err)
	}
	mined, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: trainInit.ProblemID, MinSupport: 2})
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if mined.Revision.CandidateCount != 1 {
		t.Fatalf("support does not depend on completeness; expected 1 candidate, got %d", mined.Revision.CandidateCount)
	}
	if got := mined.Revision.Candidates[0].ContrastViolatingNum; got != 0 {
		t.Fatalf("without a declaration, contrast absence must stay UNKNOWN (not a violation): got %d", got)
	}
}

// cannedNormalizer simulates a MODEL normalizer: it returns an authored result
// regardless of content. It is NOT the deterministic embedded-payload parser,
// so its completeness declarations must never be code-accepted.
type cannedNormalizer struct{ result normalize.Result }

func (c cannedNormalizer) Normalize(_ context.Context, _ normalize.Request) (provider.NormalizeResponse, error) {
	return provider.NormalizeResponse{
		Result:          c.result,
		Metadata:        provider.Metadata{ProviderName: "model-sim", ProviderVersion: "v1", ModelName: "simulated-model", SchemaVersion: normalize.SchemaVersion},
		RequestPayload:  "req",
		ResponsePayload: "resp",
	}, nil
}

// TestIntegrationCompletenessAuthorityIsCodeDecided is the finding-1
// regression (review of 87759d9): a provider-supplied declaration with an
// arbitrary nonempty basis must NOT acquire the evaluation authority of a
// validated closed-world parser result. Two inadmissible paths:
//
//	(a) an UNTRUSTED (model-simulating) normalizer declares declared_payload
//	    scope with a confident basis -> persisted DECLARED_ONLY, signature
//	    stays unobserved;
//	(b) the trusted parser carries a mechanism_exhaustive-scoped declaration
//	    (an extraction judgment) -> DECLARED_ONLY as well.
//
// In both cases the declaration is retained verbatim as an auditable claim —
// never silently dropped, never granted absence-based verification authority.
func TestIntegrationCompletenessAuthorityIsCodeDecided(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

	assertDeclaredOnly := func(t *testing.T, app *App, dbPath, problemID string) {
		t.Helper()
		repo := openTestStore(t, ctx, dbPath)
		defer repo.Close()
		approaches, err := repo.ListApproaches(ctx, problemID)
		if err != nil || len(approaches) == 0 {
			t.Fatalf("list approaches: %v (%d)", err, len(approaches))
		}
		for _, item := range approaches {
			detail, err := repo.GetApproachDetail(ctx, item.Approach.ID)
			if err != nil {
				t.Fatalf("detail: %v", err)
			}
			if len(detail.FieldCompleteness) == 0 {
				t.Fatalf("the declaration must be RETAINED as an auditable claim, got none for %s", item.Approach.LogicalIdentity)
			}
			for _, fc := range detail.FieldCompleteness {
				if fc.Admission != domain.CompletenessDeclaredOnly {
					t.Fatalf("admission must be declared_only, got %q (%s)", fc.Admission, fc.AdmissionBasis)
				}
			}
			// The signature must keep the conservative default: no evaluation authority.
			rec, _, err := app.buildAndPersistSignature(ctx, repo, detail.Mechanism.ID, "", "")
			if err != nil {
				t.Fatalf("signature: %v", err)
			}
			sig := signatureFromRecord(rec)
			if got := sig.FieldCompleteness(domain.FieldPreserves); got != domain.CompletenessUnobserved {
				t.Fatalf("a declared_only claim must NOT upgrade the signature, got %q", got)
			}
		}
	}

	// (a) Untrusted normalizer, confident arbitrary basis, declared_payload scope.
	t.Run("untrusted_provider_claim", func(t *testing.T) {
		app, dbPath := newRealStoreApp(t, now)
		payload := controlTrainResult(true)
		for i := range payload.Approaches {
			payload.Approaches[i].Mechanism.CompletenessBasis = "The model believes extraction is exhaustive."
		}
		app.normalizers = map[string]provider.Normalizer{
			provider.FixtureProviderName: provider.NewFixtureNormalizer(),
			"model-sim":                  cannedNormalizer{result: payload},
		}
		doc := writeControlDoc(t, t.TempDir(), "untrusted.md", payload)
		init, err := app.InitProblem(ctx, InitProblemInput{DBPath: dbPath, Statement: "authority (untrusted)", Slug: "pic-untrusted"})
		if err != nil {
			t.Fatalf("init: %v", err)
		}
		if _, err := app.IngestSources(ctx, IngestInput{DBPath: dbPath, ProblemID: init.ProblemID, Paths: []string{doc}}); err != nil {
			t.Fatalf("ingest: %v", err)
		}
		if _, err := app.Normalize(ctx, NormalizeInput{DBPath: dbPath, ProblemID: init.ProblemID, All: true, Provider: "model-sim"}); err != nil {
			t.Fatalf("normalize: %v", err)
		}
		assertDeclaredOnly(t, app, dbPath, init.ProblemID)
	})

	// (b) Trusted parser, but a mechanism-exhaustive scope (extraction judgment).
	t.Run("mechanism_exhaustive_scope", func(t *testing.T) {
		app, dbPath := newRealStoreApp(t, now)
		app.normalizers = map[string]provider.Normalizer{provider.FixtureProviderName: provider.NewFixtureNormalizer()}
		payload := controlTrainResult(true)
		for i := range payload.Approaches {
			payload.Approaches[i].Mechanism.CompletenessScope = string(domain.ScopeMechanismExhaustive)
		}
		doc := writeControlDoc(t, t.TempDir(), "exhaustive-scope.md", payload)
		init, err := app.InitProblem(ctx, InitProblemInput{DBPath: dbPath, Statement: "authority (scope)", Slug: "pic-exhaustive"})
		if err != nil {
			t.Fatalf("init: %v", err)
		}
		if _, err := app.IngestSources(ctx, IngestInput{DBPath: dbPath, ProblemID: init.ProblemID, Paths: []string{doc}}); err != nil {
			t.Fatalf("ingest: %v", err)
		}
		if _, err := app.Normalize(ctx, NormalizeInput{DBPath: dbPath, ProblemID: init.ProblemID, All: true}); err != nil {
			t.Fatalf("normalize: %v", err)
		}
		assertDeclaredOnly(t, app, dbPath, init.ProblemID)
	})
}
