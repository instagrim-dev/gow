package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

func TestCLIProblemLifecycle(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")

	initStdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"}, initStdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("execute(init) code = %d", code)
	}

	var initResponse struct {
		ProblemID string `json:"problem_id"`
		RunID     string `json:"run_id"`
		Created   bool   `json:"created"`
	}
	if err := json.Unmarshal(initStdout.Bytes(), &initResponse); err != nil {
		t.Fatalf("json.Unmarshal(init) error = %v", err)
	}
	if !initResponse.Created {
		t.Fatal("init created = false, want true")
	}

	listStdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "problem", "list"}, listStdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("execute(problem list) code = %d", code)
	}
	if !strings.Contains(listStdout.String(), initResponse.ProblemID) {
		t.Fatalf("problem list output does not contain problem ID %q", initResponse.ProblemID)
	}

	showStdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "problem", "show", initResponse.ProblemID}, showStdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("execute(problem show) code = %d", code)
	}
	if !strings.Contains(showStdout.String(), "Erdős-Straus conjecture") {
		t.Fatalf("problem show output = %q", showStdout.String())
	}

	runStdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "run", "show", initResponse.RunID}, runStdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("execute(run show) code = %d", code)
	}
	if !strings.Contains(runStdout.String(), `"operation": "init"`) {
		t.Fatalf("run show output = %q", runStdout.String())
	}
}

