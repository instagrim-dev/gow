package pipeline

// --- problem / run views ---

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	OK      bool        `json:"ok"`
	Command string      `json:"command"`
	Error   ErrorDetail `json:"error"`
}

type InitResponse struct {
	OK        bool   `json:"ok"`
	Command   string `json:"command"`
	ProblemID string `json:"problem_id"`
	RunID     string `json:"run_id"`
	Store     string `json:"store"`
	Created   bool   `json:"created"`
	Problem   string `json:"problem"`
}

type ProblemView struct {
	ID             string `json:"id"`
	Slug           string `json:"slug"`
	Statement      string `json:"statement"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
	CreatedByRunID string `json:"created_by_run_id"`
}

type ProblemShowResponse struct {
	OK      bool        `json:"ok"`
	Command string      `json:"command"`
	Store   string      `json:"store"`
	Problem ProblemView `json:"problem"`
}

type ProblemListResponse struct {
	OK       bool          `json:"ok"`
	Command  string        `json:"command"`
	Store    string        `json:"store"`
	Problems []ProblemView `json:"problems"`
}

type RunView struct {
	ID           string  `json:"id"`
	ProblemID    string  `json:"problem_id"`
	ParentRunID  *string `json:"parent_run_id,omitempty"`
	Operation    string  `json:"operation"`
	Status       string  `json:"status"`
	InputRef     string  `json:"input_ref"`
	ToolName     string  `json:"tool_name"`
	ToolVersion  string  `json:"tool_version"`
	StartedAt    string  `json:"started_at"`
	CompletedAt  string  `json:"completed_at"`
	ErrorSummary *string `json:"error_summary,omitempty"`
}

type RunShowResponse struct {
	OK      bool    `json:"ok"`
	Command string  `json:"command"`
	Store   string  `json:"store"`
	Run     RunView `json:"run"`
}

// --- source / ingest views ---

type IngestItemResult struct {
	Input         string `json:"input"`
	SourceID      string `json:"source_id,omitempty"`
	SnapshotID    string `json:"snapshot_id,omitempty"`
	Status        string `json:"status"`
	SHA256        string `json:"sha256,omitempty"`
	MediaType     string `json:"media_type,omitempty"`
	Bytes         int64  `json:"bytes,omitempty"`
	CreatedSource bool   `json:"created_source,omitempty"`
	Error         string `json:"error,omitempty"`
}

type IngestSummary struct {
	Total     int `json:"total"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

type IngestResponse struct {
	OK        bool               `json:"ok"`
	Command   string             `json:"command"`
	Store     string             `json:"store"`
	ProblemID string             `json:"problem_id"`
	RunID     string             `json:"run_id"`
	Results   []IngestItemResult `json:"results"`
	Summary   IngestSummary      `json:"summary"`
}

type SourceListView struct {
	ID               string  `json:"id"`
	ProblemID        string  `json:"problem_id"`
	Kind             string  `json:"kind"`
	LogicalName      string  `json:"logical_name"`
	Origin           string  `json:"origin"`
	CreatedAt        string  `json:"created_at"`
	SnapshotCount    int     `json:"snapshot_count"`
	LatestSnapshotID *string `json:"latest_snapshot_id,omitempty"`
}

type SourceListResponse struct {
	OK        bool             `json:"ok"`
	Command   string           `json:"command"`
	Store     string           `json:"store"`
	ProblemID string           `json:"problem_id"`
	Sources   []SourceListView `json:"sources"`
}

type SourceShowResponse struct {
	OK        bool                 `json:"ok"`
	Command   string               `json:"command"`
	Store     string               `json:"store"`
	Source    SourceListView       `json:"source"`
	Snapshots []SourceSnapshotView `json:"snapshots"`
}

type SourceSnapshotView struct {
	ID                   string  `json:"id"`
	SourceID             string  `json:"source_id"`
	SHA256               string  `json:"sha256"`
	ByteLength           int64   `json:"byte_length"`
	MediaType            string  `json:"media_type"`
	ObjectPath           string  `json:"object_path"`
	ObservedAt           string  `json:"observed_at"`
	IngestRunID          string  `json:"ingest_run_id"`
	SupersedesSnapshotID *string `json:"supersedes_snapshot_id,omitempty"`
}

type SourceSnapshotShowResponse struct {
	OK                 bool               `json:"ok"`
	Command            string             `json:"command"`
	Store              string             `json:"store"`
	Source             SourceListView     `json:"source"`
	Snapshot           SourceSnapshotView `json:"snapshot"`
	ObjectAbsolutePath string             `json:"object_absolute_path"`
}

type SourceSnapshotVerifyResponse struct {
	OK         bool   `json:"ok"`
	Command    string `json:"command"`
	Store      string `json:"store"`
	SnapshotID string `json:"snapshot_id"`
	Status     string `json:"status"`
	SHA256     string `json:"sha256,omitempty"`
	Bytes      int64  `json:"bytes,omitempty"`
}

// SourceLineageDiffSide names one endpoint of a lineage-diff comparison.
// Store paths are echoed so operators can audit which two datasets were
// compared (the diagnostic is only as strong as the two sides).
type SourceLineageDiffSide struct {
	Store          string `json:"store"`
	ProblemID      string `json:"problem_id"`
	SnapshotCount  int    `json:"snapshot_count"`
	DistinctSHA256 int    `json:"distinct_sha256"`
}

// SourceLineageOverlapEntry names one shared SHA-256 with the operator-legible
// logical_name each side used for it. Same-bytes-different-name is a
// legitimate outcome (an operator may relabel an attested source), so both
// names are surfaced verbatim.
type SourceLineageOverlapEntry struct {
	SHA256           string `json:"sha256"`
	LeftLogicalName  string `json:"left_logical_name"`
	RightLogicalName string `json:"right_logical_name"`
}

// SourceLineageDiffResponse is the JSON shape of the diagnostic. It is
// deliberately not a lineage record persisted to the store: this is
// operator-consumed comparison output, not an admission-scoped artifact.
type SourceLineageDiffResponse struct {
	OK             bool                  `json:"ok"`
	Command        string                `json:"command"`
	Left           SourceLineageDiffSide `json:"left"`
	Right          SourceLineageDiffSide `json:"right"`
	SharedCount    int                   `json:"shared_count"`
	LeftOnlyCount  int                   `json:"left_only_count"`
	RightOnlyCount int                   `json:"right_only_count"`
	// Verdict summarizes the overlap: "disjoint" (0 shared),
	// "partial_overlap" (>0 shared but not identical), "identical" (both
	// sides have exactly the same non-empty SHA-256 set), "subset_left"
	// (left is a non-empty proper subset of right), "subset_right"
	// (right is a non-empty proper subset of left), or "empty" (at least
	// one side has no snapshots).
	Verdict string `json:"verdict"`
	// SharedSample carries up to a bounded number of overlapping SHA-256
	// values with each side's logical_name for operator inspection. It is
	// a sample, not an exhaustive list: the counts above are authoritative
	// for the overlap size.
	SharedSample []SourceLineageOverlapEntry `json:"shared_sample,omitempty"`
}

// SourceLineageDiffInput selects the two sides of the comparison. When
// AgainstDBPath is empty, the diff runs against another problem in the same
// database. Both AgainstProblemID and (DBPath, ProblemID) are required; the
// diff is deliberately problem-scoped on both sides so operators cannot
// silently compare a specific corpus against an unrelated aggregate.
type SourceLineageDiffInput struct {
	DBPath            string
	ProblemID         string
	AgainstDBPath     string
	AgainstProblemID  string
	JSONOutput        bool
	SharedSampleLimit int
}

type NormalizeApproachResult struct {
	ApproachID      string `json:"approach_id"`
	RevisionID      string `json:"revision_id"`
	MechanismID     string `json:"mechanism_id"`
	CreatedApproach bool   `json:"created_approach"`
}

type NormalizeResult struct {
	SnapshotID string                    `json:"source_snapshot_id"`
	Status     string                    `json:"status"`
	Reason     string                    `json:"reason,omitempty"`
	Message    string                    `json:"message,omitempty"`
	RevisionID string                    `json:"revision_id,omitempty"`
	Approaches []NormalizeApproachResult `json:"approaches,omitempty"`
	Warnings   []string                  `json:"warnings,omitempty"`
}

type NormalizeResponse struct {
	OK        bool              `json:"ok"`
	Command   string            `json:"command"`
	Store     string            `json:"store"`
	ProblemID string            `json:"problem_id"`
	RunID     string            `json:"run_id"`
	Provider  string            `json:"provider"`
	Schema    string            `json:"schema_version"`
	Results   []NormalizeResult `json:"results"`
}

// --- approach / mechanism views ---

type FieldSupportView struct {
	FieldPath   string `json:"field_path"`
	SupportKind string `json:"support_kind"`
	Locator     string `json:"locator,omitempty"`
	Confidence  string `json:"confidence,omitempty"`
}

type MechanismView struct {
	ID               string   `json:"id"`
	ApproachRevision string   `json:"approach_revision_id"`
	Representations  []string `json:"representations,omitempty"`
	Assumptions      []string `json:"assumptions,omitempty"`
	Operators        []string `json:"operators,omitempty"`
	Preserves        []string `json:"preserves,omitempty"`
	Breaks           []string `json:"breaks,omitempty"`
	AuxiliaryObjects []string `json:"auxiliary_objects,omitempty"`
	Locality         string   `json:"locality"`
	ConstructionMode string   `json:"construction_mode"`
	UncertaintyMode  string   `json:"uncertainty_mode"`
	Notes            string   `json:"notes,omitempty"`
}

type OutcomeView struct {
	Class              string   `json:"class"`
	BoundaryStatement  string   `json:"boundary_statement,omitempty"`
	BoundaryConditions []string `json:"boundary_conditions,omitempty"`
	Notes              string   `json:"notes,omitempty"`
}

type ApproachRevisionView struct {
	ID                      string  `json:"id"`
	ApproachID              string  `json:"approach_id"`
	NormalizationRevisionID string  `json:"normalization_revision_id"`
	Label                   string  `json:"label"`
	Description             string  `json:"description,omitempty"`
	SupersedesRevisionID    *string `json:"supersedes_revision_id,omitempty"`
	CreatedAt               string  `json:"created_at"`
}

type ApproachListView struct {
	ID              string `json:"id"`
	ProblemID       string `json:"problem_id"`
	LogicalIdentity string `json:"logical_identity"`
	Label           string `json:"label"`
	OutcomeClass    string `json:"outcome_class,omitempty"`
	SnapshotID      string `json:"source_snapshot_id,omitempty"`
	RevisionCount   int    `json:"revision_count"`
	LatestRevision  string `json:"latest_revision_id,omitempty"`
}

type ApproachListResponse struct {
	OK         bool               `json:"ok"`
	Command    string             `json:"command"`
	Store      string             `json:"store"`
	ProblemID  string             `json:"problem_id"`
	Approaches []ApproachListView `json:"approaches"`
}

type ProviderInvocationView struct {
	ID              string `json:"id"`
	Role            string `json:"role"`
	ProviderName    string `json:"provider_name"`
	ProviderVersion string `json:"provider_version,omitempty"`
	ModelName       string `json:"model_name,omitempty"`
	SchemaVersion   string `json:"schema_version"`
	RequestHash     string `json:"request_hash"`
}

type ApproachShowResponse struct {
	OK                    bool                   `json:"ok"`
	Command               string                 `json:"command"`
	Store                 string                 `json:"store"`
	ApproachID            string                 `json:"approach_id"`
	ProblemID             string                 `json:"problem_id"`
	LogicalIdentity       string                 `json:"logical_identity"`
	Revision              ApproachRevisionView   `json:"revision"`
	NormalizationRevision string                 `json:"normalization_revision_id"`
	SnapshotID            string                 `json:"source_snapshot_id"`
	RunID                 string                 `json:"run_id"`
	Provider              ProviderInvocationView `json:"provider"`
	Mechanism             MechanismView          `json:"mechanism"`
	Outcome               OutcomeView            `json:"outcome"`
	Support               []FieldSupportView     `json:"support"`
	RevisionCount         int                    `json:"revision_count"`
}

type ApproachRevisionsResponse struct {
	OK         bool                   `json:"ok"`
	Command    string                 `json:"command"`
	Store      string                 `json:"store"`
	ApproachID string                 `json:"approach_id"`
	ProblemID  string                 `json:"problem_id"`
	Revisions  []ApproachRevisionView `json:"revisions"`
}

type MechanismShowResponse struct {
	OK          bool          `json:"ok"`
	Command     string        `json:"command"`
	Store       string        `json:"store"`
	MechanismID string        `json:"mechanism_id"`
	ApproachID  string        `json:"approach_id"`
	RevisionID  string        `json:"approach_revision_id"`
	Mechanism   MechanismView `json:"mechanism"`
	Outcome     OutcomeView   `json:"outcome"`
}

// --- canonicalization / signature / comparison views (#9) ---

type VocabularyView struct {
	Version   string `json:"version"`
	CreatedAt string `json:"created_at"`
	Notes     string `json:"notes,omitempty"`
}

type TermView struct {
	VocabularyVersion string   `json:"vocabulary_version"`
	CanonicalID       string   `json:"canonical_id"`
	FieldKind         string   `json:"field_kind"`
	Description       string   `json:"description,omitempty"`
	ParentCanonicalID string   `json:"parent_canonical_id,omitempty"`
	Aliases           []string `json:"aliases,omitempty"`
}

type VocabularyListResponse struct {
	OK           bool             `json:"ok"`
	Command      string           `json:"command"`
	Store        string           `json:"store"`
	Version      string           `json:"version,omitempty"`
	Vocabularies []VocabularyView `json:"vocabularies,omitempty"`
	Terms        []TermView       `json:"terms,omitempty"`
}

type VocabularyShowResponse struct {
	OK          bool       `json:"ok"`
	Command     string     `json:"command"`
	Store       string     `json:"store"`
	CanonicalID string     `json:"canonical_id"`
	Found       bool       `json:"found"`
	Terms       []TermView `json:"terms"`
}

type VocabularyResolveResponse struct {
	OK          bool     `json:"ok"`
	Command     string   `json:"command"`
	Store       string   `json:"store"`
	Version     string   `json:"vocabulary_version"`
	Field       string   `json:"field"`
	SurfaceKey  string   `json:"surface_key"`
	State       string   `json:"state"`
	CanonicalID string   `json:"canonical_id,omitempty"`
	Candidates  []string `json:"candidates,omitempty"`
}

type SignatureFieldClaimView struct {
	FieldKind          string `json:"field_kind"`
	SurfaceLabel       string `json:"surface_label"`
	ResolutionState    string `json:"resolution_state"`
	CanonicalID        string `json:"canonical_id,omitempty"`
	ClaimStatus        string `json:"claim_status"`
	SupportSnapshotID  string `json:"support_snapshot_id,omitempty"`
	SupportLocator     string `json:"support_locator,omitempty"`
	Confidence         string `json:"confidence,omitempty"`
	ClassifierContract string `json:"classifier_contract,omitempty"`
}

type SignatureBoundaryView struct {
	SurfaceLabel    string `json:"surface_label"`
	ResolutionState string `json:"resolution_state"`
	CanonicalID     string `json:"canonical_id,omitempty"`
	Relation        string `json:"relation,omitempty"`
	ClaimStatus     string `json:"claim_status"`
}

type SignatureView struct {
	ID                string                    `json:"id"`
	MechanismID       string                    `json:"mechanism_id"`
	SchemaVersion     string                    `json:"schema_version"`
	VocabularyVersion string                    `json:"vocabulary_version"`
	Fingerprint       string                    `json:"fingerprint"`
	OutcomeClass      string                    `json:"outcome_class"`
	OutcomeStatus     string                    `json:"outcome_status"`
	Posture           map[string]string         `json:"posture"`
	PostureStatus     map[string]string         `json:"posture_status"`
	FieldClaims       []SignatureFieldClaimView `json:"field_claims"`
	Boundaries        []SignatureBoundaryView   `json:"boundaries,omitempty"`
	CreatedAt         string                    `json:"created_at"`
}

type SignatureResponse struct {
	OK        bool          `json:"ok"`
	Command   string        `json:"command"`
	Store     string        `json:"store"`
	Status    string        `json:"status"`
	Signature SignatureView `json:"signature"`
}

type ComparisonFieldView struct {
	FieldKind    string  `json:"field_kind"`
	OverlapCount int     `json:"overlap_count"`
	UnionCount   int     `json:"union_count"`
	Jaccard      float64 `json:"jaccard"`
	Ordinal      string  `json:"ordinal"`
	Incomparable bool    `json:"incomparable"`
}

type PostureComparisonView struct {
	LocalityEqual     bool `json:"locality_equal"`
	ConstructionEqual bool `json:"construction_equal"`
	UncertaintyEqual  bool `json:"uncertainty_equal"`
}

type ComparisonView struct {
	WeightsVersion  string                `json:"weights_version"`
	ClassifyVersion string                `json:"classify_version"`
	ProfileHash     string                `json:"profile_hash"`
	Classification  string                `json:"classification"`
	OutcomeEqual    bool                  `json:"outcome_equal"`
	Posture         PostureComparisonView `json:"posture"`
	Fields          []ComparisonFieldView `json:"fields"`
}

type CompareResponse struct {
	OK              bool           `json:"ok"`
	Command         string         `json:"command"`
	Store           string         `json:"store"`
	SignatureAID    string         `json:"signature_a_id"`
	SignatureBID    string         `json:"signature_b_id"`
	FingerprintA    string         `json:"fingerprint_a"`
	FingerprintB    string         `json:"fingerprint_b"`
	ComparisonRunID string         `json:"comparison_run_id,omitempty"`
	Comparison      ComparisonView `json:"comparison"`
}

type SeedFixtureResponse struct {
	OK           bool     `json:"ok"`
	Command      string   `json:"command"`
	Store        string   `json:"store"`
	ProblemID    string   `json:"problem_id"`
	RunID        string   `json:"run_id"`
	SnapshotID   string   `json:"source_snapshot_id"`
	RevisionID   string   `json:"normalization_revision_id"`
	ApproachIDs  []string `json:"approach_ids"`
	MechanismIDs []string `json:"mechanism_ids"`
}

// --- cluster / failure-space views ---

// ClusterMemberView is one signature in a cluster.
type ClusterMemberView struct {
	SignatureID string `json:"signature_id"`
	MechanismID string `json:"mechanism_id"`
	Redundant   bool   `json:"redundant"`
	// AdmittedObservationKind labels a member that entered the population via
	// the evidence-admission ledger with its epistemic kind (e.g.
	// model-judged-failure for an attested model verdict). Empty for members
	// that entered through ordinary source normalization (2026-09-12 review
	// F4: an admitted observation must not shed its kind inside the atlas).
	AdmittedObservationKind string `json:"admitted_observation_kind,omitempty"`
}

// ClusterView is one mechanism family.
type ClusterView struct {
	ID                        string              `json:"id"`
	Fingerprint               string              `json:"fingerprint"`
	RepresentativeSignatureID string              `json:"representative_signature_id"`
	MemberCount               int                 `json:"member_count"`
	Isolate                   bool                `json:"isolate"`
	OutcomeClass              string              `json:"outcome_class"`
	OutcomeMixed              bool                `json:"outcome_mixed"`
	IntraVariation            string              `json:"intra_variation"`
	Members                   []ClusterMemberView `json:"members"`
}

// ClusterDistanceView is one representative-vs-representative distance.
type ClusterDistanceView struct {
	ClusterAID     string `json:"cluster_a_id"`
	ClusterBID     string `json:"cluster_b_id"`
	Classification string `json:"classification"`
}

// ClusterCoverageAxisView is one coverage axis.
type ClusterCoverageAxisView struct {
	Axis               string `json:"axis"`
	DistinctValueCount int    `json:"distinct_value_count"`
	UnderSampled       bool   `json:"under_sampled"`
}

// DiscriminationLossView is one abstraction-loss finding.
type DiscriminationLossView struct {
	MechanismAID string `json:"mechanism_a_id"`
	MechanismBID string `json:"mechanism_b_id"`
	OutcomeA     string `json:"outcome_a"`
	OutcomeB     string `json:"outcome_b"`
}

// ClusterRunView is the full clustering pass.
type ClusterRunView struct {
	ID                 string                    `json:"id"`
	ProblemID          string                    `json:"problem_id"`
	RunID              string                    `json:"run_id"`
	SchemaVersion      string                    `json:"schema_version"`
	VocabularyVersion  string                    `json:"vocabulary_version"`
	ProfileVersion     string                    `json:"profile_version"`
	ClusterAlgoVersion string                    `json:"cluster_algo_version"`
	ThresholdsHash     string                    `json:"thresholds_hash"`
	InputSetHash       string                    `json:"input_set_hash"`
	SignatureCount     int                       `json:"signature_count"`
	FamilyCount        int                       `json:"family_count"`
	Status             string                    `json:"status"`
	Clusters           []ClusterView             `json:"clusters"`
	Distances          []ClusterDistanceView     `json:"distances"`
	CoverageAxes       []ClusterCoverageAxisView `json:"coverage_axes"`
	DiscriminationLoss []DiscriminationLossView  `json:"discrimination_loss,omitempty"`
}

// ClusterBuildResponse is returned by `newf cluster build`.
type ClusterBuildResponse struct {
	OK         bool           `json:"ok"`
	Command    string         `json:"command"`
	Store      string         `json:"store"`
	Created    bool           `json:"created"`
	ClusterRun ClusterRunView `json:"cluster_run"`
}

// ClusterShowResponse is returned by `newf cluster show`.
type ClusterShowResponse struct {
	OK         bool           `json:"ok"`
	Command    string         `json:"command"`
	Store      string         `json:"store"`
	ClusterRun ClusterRunView `json:"cluster_run"`
}

// ClusterRunSummaryView is a cluster-run header for list output.
type ClusterRunSummaryView struct {
	ID             string `json:"id"`
	ProfileVersion string `json:"profile_version"`
	InputSetHash   string `json:"input_set_hash"`
	SignatureCount int    `json:"signature_count"`
	FamilyCount    int    `json:"family_count"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
}

// ClusterListResponse is returned by `newf cluster list`.
type ClusterListResponse struct {
	OK          bool                    `json:"ok"`
	Command     string                  `json:"command"`
	Store       string                  `json:"store"`
	ClusterRuns []ClusterRunSummaryView `json:"cluster_runs"`
}

// FailureSpaceOutcomeView is the family count for one outcome class.
type FailureSpaceOutcomeView struct {
	OutcomeClass string `json:"outcome_class"`
	FamilyCount  int    `json:"family_count"`
}

// FailureSpaceAxisView is one coverage axis of a failure space.
type FailureSpaceAxisView struct {
	AxisKind           string `json:"axis_kind"`
	DistinctValueCount int    `json:"distinct_value_count"`
	UnderSampled       bool   `json:"under_sampled"`
}

// FailureSpaceView is a materialized failure-space revision.
type FailureSpaceView struct {
	ID                   string                    `json:"id"`
	ProblemID            string                    `json:"problem_id"`
	ClusterRunID         string                    `json:"cluster_run_id"`
	Revision             int                       `json:"revision"`
	DistinctFamilyCount  int                       `json:"distinct_family_count"`
	RedundantMemberCount int                       `json:"redundant_member_count"`
	Outcomes             []FailureSpaceOutcomeView `json:"outcomes"`
	Axes                 []FailureSpaceAxisView    `json:"axes"`
}

// FailureSpaceBuildResponse is returned by `newf failure-space build`.
type FailureSpaceBuildResponse struct {
	OK           bool             `json:"ok"`
	Command      string           `json:"command"`
	Store        string           `json:"store"`
	Created      bool             `json:"created"`
	FailureSpace FailureSpaceView `json:"failure_space"`
}

// FailureSpaceShowResponse is returned by `newf failure-space show`.
type FailureSpaceShowResponse struct {
	OK           bool             `json:"ok"`
	Command      string           `json:"command"`
	Store        string           `json:"store"`
	FailureSpace FailureSpaceView `json:"failure_space"`
}

// FailureSpaceCoverageResponse is returned by `newf failure-space coverage`.
type FailureSpaceCoverageResponse struct {
	OK           bool                   `json:"ok"`
	Command      string                 `json:"command"`
	Store        string                 `json:"store"`
	FailureSpace string                 `json:"failure_space_id"`
	Axes         []FailureSpaceAxisView `json:"axes"`
}

// InvariantFamilyEvaluationView is a candidate's per-family verdict + the
// epistemic composition of the matched claims (support vs. contrast side).
type InvariantFamilyEvaluationView struct {
	ClusterID     string `json:"cluster_id"`
	OutcomeClass  string `json:"outcome_class"`
	Role          string `json:"role"`
	Verdict       string `json:"verdict"`
	ExplicitCount int    `json:"explicit_count"`
	InferredCount int    `json:"inferred_count"`
	OtherCount    int    `json:"other_count"`
}

// InvariantCounterexampleView is an eligible failure family that violated the
// predicate.
type InvariantCounterexampleView struct {
	ClusterID string `json:"cluster_id"`
	Reason    string `json:"reason"`
}

// CandidateInvariantView is one proposed candidate invariant: the typed
// predicate (semantic identity), its human render, code-computed support and
// contrast, retained epistemic composition, and the model-hypothesis flag. The
// state is always `proposed` in this slice.
type CandidateInvariantView struct {
	ID                           string                          `json:"id"`
	PredicateFingerprint         string                          `json:"predicate_fingerprint"`
	Predicate                    string                          `json:"predicate"`
	Statement                    string                          `json:"statement"`
	AbstractionLevel             string                          `json:"abstraction_level"`
	State                        string                          `json:"state"`
	AssociationStatus            string                          `json:"association_status"`
	ObstructionIsModelHypothesis bool                            `json:"obstruction_is_model_hypothesis"`
	DistinctFamilySupport        int                             `json:"distinct_family_support"`
	FailureCoverageNum           int                             `json:"failure_coverage_num"`
	FailureCoverageDen           int                             `json:"failure_coverage_den"`
	ContrastViolatingNum         int                             `json:"contrast_violating_num"`
	ContrastEligibleDen          int                             `json:"contrast_eligible_den"`
	SupportExplicitCount         int                             `json:"support_explicit_count"`
	SupportInferredCount         int                             `json:"support_inferred_count"`
	SupportOtherCount            int                             `json:"support_other_count"`
	FamilyEvaluations            []InvariantFamilyEvaluationView `json:"family_evaluations"`
	Counterexamples              []InvariantCounterexampleView   `json:"counterexamples"`
}

// InvariantRevisionView is a full mining pass.
type InvariantRevisionView struct {
	ID              string                   `json:"id"`
	ProblemID       string                   `json:"problem_id"`
	FailureSpaceID  string                   `json:"failure_space_id"`
	ClusterRunID    string                   `json:"cluster_run_id"`
	RunID           string                   `json:"run_id"`
	MinerVersion    string                   `json:"miner_version"`
	PredicateSchema string                   `json:"predicate_schema"`
	MinSupport      int                      `json:"min_support"`
	Revision        int                      `json:"revision"`
	CandidateCount  int                      `json:"candidate_count"`
	CreatedAt       string                   `json:"created_at"`
	Candidates      []CandidateInvariantView `json:"candidates"`
}

// InvariantRevisionSummaryView is a compact list row.
type InvariantRevisionSummaryView struct {
	ID             string `json:"id"`
	FailureSpaceID string `json:"failure_space_id"`
	MinerVersion   string `json:"miner_version"`
	MinSupport     int    `json:"min_support"`
	Revision       int    `json:"revision"`
	CandidateCount int    `json:"candidate_count"`
	CreatedAt      string `json:"created_at"`
}

// InvariantMineResponse is returned by `newf invariants mine`.
type InvariantMineResponse struct {
	OK       bool                  `json:"ok"`
	Command  string                `json:"command"`
	Store    string                `json:"store"`
	Created  bool                  `json:"created"`
	Revision InvariantRevisionView `json:"revision"`
}

// InvariantListResponse is returned by `newf invariant list`.
type InvariantListResponse struct {
	OK        bool                           `json:"ok"`
	Command   string                         `json:"command"`
	Store     string                         `json:"store"`
	Revisions []InvariantRevisionSummaryView `json:"revisions"`
}

// InvariantShowResponse is returned by `newf invariant show`.
type InvariantShowResponse struct {
	OK       bool                  `json:"ok"`
	Command  string                `json:"command"`
	Store    string                `json:"store"`
	Revision InvariantRevisionView `json:"revision"`
}

// FrontierTargetView is one targeted surviving invariant plus the code-verified
// per-target violation verdict.
type FrontierTargetView struct {
	InvariantID string `json:"invariant_id"`
	Verdict     string `json:"verdict"`
	Violated    bool   `json:"violated"`
}

// FrontierNearestView is one code-computed nearest failure family.
type FrontierNearestView struct {
	ClusterID      string `json:"cluster_id"`
	Classification string `json:"classification"`
	Proximity      string `json:"proximity_ordinal"`
}

// FrontierProposalView is one ranked frontier proposal: the required directed-
// generation prose, code-computed mechanistic distance + violation checks +
// nearest families, and the ordinal scores. Result is empty until M5.2
// evaluation populates it.
type FrontierProposalView struct {
	ID                        string `json:"id"`
	ProposalHash              string `json:"proposal_hash"`
	StructuralViolationClaim  string `json:"structural_violation_claim"`
	NoveltyArgument           string `json:"novelty_argument"`
	CheapestFalsificationPath string `json:"cheapest_falsification_path"`
	MechanisticDistance       string `json:"mechanistic_distance_ordinal"`
	ExpectedInformationGain   string `json:"expected_information_gain_ordinal"`
	EvaluationCost            string `json:"evaluation_cost_ordinal"`
	ViolatesAnyTarget         bool   `json:"violates_any_target"`
	Rank                      int    `json:"rank"`
	Result                    string `json:"result,omitempty"`
	// Result provenance: the ledger evaluation whose verdict Result carries.
	// A reader never sees an occurrence outcome without its epistemic strength.
	ResultEvaluationID         string                `json:"result_evaluation_id,omitempty"`
	ResultVerifierKind         string                `json:"result_verifier_kind,omitempty"`
	ResultVerificationStrength string                `json:"result_verification_strength,omitempty"`
	Targets                    []FrontierTargetView  `json:"targets"`
	NearestClusters            []FrontierNearestView `json:"nearest_clusters"`
}

// FrontierGenerationView is a full generation pass.
type FrontierGenerationView struct {
	ID               string `json:"id"`
	ProblemID        string `json:"problem_id"`
	ClusterRunID     string `json:"cluster_run_id"`
	RunID            string `json:"run_id"`
	GeneratorVersion string `json:"generator_version"`
	RequestedCount   int    `json:"requested_count"`
	ProposalCount    int    `json:"proposal_count"`
	Revision         int    `json:"revision"`
	CreatedAt        string `json:"created_at"`
	// Admission audit (v29): zero for trusted code-derived generators.
	AdmissionCorrected  int                    `json:"admission_corrected"`
	AdmissionDowngraded int                    `json:"admission_downgraded"`
	AdmissionStripped   int                    `json:"admission_stripped"`
	AdmissionRejected   int                    `json:"admission_rejected"`
	AdmissionOverflow   int                    `json:"admission_overflow"`
	Proposals           []FrontierProposalView `json:"proposals"`
}

// FrontierGenerationSummaryView is a compact list row.
type FrontierGenerationSummaryView struct {
	ID               string `json:"id"`
	ClusterRunID     string `json:"cluster_run_id"`
	GeneratorVersion string `json:"generator_version"`
	RequestedCount   int    `json:"requested_count"`
	ProposalCount    int    `json:"proposal_count"`
	Revision         int    `json:"revision"`
	CreatedAt        string `json:"created_at"`
}

// FrontierGenerateResponse is returned by `newf frontier generate`.
type FrontierGenerateResponse struct {
	OK         bool                   `json:"ok"`
	Command    string                 `json:"command"`
	Store      string                 `json:"store"`
	Created    bool                   `json:"created"`
	Generation FrontierGenerationView `json:"generation"`
	// ExcludedStaleAuthority lists targetable-state invariants withheld from
	// THIS generation's target set because the campaign behind their state
	// assessed an obsolete population (F1). Their history is untouched; the
	// exclusion is reported so a reader never has to infer why a survivor was
	// not targeted.
	ExcludedStaleAuthority []StaleAuthorityView `json:"excluded_stale_authority,omitempty"`
}

// StaleAuthorityView reports one survivor withheld from current target
// selection, with the population mismatch that withheld it and the reassessment
// that restores eligibility.
type StaleAuthorityView struct {
	InvariantID string `json:"invariant_id"`
	State       string `json:"state"`
	Reason      string `json:"reason"`
	// AssessedClusterRunID is the population the authority campaign searched.
	AssessedClusterRunID string `json:"assessed_cluster_run_id"`
	// CurrentClusterRunID is the compatible current population it must match.
	CurrentClusterRunID string `json:"current_cluster_run_id"`
	// AuthorityPopulation is the population policy that campaign recorded
	// (`latest` or `discovery`).
	AuthorityPopulation     string `json:"authority_population_policy,omitempty"`
	ReassessmentInstruction string `json:"reassessment_instruction"`
}

// FrontierListResponse is returned by `newf frontier list`.
type FrontierListResponse struct {
	OK          bool                            `json:"ok"`
	Command     string                          `json:"command"`
	Store       string                          `json:"store"`
	Generations []FrontierGenerationSummaryView `json:"generations"`
}

// FrontierShowResponse is returned by `newf frontier show`.
type FrontierShowResponse struct {
	OK         bool                   `json:"ok"`
	Command    string                 `json:"command"`
	Store      string                 `json:"store"`
	Generation FrontierGenerationView `json:"generation"`
}

// EvaluationMetricView is one typed metric on an evaluation (no fake precision:
// the scale is explicit and the value is rendered from the scale-appropriate
// column).
type EvaluationMetricView struct {
	Name  string `json:"name"`
	Scale string `json:"scale"`
	Value string `json:"value,omitempty"`
}

// EvaluationTargetVerdictView is one per-target break verdict computed by an
// assessment against its assessed content revision, with the provenance regime
// that produced it ('recomputed' vs 'unverified_legacy').
type EvaluationTargetVerdictView struct {
	InvariantID string `json:"invariant_id"`
	Verdict     string `json:"verdict"`
	Violated    bool   `json:"violated"`
	Provenance  string `json:"provenance"`
}

// EvaluationView is one persisted evaluation. It ALWAYS carries the verifier
// kind and verification strength alongside the verdict, so a reader can never
// see an outcome without its epistemic strength (R1).
type EvaluationView struct {
	ID                   string `json:"id"`
	ProposalID           string `json:"proposal_id"`
	Verdict              string `json:"verdict"`
	VerifierKind         string `json:"verifier_kind"`
	VerificationStrength string `json:"verification_strength"`
	// VerificationSubject (v38/#21): what OBJECT the verdict is about —
	// annotation / realization / domain-goal. Empty only for pre-v38 history.
	VerificationSubject  string                        `json:"verification_subject,omitempty"`
	ConfidenceOrdinal    string                        `json:"confidence_ordinal,omitempty"`
	ToolName             string                        `json:"tool_name,omitempty"`
	ToolVersion          string                        `json:"tool_version,omitempty"`
	ProviderInvocationID string                        `json:"provider_invocation_id,omitempty"`
	Notes                string                        `json:"notes,omitempty"`
	Metrics              []EvaluationMetricView        `json:"metrics,omitempty"`
	TargetVerdicts       []EvaluationTargetVerdictView `json:"target_verdicts,omitempty"`
}

// EvaluationRunView is one evaluation pass with its evaluations.
type EvaluationRunView struct {
	ID                      string           `json:"id"`
	ProblemID               string           `json:"problem_id"`
	RunID                   string           `json:"run_id"`
	FrontierGenerationRunID string           `json:"frontier_generation_run_id,omitempty"`
	ClusterRunID            string           `json:"cluster_run_id,omitempty"`
	Mode                    string           `json:"mode"`
	RoutingPolicy           string           `json:"routing_policy"`
	EvaluationCount         int              `json:"evaluation_count"`
	CreatedAt               string           `json:"created_at"`
	Evaluations             []EvaluationView `json:"evaluations"`
}

// SkippedProposalView is one occurrence a batch evaluation deliberately did
// NOT assess, with the reason — a silent skip would hide the difference
// between "assessed and failing" and "never assessed in this context".
type SkippedProposalView struct {
	ProposalID string `json:"proposal_id"`
	Reason     string `json:"reason"`
}

// EvaluateResponse is returned by `newf evaluate`.
type EvaluateResponse struct {
	OK      bool              `json:"ok"`
	Command string            `json:"command"`
	Store   string            `json:"store"`
	Run     EvaluationRunView `json:"run"`
	// Skipped lists batch-mode occurrences excluded from this run because the
	// exact occurrence context already carries a ledger-recorded assessment.
	// Explicit by-id evaluation remains available for every entry.
	Skipped []SkippedProposalView `json:"skipped,omitempty"`
}

// EvaluationListResponse is returned by `newf evaluation list`.
type EvaluationListResponse struct {
	OK      bool                `json:"ok"`
	Command string              `json:"command"`
	Store   string              `json:"store"`
	Runs    []EvaluationRunView `json:"runs"`
}

// EvaluationShowResponse is returned by `newf evaluation show`.
type EvaluationShowResponse struct {
	OK      bool              `json:"ok"`
	Command string            `json:"command"`
	Store   string            `json:"store"`
	Run     EvaluationRunView `json:"run"`
}

// EvaluatedFailureView is one re-entered failure marker (R6): a proposal whose
// failing evaluation made its mechanism eligible for the next clustering pass.
type EvaluatedFailureView struct {
	ProposalID   string `json:"proposal_id"`
	EvaluationID string `json:"evaluation_id"`
	Verdict      string `json:"verdict"`
	CreatedAt    string `json:"created_at"`
}

// EvaluatedFailureListResponse is returned by `newf evaluation failures`.
type EvaluatedFailureListResponse struct {
	OK       bool                   `json:"ok"`
	Command  string                 `json:"command"`
	Store    string                 `json:"store"`
	Failures []EvaluatedFailureView `json:"failures"`
}
