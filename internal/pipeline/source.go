package pipeline

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/store"
)

type IngestInput struct {
	DBPath     string
	ProblemID  string
	Paths      []string
	UseStdin   bool
	Name       string
	MediaType  string
	Recursive  bool
	JSONOutput bool
}

type SourceListInput struct {
	DBPath     string
	ProblemID  string
	JSONOutput bool
}

type SnapshotVerifyInput struct {
	DBPath     string
	SnapshotID string
	JSONOutput bool
}

func (a *App) IngestSources(ctx context.Context, input IngestInput) (IngestResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return IngestResponse{}, err
	}
	defer repoStore.Close()

	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return IngestResponse{}, err
	}

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   input.ProblemID,
		Operation:   "ingest",
		Status:      domain.RunStatusCompleted,
		InputRef:    "ingest",
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return IngestResponse{}, err
	}

	resolvedInputs, resolveErrs := a.resolveIngestInputs(input)
	results := make([]IngestItemResult, 0, len(resolvedInputs)+len(resolveErrs))

	objectsRoot := filepath.Join(filepath.Dir(dbPath), "objects")
	shaRoot := filepath.Join(objectsRoot, "sha256")
	for _, candidate := range resolvedInputs {
		result := IngestItemResult{Input: candidate.origin}
		outcome, ingestErr := ingestOneCandidate(ctx, repoStore, shaRoot, candidate, store.SnapshotAdmission{
			ProblemID:   input.ProblemID,
			Kind:        candidate.kind,
			LogicalName: candidate.logicalName,
			Origin:      candidate.origin,
			IngestRunID: run.ID,
			ObservedAt:  a.now(),
		}, input.MediaType)
		if ingestErr != nil {
			result.Status = "failed"
			result.Error = ingestErr.Error()
		} else {
			result.SourceID = outcome.Source.ID
			result.SnapshotID = outcome.Snapshot.ID
			result.Status = outcome.Status
			result.SHA256 = outcome.Snapshot.SHA256
			result.MediaType = outcome.Snapshot.MediaType
			result.Bytes = outcome.Snapshot.ByteLength
			result.CreatedSource = outcome.CreatedSource
		}
		results = append(results, result)
	}

	for _, resolveErr := range resolveErrs {
		results = append(results, IngestItemResult{
			Input:  resolveErr.input,
			Status: "failed",
			Error:  resolveErr.err.Error(),
		})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Input < results[j].Input })

	successCount := 0
	for _, result := range results {
		if result.Status != "failed" {
			successCount++
		}
	}

	return IngestResponse{
		OK:        true,
		Command:   "ingest",
		Store:     dbPath,
		ProblemID: input.ProblemID,
		RunID:     run.ID,
		Results:   results,
		Summary: IngestSummary{
			Total:     len(results),
			Succeeded: successCount,
			Failed:    len(results) - successCount,
		},
	}, nil
}

func (a *App) ListSources(ctx context.Context, input SourceListInput) (SourceListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return SourceListResponse{}, err
	}
	defer repoStore.Close()

	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return SourceListResponse{}, err
	}

	summaries, err := repoStore.ListSourcesWithSnapshotStats(ctx, input.ProblemID)
	if err != nil {
		return SourceListResponse{}, err
	}

	views := make([]SourceListView, 0, len(summaries))
	for _, summary := range summaries {
		views = append(views, sourceListView(summary.Source, summary.SnapshotCount, summary.LatestSnapshotID))
	}

	return SourceListResponse{
		OK:        true,
		Command:   "source list",
		Store:     dbPath,
		ProblemID: input.ProblemID,
		Sources:   views,
	}, nil
}

func (a *App) ShowSource(ctx context.Context, input LookupInput) (SourceShowResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return SourceShowResponse{}, err
	}
	defer repoStore.Close()

	source, err := repoStore.GetSource(ctx, input.ID)
	if err != nil {
		return SourceShowResponse{}, err
	}
	snapshots, err := repoStore.ListSourceSnapshots(ctx, source.ID)
	if err != nil {
		return SourceShowResponse{}, err
	}

	return SourceShowResponse{
		OK:        true,
		Command:   "source show",
		Store:     dbPath,
		Source:    sourceListView(source, len(snapshots), latestSnapshotID(snapshots)),
		Snapshots: snapshotViews(snapshots),
	}, nil
}

func (a *App) ShowSnapshot(ctx context.Context, input LookupInput) (SourceSnapshotShowResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return SourceSnapshotShowResponse{}, err
	}
	defer repoStore.Close()

	snapshot, err := repoStore.GetSourceSnapshot(ctx, input.ID)
	if err != nil {
		return SourceSnapshotShowResponse{}, err
	}
	source, err := repoStore.GetSource(ctx, snapshot.SourceID)
	if err != nil {
		return SourceSnapshotShowResponse{}, err
	}

	return SourceSnapshotShowResponse{
		OK:                 true,
		Command:            "source snapshot show",
		Store:              dbPath,
		Snapshot:           snapshotView(snapshot),
		Source:             sourceListView(source, 0, nil),
		ObjectAbsolutePath: filepath.Join(filepath.Dir(dbPath), "objects", filepath.FromSlash(snapshot.ObjectPath)),
	}, nil
}