func TestCLIJSONError(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	stdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "problem", "show", "run_01K4Y8X6YJJ66Y5QY9G7DNE1H2"}, stdout, &bytes.Buffer{}); code == 0 {
		t.Fatal("execute(problem show) succeeded, want failure")
	}

	var response struct {
		OK    bool `json:"ok"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal(error) error = %v", err)
	}
	if response.OK {
		t.Fatal("error response ok = true, want false")
	}
	if response.Error.Code != "invalid_input" {
		t.Fatalf("error code = %q, want invalid_input", response.Error.Code)
	}
}

func TestCLIJSONNotFoundError(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	missingProblemID := domain.NewProblemID(time.Now().UTC())

	stdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "problem", "show", missingProblemID}, stdout, &bytes.Buffer{}); code == 0 {
		t.Fatal("execute(problem show) succeeded, want failure")
	}

	var response struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal(error) error = %v", err)
	}
	if response.Error.Code != "not_found" {
		t.Fatalf("error code = %q, want not_found", response.Error.Code)
	}
}

func TestCLISourceIngestLifecycle(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, ".newf", "newf.db")
	sourceFile := filepath.Join(workspace, "paper-a.md")
	if err := os.WriteFile(sourceFile, []byte("# paper-a\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(text) error = %v", err)
	}

	binaryFile := filepath.Join(workspace, "paper-b.pdf")
	if err := os.WriteFile(binaryFile, []byte("%PDF-1.4\nbinary\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(binary) error = %v", err)
	}

	initStdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"}, initStdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("execute(init) code = %d", code)
	}
	var initResponse struct {
		ProblemID string `json:"problem_id"`
	}
	decodeJSONBuffer(t, initStdout, &initResponse)

	ingestText := runCLIJSON(t, []string{"--db", dbPath, "--json", "ingest", sourceFile, "--problem", initResponse.ProblemID})
	first := ingestText["results"].([]any)[0].(map[string]any)
	sourceID := first["source_id"].(string)
	snapshotID := first["snapshot_id"].(string)
	if first["status"] != "created_snapshot" {
		t.Fatalf("first ingest status = %v", first["status"])
	}

	ingestTextAgain := runCLIJSON(t, []string{"--db", dbPath, "--json", "ingest", sourceFile, "--problem", initResponse.ProblemID})
	second := ingestTextAgain["results"].([]any)[0].(map[string]any)
	if second["status"] != "existing_snapshot" {
		t.Fatalf("second ingest status = %v", second["status"])
	}

	if err := os.WriteFile(sourceFile, []byte("# paper-a revised\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(revision) error = %v", err)
	}
	ingestRevision := runCLIJSON(t, []string{"--db", dbPath, "--json", "ingest", sourceFile, "--problem", initResponse.ProblemID})
	third := ingestRevision["results"].([]any)[0].(map[string]any)
	if third["status"] != "new_revision" {
		t.Fatalf("revision ingest status = %v", third["status"])
	}

	ingestBinary := runCLIJSON(t, []string{"--db", dbPath, "--json", "ingest", binaryFile, "--problem", initResponse.ProblemID})
	fourth := ingestBinary["results"].([]any)[0].(map[string]any)
	if fourth["media_type"] != "application/pdf" {
		t.Fatalf("binary ingest media_type = %v, want application/pdf", fourth["media_type"])
	}

	sourceList := runCLIJSON(t, []string{"--db", dbPath, "--json", "source", "list", "--problem", initResponse.ProblemID})
	if len(sourceList["sources"].([]any)) != 2 {
		t.Fatalf("source list count = %d, want 2", len(sourceList["sources"].([]any)))
	}

	sourceShow := runCLIJSON(t, []string{"--db", dbPath, "--json", "source", "show", sourceID})
	if len(sourceShow["snapshots"].([]any)) != 2 {
		t.Fatalf("source show snapshots = %d, want 2", len(sourceShow["snapshots"].([]any)))
	}

	snapshotShow := runCLIJSON(t, []string{"--db", dbPath, "--json", "source", "snapshot", "show", snapshotID})
	objectAbs := snapshotShow["object_absolute_path"].(string)

	verify := runCLIJSON(t, []string{"--db", dbPath, "--json", "source", "snapshot", "verify", snapshotID})
	if verify["status"] != "verified" {
		t.Fatalf("verify status = %v, want verified", verify["status"])
	}

	if err := os.WriteFile(objectAbs, []byte("corrupt"), 0o644); err != nil {
		t.Fatalf("WriteFile(corrupt) error = %v", err)
	}
	verifyCorrupt := runCLIJSON(t, []string{"--db", dbPath, "--json", "source", "snapshot", "verify", snapshotID})
	if verifyCorrupt["status"] != "hash_mismatch" {
		t.Fatalf("verify corrupt status = %v, want hash_mismatch", verifyCorrupt["status"])
	}
}

func TestCLIIngestReportsPerInputFailures(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, ".newf", "newf.db")
	validFile := filepath.Join(workspace, "ok.md")
	if err := os.WriteFile(validFile, []byte("ok"), 0o644); err != nil {
		t.Fatalf("WriteFile(valid) error = %v", err)
	}

	initResponse := runCLIJSON(t, []string{"--db", dbPath, "--json", "init", "Problem"})
	problemID := initResponse["problem_id"].(string)

	result := runCLIJSON(t, []string{"--db", dbPath, "--json", "ingest", validFile, filepath.Join(workspace, "missing.md"), "--problem", problemID})
	var failed bool
	for _, item := range result["results"].([]any) {
		entry := item.(map[string]any)
		if entry["status"] == "failed" {
			failed = true
		}
	}
	if !failed {
		t.Fatalf("expected at least one failed ingest result: %#v", result["results"])
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	// tests run from cmd/newf; repo root is two levels up.
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

func TestCLINormalizeLifecycle(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, ".newf", "newf.db")
	fixture := filepath.Join(repoRoot(t), "fixtures", "erdos-straus-approach-a.md")

	initResponse := runCLIJSON(t, []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"})
	problemID := initResponse["problem_id"].(string)

	ingest := runCLIJSON(t, []string{"--db", dbPath, "--json", "ingest", fixture, "--problem", problemID})
	snapshotID := ingest["results"].([]any)[0].(map[string]any)["snapshot_id"].(string)

	normalizeResult := runCLIJSON(t, []string{"--db", dbPath, "--json", "normalize", "--source", snapshotID, "--problem", problemID, "--provider", "fixture"})
	results := normalizeResult["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("normalize results = %d, want 1", len(results))
	}
	first := results[0].(map[string]any)
	if first["status"] != "created" {
		t.Fatalf("normalize status = %v, want created", first["status"])
	}
	approaches := first["approaches"].([]any)
	if len(approaches) != 1 {
		t.Fatalf("approaches = %d, want 1", len(approaches))
	}
	approachID := approaches[0].(map[string]any)["approach_id"].(string)
	mechanismID := approaches[0].(map[string]any)["mechanism_id"].(string)

	// Idempotency: re-running without --force reports duplicate_existing.
	again := runCLIJSON(t, []string{"--db", dbPath, "--json", "normalize", "--source", snapshotID, "--problem", problemID, "--provider", "fixture"})
	if again["results"].([]any)[0].(map[string]any)["status"] != "duplicate_existing" {
		t.Fatalf("expected duplicate_existing on rerun, got %v", again["results"])
	}

	// --force creates a new lineage-preserving revision.
	forced := runCLIJSON(t, []string{"--db", dbPath, "--json", "normalize", "--source", snapshotID, "--problem", problemID, "--provider", "fixture", "--force"})
	if forced["results"].([]any)[0].(map[string]any)["status"] != "created" {
		t.Fatalf("expected forced rerun to create, got %v", forced["results"])
	}

	list := runCLIJSON(t, []string{"--db", dbPath, "--json", "approach", "list", "--problem", problemID})
	if len(list["approaches"].([]any)) != 1 {
		t.Fatalf("approach list = %d, want 1 (one logical identity across revisions)", len(list["approaches"].([]any)))
	}

	show := runCLIJSON(t, []string{"--db", dbPath, "--json", "approach", "show", approachID})
	if show["source_snapshot_id"] != snapshotID {
		t.Fatalf("approach show snapshot = %v, want %v", show["source_snapshot_id"], snapshotID)
	}
	support := show["support"].([]any)
	if len(support) == 0 {
		t.Fatal("approach show should expose field support")
	}
	var sawExplicit, sawInferred bool
	for _, item := range support {
		switch item.(map[string]any)["support_kind"] {
		case "explicit":
			sawExplicit = true
		case "inferred":
			sawInferred = true
		}
	}
	if !sawExplicit || !sawInferred {
		t.Fatalf("expected explicit and inferred support in %v", support)
	}

	revisions := runCLIJSON(t, []string{"--db", dbPath, "--json", "approach", "revisions", approachID})
	revList := revisions["revisions"].([]any)
	if len(revList) != 2 {
		t.Fatalf("approach revisions = %d, want 2 after forced rerun", len(revList))
	}
	// Lineage must be derived end-to-end: revisions are newest-first, so the
	// newest revision must supersede the original, and the original must start
	// the chain. This catches the whole pipeline->store->CLI path, not an
	// injected fixture field.
	newest := revList[0].(map[string]any)
	oldest := revList[1].(map[string]any)
	if newest["supersedes_revision_id"] != oldest["id"] {
		t.Fatalf("newest revision supersedes = %v, want prior revision id %v", newest["supersedes_revision_id"], oldest["id"])
	}
	if s, ok := oldest["supersedes_revision_id"]; ok && s != nil && s != "" {
		t.Fatalf("first revision must not supersede anything, got %v", s)
	}

	mechanism := runCLIJSON(t, []string{"--db", dbPath, "--json", "mechanism", "show", mechanismID})
	mech := mechanism["mechanism"].(map[string]any)
	if mech["locality"] != "local" || mech["construction_mode"] != "constructive" {
		t.Fatalf("mechanism axes = %v", mech)
	}
}

func TestCLINormalizeMultipleApproachesFromOneSource(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, ".newf", "newf.db")
	fixture := filepath.Join(repoRoot(t), "fixtures", "erdos-straus-two-approaches.md")

	initResponse := runCLIJSON(t, []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"})
	problemID := initResponse["problem_id"].(string)
	ingest := runCLIJSON(t, []string{"--db", dbPath, "--json", "ingest", fixture, "--problem", problemID})
	snapshotID := ingest["results"].([]any)[0].(map[string]any)["snapshot_id"].(string)

	normalizeResult := runCLIJSON(t, []string{"--db", dbPath, "--json", "normalize", "--source", snapshotID, "--problem", problemID, "--provider", "fixture"})
	approaches := normalizeResult["results"].([]any)[0].(map[string]any)["approaches"].([]any)
	if len(approaches) != 2 {
		t.Fatalf("one source produced %d approaches, want 2", len(approaches))
	}

	list := runCLIJSON(t, []string{"--db", dbPath, "--json", "approach", "list", "--problem", problemID})
	if len(list["approaches"].([]any)) != 2 {
		t.Fatalf("approach list = %d, want 2 distinct approaches", len(list["approaches"].([]any)))
	}

	// Surface wording differs but normalized mechanism axes must be comparable:
	// one local/constructive/deterministic, one global/existential/probabilistic.
	localityByIdentity := map[string]string{}
	for _, item := range list["approaches"].([]any) {
		entry := item.(map[string]any)
		detail := runCLIJSON(t, []string{"--db", dbPath, "--json", "approach", "show", entry["id"].(string)})
		mech := detail["mechanism"].(map[string]any)
		localityByIdentity[entry["logical_identity"].(string)] = mech["locality"].(string)
	}
	if localityByIdentity["erdos-straus/modular-residue-cover"] != "local" {
		t.Fatalf("modular approach locality = %v, want local", localityByIdentity["erdos-straus/modular-residue-cover"])
	}
	if localityByIdentity["erdos-straus/averaged-covering-density"] != "global" {
		t.Fatalf("averaged approach locality = %v, want global", localityByIdentity["erdos-straus/averaged-covering-density"])
	}
}

func TestCLINormalizeSkipsBinaryContent(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, ".newf", "newf.db")
	binaryFile := filepath.Join(workspace, "paper.pdf")
	if err := os.WriteFile(binaryFile, []byte("%PDF-1.4\nbinary\x00bytes"), 0o644); err != nil {
		t.Fatalf("WriteFile(binary) error = %v", err)
	}

	initResponse := runCLIJSON(t, []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"})
	problemID := initResponse["problem_id"].(string)
	ingest := runCLIJSON(t, []string{"--db", dbPath, "--json", "ingest", binaryFile, "--problem", problemID})
	snapshotID := ingest["results"].([]any)[0].(map[string]any)["snapshot_id"].(string)

	result := runCLIJSON(t, []string{"--db", dbPath, "--json", "normalize", "--source", snapshotID, "--problem", problemID, "--provider", "fixture"})
	entry := result["results"].([]any)[0].(map[string]any)
	if entry["status"] != "skipped" || entry["reason"] != "unsupported_content_representation" {
		t.Fatalf("expected typed unsupported-content skip, got %v", entry)
	}
}

func TestCLINormalizeUnknownProvider(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, ".newf", "newf.db")
	fixture := filepath.Join(repoRoot(t), "fixtures", "erdos-straus-approach-a.md")
	initResponse := runCLIJSON(t, []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"})
	problemID := initResponse["problem_id"].(string)
	ingest := runCLIJSON(t, []string{"--db", dbPath, "--json", "ingest", fixture, "--problem", problemID})
	snapshotID := ingest["results"].([]any)[0].(map[string]any)["snapshot_id"].(string)

	stdout := &bytes.Buffer{}
	code := execute(context.Background(), []string{"--db", dbPath, "--json", "normalize", "--source", snapshotID, "--problem", problemID, "--provider", "nope"}, stdout, &bytes.Buffer{})
	if code == 0 {
		t.Fatal("normalize with unknown provider succeeded, want failure")
	}
	var response struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	decodeJSONBuffer(t, stdout, &response)
	if response.Error.Code != "unknown_provider" {
		t.Fatalf("error code = %q, want unknown_provider", response.Error.Code)
	}
}

func runCLIJSON(t *testing.T, args []string) map[string]any {
	t.Helper()
	stdout := &bytes.Buffer{}
	if code := execute(context.Background(), args, stdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("execute(%v) code = %d, stdout=%s", args, code, stdout.String())
	}
	var decoded map[string]any
	decodeJSONBuffer(t, stdout, &decoded)
	return decoded
}

func decodeJSONBuffer(t *testing.T, buffer *bytes.Buffer, target any) {
	t.Helper()
	if err := json.Unmarshal(buffer.Bytes(), target); err != nil {
		t.Fatalf("json.Unmarshal() error = %v raw=%s", err, buffer.String())
	}
}

func TestCLIVocabularyListAndResolve(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	// init creates + migrates + seeds the vocabulary.
	runCLIJSON(t, []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"})

	list := runCLIJSON(t, []string{"--db", dbPath, "--json", "vocabulary", "list"})
	vocabs, ok := list["vocabularies"].([]any)
	if !ok || len(vocabs) == 0 {
		t.Fatalf("vocabulary list returned no vocabularies: %v", list)
	}

	resolve := runCLIJSON(t, []string{"--db", dbPath, "--json", "vocabulary", "resolve", "works residue-by-residue", "--field", "preserves"})
	if resolve["state"] != "resolved" {
		t.Fatalf("resolve state = %v, want resolved", resolve["state"])
	}
	if resolve["canonical_id"] != "domain.number_theory.property.residue_locality" {
		t.Fatalf("resolve canonical_id = %v", resolve["canonical_id"])
	}

	// An unknown label must not be coerced.
	unknown := runCLIJSON(t, []string{"--db", dbPath, "--json", "vocabulary", "resolve", "totally unseen phrase", "--field", "operator"})
	if unknown["state"] != "unknown" {
		t.Fatalf("unknown resolve state = %v, want unknown", unknown["state"])
	}
	if _, present := unknown["canonical_id"]; present {
		t.Fatalf("unknown resolve leaked canonical_id: %v", unknown)
	}
}

func TestCLISeedFixtureAndSignatureCompare(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	fixtureDir := filepath.Join(repoRoot(t), "testdata", "fixtures", "mechanism")

	initResp := runCLIJSON(t, []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"})
	problemID := initResp["problem_id"].(string)

	seedA := runCLIJSON(t, []string{"--db", dbPath, "--json", "mechanism", "seed-fixture",
		filepath.Join(fixtureDir, "case1_same_canonical_a.json"), "--problem", problemID})
	seedB := runCLIJSON(t, []string{"--db", dbPath, "--json", "mechanism", "seed-fixture",
		filepath.Join(fixtureDir, "case1_same_canonical_b.json"), "--problem", problemID})

	mechA := seedA["mechanism_ids"].([]any)[0].(string)
	mechB := seedB["mechanism_ids"].([]any)[0].(string)

	// case1: differently worded, same canonical mechanism -> equal fingerprints,
	// mechanism-near classification.
	compare := runCLIJSON(t, []string{"--db", dbPath, "--json", "mechanism", "compare", mechA, mechB})
	if compare["fingerprint_a"] != compare["fingerprint_b"] {
		t.Fatalf("case1 fingerprints differ: %v vs %v", compare["fingerprint_a"], compare["fingerprint_b"])
	}
	cls := compare["comparison"].(map[string]any)["classification"]
	if cls != "mechanism-near" && cls != "surface-distinct+mechanism-near" {
		t.Fatalf("case1 classification = %v, want a mechanism-near variant", cls)
	}

	// signature is idempotent through the CLI.
	sig1 := runCLIJSON(t, []string{"--db", dbPath, "--json", "mechanism", "signature", mechA})
	sig2 := runCLIJSON(t, []string{"--db", dbPath, "--json", "mechanism", "signature", mechA})
	if sig2["status"] != "existing" {
		t.Fatalf("second signature status = %v, want existing", sig2["status"])
	}
	if sig1["signature"].(map[string]any)["fingerprint"] != sig2["signature"].(map[string]any)["fingerprint"] {
		t.Fatal("idempotent signature produced different fingerprints via CLI")
	}
}

func TestCLISeedFixtureRequiresProblem(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	fixtureDir := filepath.Join(repoRoot(t), "testdata", "fixtures", "mechanism")
	runCLIJSON(t, []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"})

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := execute(context.Background(), []string{"--db", dbPath, "--json", "mechanism", "seed-fixture",
		filepath.Join(fixtureDir, "case1_same_canonical_a.json")}, stdout, stderr)
	if code == 0 {
		t.Fatal("seed-fixture without --problem succeeded, want failure")
	}
}

func TestCLIVocabularyShowNotFound(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	runCLIJSON(t, []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"})

	stdout := &bytes.Buffer{}
	code := execute(context.Background(), []string{"--db", dbPath, "--json", "vocabulary", "show", "core.operator.does_not_exist"}, stdout, &bytes.Buffer{})
	if code == 0 {
		t.Fatal("vocabulary show of missing term succeeded, want failure")
	}
	var response struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	decodeJSONBuffer(t, stdout, &response)
	if response.Error.Code != "not_found" {
		t.Fatalf("error code = %q, want not_found", response.Error.Code)
	}
}
