package pipeline

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

// --- #9 canonicalization / signature / comparison views ---

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
}

type SignatureView struct {
	ID                string                    `json:"id"`
	MechanismID       string                    `json:"mechanism_id"`
	SchemaVersion     string                    `json:"schema_version"`
	VocabularyVersion string                    `json:"vocabulary_version"`
	Fingerprint       string                    `json:"fingerprint"`
	OutcomeClass      string                    `json:"outcome_class"`
	Posture           map[string]string         `json:"posture"`
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