func (a *App) VerifySnapshot(ctx context.Context, input SnapshotVerifyInput) (SourceSnapshotVerifyResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return SourceSnapshotVerifyResponse{}, err
	}
	defer repoStore.Close()

	snapshot, err := repoStore.GetSourceSnapshot(ctx, input.SnapshotID)
	if err != nil {
		return SourceSnapshotVerifyResponse{}, err
	}

	absPath := filepath.Join(filepath.Dir(dbPath), "objects", filepath.FromSlash(snapshot.ObjectPath))
	file, err := os.Open(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return SourceSnapshotVerifyResponse{
				OK:         true,
				Command:    "source snapshot verify",
				Store:      dbPath,
				SnapshotID: snapshot.ID,
				Status:     "missing_object",
			}, nil
		}
		return SourceSnapshotVerifyResponse{}, err
	}
	defer file.Close()

	hash := sha256.New()
	written, err := io.Copy(hash, file)
	if err != nil {
		return SourceSnapshotVerifyResponse{}, err
	}
	digest := hex.EncodeToString(hash.Sum(nil))
	status := "verified"
	if digest != snapshot.SHA256 || written != snapshot.ByteLength {
		status = "hash_mismatch"
	}

	return SourceSnapshotVerifyResponse{
		OK:         true,
		Command:    "source snapshot verify",
		Store:      dbPath,
		SnapshotID: snapshot.ID,
		Status:     status,
		SHA256:     snapshot.SHA256,
		Bytes:      snapshot.ByteLength,
	}, nil
}

type ingestCandidate struct {
	origin      string
	logicalName string
	kind        domain.SourceKind
	open        func() (io.ReadCloser, error)
}

type ingestResolveError struct {
	input string
	err   error
}

func (a *App) resolveIngestInputs(input IngestInput) ([]ingestCandidate, []ingestResolveError) {
	var candidates []ingestCandidate
	var failures []ingestResolveError

	if input.UseStdin {
		name := strings.TrimSpace(input.Name)
		if name == "" {
			name = "stdin"
		}
		candidates = append(candidates, ingestCandidate{
			origin:      "stdin",
			logicalName: name,
			kind:        domain.SourceKindStdin,
			open: func() (io.ReadCloser, error) {
				raw, err := io.ReadAll(a.stdin)
				if err != nil {
					return nil, err
				}
				if len(raw) == 0 {
					return nil, errors.New("stdin is empty")
				}
				return io.NopCloser(bytes.NewReader(raw)), nil
			},
		})
	}

	for _, path := range input.Paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			failures = append(failures, ingestResolveError{input: path, err: err})
			continue
		}
		info, err := os.Stat(absPath)
		if err != nil {
			failures = append(failures, ingestResolveError{input: path, err: err})
			continue
		}
		if info.IsDir() {
			if !input.Recursive {
				failures = append(failures, ingestResolveError{input: path, err: errors.New("path is a directory; use --recursive")})
				continue
			}
			walkErr := filepath.WalkDir(absPath, func(item string, d os.DirEntry, err error) error {
				if err != nil {
					failures = append(failures, ingestResolveError{input: item, err: err})
					return nil
				}
				if d.IsDir() {
					return nil
				}
				name := filepath.Base(item)
				candidates = append(candidates, ingestCandidate{
					origin:      item,
					logicalName: name,
					kind:        domain.SourceKindLocalPath,
					open: func() (io.ReadCloser, error) {
						return os.Open(item)
					},
				})
				return nil
			})
			if walkErr != nil {
				failures = append(failures, ingestResolveError{input: path, err: walkErr})
			}
			continue
		}

		logicalName := filepath.Base(absPath)
		if explicit := strings.TrimSpace(input.Name); explicit != "" && len(input.Paths) == 1 && !input.UseStdin {
			logicalName = explicit
		}
		pathCopy := absPath
		candidates = append(candidates, ingestCandidate{
			origin:      pathCopy,
			logicalName: logicalName,
			kind:        domain.SourceKindLocalPath,
			open: func() (io.ReadCloser, error) {
				return os.Open(pathCopy)
			},
		})
	}

	return candidates, failures
}

