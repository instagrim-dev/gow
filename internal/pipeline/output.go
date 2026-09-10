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
