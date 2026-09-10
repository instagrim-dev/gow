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
