package pipeline

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// This file is the E3-2 regression (docs/reviews/2026-09-12-e3-w2-migration-run.md).
//
// The defect: `resolveProjectRevision` appended a bare `+dirty` marker, so two
// materially different trees at the same commit stringified identically. An
// assessment recorded against pre-fix code and one recorded against post-fix code
// declared the SAME `project_revision` dependency, so neither could go stale
// relative to the other — and because a compatible demonstrated nonconformance
// outranks any later favorable assessment, a blocker could never be retired. In a
// repository whose norm is a dirty working tree, that covers most changes.
//
// The acceptance criterion, stated in the review and asserted below: an
// assessment recorded before an edit goes stale after it, with no commit in
// between.

// gitFixture builds a throwaway repository with one commit. It is a fixture, not
// a manipulation of the real checkout: `AGENTS.md` forbids operating on the
// working tree, and a resolver test must be able to make a tree dirty.
func gitFixture(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q", "."},
		{"config", "user.email", "fixture@example.invalid"},
		{"config", "user.name", "fixture"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	writeFixtureFile(t, dir, "source.go", "package p\n\nconst V = 1\n")
	commitFixture(t, dir, "initial")
	return dir
}

func writeFixtureFile(t *testing.T, dir, name, content string) {
	t.Helper()
	full := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func commitFixture(t *testing.T, dir, message string) {
	t.Helper()
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-q", "-m", message}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

// TestProjectRevisionDistinguishesTreesAtOneCommit is E3-2's acceptance check.
func TestProjectRevisionDistinguishesTreesAtOneCommit(t *testing.T) {
	t.Parallel()
	dir := gitFixture(t)

	clean := resolveProjectRevisionIn(dir)
	if clean == projectRevisionUnknown || strings.Contains(clean, "+dirty") {
		t.Fatalf("a clean fixture must resolve to a bare commit id, got %q", clean)
	}

	// Edit a TRACKED file. No commit.
	writeFixtureFile(t, dir, "source.go", "package p\n\nconst V = 2\n")
	dirtyA := resolveProjectRevisionIn(dir)
	if !strings.HasPrefix(dirtyA, clean+"+dirty:") {
		t.Fatalf("a dirty tree must annotate its commit, got %q (clean was %q)", dirtyA, clean)
	}
	if strings.HasSuffix(dirtyA, projectRevisionUnknown) {
		t.Fatalf("the digest must be computable for an ordinary edit, got %q", dirtyA)
	}
	if dirtyA == clean {
		t.Fatal("an uncommitted edit must change the revision")
	}

	// A DIFFERENT edit at the same commit. This is the assertion the bare
	// `+dirty` marker could not satisfy: both trees are dirty at one commit, and
	// they must not be confusable.
	writeFixtureFile(t, dir, "source.go", "package p\n\nconst V = 3\n")
	dirtyB := resolveProjectRevisionIn(dir)
	if dirtyB == dirtyA {
		t.Fatalf("two different trees at one commit must not share a revision: both %q", dirtyA)
	}

	// Reverting the content must return the earlier revision: the identifier is
	// content-addressed, not a monotonic edit counter.
	writeFixtureFile(t, dir, "source.go", "package p\n\nconst V = 2\n")
	if got := resolveProjectRevisionIn(dir); got != dirtyA {
		t.Fatalf("restoring identical content must restore the revision: got %q, want %q", got, dirtyA)
	}

	// Committing the change yields a new bare commit id, with no dirty marker.
	commitFixture(t, dir, "second")
	committed := resolveProjectRevisionIn(dir)
	if strings.Contains(committed, "+dirty") {
		t.Fatalf("a committed tree must not be marked dirty, got %q", committed)
	}
	if committed == clean {
		t.Fatal("a new commit must change the revision")
	}
}

// TestProjectRevisionCoversUntrackedAndBinaryContent pins two properties a naive
// digest would miss.
//
// Untracked files are where new work lives before it is added, so a digest that
// ignored them would miss exactly the changes most likely to be under review.
//
// On binary content, honesty about what this asserts: `git diff HEAD` renders a
// modified binary as "Binary files … differ" with no bytes, but its `index` line
// still carries abbreviated before/after blob hashes — so this case passes with
// or without `--binary`, which was confirmed by mutation. It is retained as a
// property worth pinning (a one-byte binary change must move the revision), not
// as proof that `--binary` is load-bearing.
func TestProjectRevisionCoversUntrackedAndBinaryContent(t *testing.T) {
	t.Parallel()
	dir := gitFixture(t)
	base := resolveProjectRevisionIn(dir)

	// Untracked file: presence must change the revision.
	writeFixtureFile(t, dir, "notes/new.txt", "first\n")
	withUntracked := resolveProjectRevisionIn(dir)
	if withUntracked == base {
		t.Fatal("an untracked file must change the revision")
	}

	// Untracked CONTENT must change it too — not merely its path.
	writeFixtureFile(t, dir, "notes/new.txt", "second\n")
	withChangedUntracked := resolveProjectRevisionIn(dir)
	if withChangedUntracked == withUntracked {
		t.Fatal("changing an untracked file's content must change the revision, not just its path")
	}

	// Ignored files must NOT change it. In the real repository `docs/reviews` is
	// ignored, so writing a review must not invalidate the assessment it records.
	writeFixtureFile(t, dir, ".gitignore", "ignored/\n")
	commitFixture(t, dir, "add ignore rule")
	ignoredBase := resolveProjectRevisionIn(dir)
	writeFixtureFile(t, dir, "ignored/scratch.txt", "noise\n")
	if got := resolveProjectRevisionIn(dir); got != ignoredBase {
		t.Fatalf("an ignored file must not change the revision: %q -> %q", ignoredBase, got)
	}

	// Binary content: commit a binary file, then change one byte.
	if err := os.WriteFile(filepath.Join(dir, "blob.dat"), []byte{0x00, 0x01, 0x02}, 0o644); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	commitFixture(t, dir, "add binary")
	binBase := resolveProjectRevisionIn(dir)
	if err := os.WriteFile(filepath.Join(dir, "blob.dat"), []byte{0x00, 0x01, 0x03}, 0o644); err != nil {
		t.Fatalf("rewrite binary: %v", err)
	}
	binChanged := resolveProjectRevisionIn(dir)
	if binChanged == binBase {
		t.Fatal("a one-byte binary change must change the revision")
	}
}

// TestProjectRevisionReportsUnknownRatherThanGuessing pins the honesty rule: when
// the revision cannot be resolved, the resolver says so instead of returning a
// value that looks precise.
func TestProjectRevisionReportsUnknownRatherThanGuessing(t *testing.T) {
	t.Parallel()
	// A directory that is not a git checkout at all.
	if got := resolveProjectRevisionIn(t.TempDir()); got != projectRevisionUnknown {
		t.Fatalf("outside a checkout the revision must be %q, got %q", projectRevisionUnknown, got)
	}
}