func ingestOneCandidate(ctx context.Context, repoStore problemStore, objectsRoot string, candidate ingestCandidate, admission store.SnapshotAdmission, mediaTypeOverride string) (store.SnapshotAdmissionResult, error) {
	reader, err := candidate.open()
	if err != nil {
		return store.SnapshotAdmissionResult{}, err
	}
	defer reader.Close()

	written, digest, mediaType, objectPath, err := persistObject(objectsRoot, candidate.logicalName, mediaTypeOverride, reader)
	if err != nil {
		return store.SnapshotAdmissionResult{}, err
	}

	admission.SHA256 = digest
	admission.ByteLength = written
	admission.MediaType = mediaType
	admission.ObjectPath = objectPath
	return repoStore.CreateSourceSnapshot(ctx, admission)
}

func persistObject(shaRoot, logicalName, mediaTypeOverride string, reader io.Reader) (int64, string, string, string, error) {
	if err := os.MkdirAll(shaRoot, 0o755); err != nil {
		return 0, "", "", "", err
	}
	tmpDir := filepath.Join(shaRoot, "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return 0, "", "", "", err
	}

	temp, err := os.CreateTemp(tmpDir, "stage-*")
	if err != nil {
		return 0, "", "", "", err
	}
	tempPath := temp.Name()
	cleanupTemp := func() {
		_ = temp.Close()
		_ = os.Remove(tempPath)
	}

	hash := sha256.New()
	var written int64
	var sniff []byte
	buf := make([]byte, 32*1024)
	for {
		n, readErr := reader.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			if len(sniff) < 512 {
				need := 512 - len(sniff)
				if need > len(chunk) {
					need = len(chunk)
				}
				sniff = append(sniff, chunk[:need]...)
			}
			if _, err := hash.Write(chunk); err != nil {
				cleanupTemp()
				return 0, "", "", "", err
			}
			wrote, err := temp.Write(chunk)
			if err != nil {
				cleanupTemp()
				return 0, "", "", "", err
			}
			written += int64(wrote)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			cleanupTemp()
			return 0, "", "", "", readErr
		}
	}

	if err := temp.Close(); err != nil {
		cleanupTemp()
		return 0, "", "", "", err
	}

	digest := hex.EncodeToString(hash.Sum(nil))
	prefix := digest[:2]
	destDir := filepath.Join(shaRoot, prefix)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		_ = os.Remove(tempPath)
		return 0, "", "", "", err
	}
	dest := filepath.Join(destDir, digest)
	if _, err := os.Stat(dest); err == nil {
		_ = os.Remove(tempPath)
	} else if errors.Is(err, os.ErrNotExist) {
		if err := os.Rename(tempPath, dest); err != nil {
			_ = os.Remove(tempPath)
			return 0, "", "", "", err
		}
	} else {
		_ = os.Remove(tempPath)
		return 0, "", "", "", err
	}

	mediaType := detectMediaType(logicalName, sniff, mediaTypeOverride)
	return written, digest, mediaType, filepath.ToSlash(filepath.Join("sha256", prefix, digest)), nil
}

func detectMediaType(logicalName string, sniff []byte, override string) string {
	if explicit := strings.TrimSpace(override); explicit != "" {
		return explicit
	}
	detected := http.DetectContentType(sniff)
	ext := strings.ToLower(filepath.Ext(logicalName))
	switch ext {
	case ".md", ".markdown":
		return "text/markdown"
	case ".json":
		return "application/json"
	case ".yaml", ".yml":
		return "application/yaml"
	}
	if detected == "" {
		return "application/octet-stream"
	}
	return detected
}

func sourceListView(source domain.Source, snapshotCount int, latestSnapshotID *string) SourceListView {
	view := SourceListView{
		ID:               source.ID,
		ProblemID:        source.ProblemID,
		Kind:             string(source.Kind),
		LogicalName:      source.LogicalName,
		Origin:           source.Origin,
		CreatedAt:        source.CreatedAt.Format(timeLayout),
		SnapshotCount:    snapshotCount,
		LatestSnapshotID: latestSnapshotID,
	}
	return view
}

func latestSnapshotID(snapshots []domain.SourceSnapshot) *string {
	if len(snapshots) == 0 {
		return nil
	}
	return &snapshots[0].ID
}

func snapshotViews(snapshots []domain.SourceSnapshot) []SourceSnapshotView {
	out := make([]SourceSnapshotView, 0, len(snapshots))
	for _, snapshot := range snapshots {
		out = append(out, snapshotView(snapshot))
	}
	return out
}

func snapshotView(snapshot domain.SourceSnapshot) SourceSnapshotView {
	return SourceSnapshotView{
		ID:                   snapshot.ID,
		SourceID:             snapshot.SourceID,
		SHA256:               snapshot.SHA256,
		ByteLength:           snapshot.ByteLength,
		MediaType:            snapshot.MediaType,
		ObjectPath:           snapshot.ObjectPath,
		ObservedAt:           snapshot.ObservedAt.Format(timeLayout),
		IngestRunID:          snapshot.IngestRunID,
		SupersedesSnapshotID: snapshot.SupersedesSnapshotID,
	}
}

const timeLayout = "2006-01-02T15:04:05.999999999Z07:00"
