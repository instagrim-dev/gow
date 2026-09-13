package store

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestForeignKeyEnforcementSurvivesConnectionReplacement is the E3-1 regression
// (docs/reviews/2026-09-12-e3-w2-migration-run.md, engineering review E3 on the
// Migrate write path).
//
// The defect: `Open` established foreign-key enforcement with a single
// post-open `Exec`. SQLite's `foreign_keys` setting is connection-local, so that
// configured only whichever pooled connection served the call. A replacement
// connection started with enforcement OFF, and the failure mode was not an
// error — an INSERT referencing a nonexistent parent was ACCEPTED, leaving a
// readable orphan row.
//
// It was latent rather than active: `SetMaxOpenConns(1)` plus Go's idle-
// connection retention kept the one configured connection alive, so ordinary
// operation never observed it. That is exactly why it needs a regression: the
// guarantee depended on a pool-sizing decision that looks unrelated to
// integrity, and a future `SetConnMaxLifetime` or `SetMaxIdleConns` would have
// silently disabled referential integrity.
//
// Both halves are asserted. The pragma read-back alone would pass against a
// driver that reports the setting without applying it, so the consequence is
// exercised too.
func TestForeignKeyEnforcementSurvivesConnectionReplacement(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "newf.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer s.Close()
	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	assertEnforcing := func(stage string) {
		t.Helper()
		var on int
		if err := s.db.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&on); err != nil {
			t.Fatalf("%s: read foreign_keys: %v", stage, err)
		}
		if on != 1 {
			t.Fatalf("%s: foreign_keys = %d, want 1", stage, on)
		}
		// The consequence, not just the setting: a row referencing a parent that
		// does not exist must be refused.
		_, insErr := s.db.ExecContext(ctx,
			`INSERT INTO runs(id, problem_id, parent_run_id, operation, status, input_ref, tool_name, tool_version, started_at, completed_at)
			 VALUES('run_fk_probe_`+stage+`','prob_does_not_exist',NULL,'probe','succeeded','x','newf','probe','t0','t0')`)
		if insErr == nil {
			t.Fatalf("%s: a run referencing a nonexistent problem was accepted; referential integrity is unenforced", stage)
		}
		if !strings.Contains(strings.ToLower(insErr.Error()), "foreign key") {
			t.Fatalf("%s: insert failed for the wrong reason (want a foreign-key violation): %v", stage, insErr)
		}
	}

	assertEnforcing("initial")

	// Force the pooled connection to be discarded, so the next statement runs on
	// a connection `Open` never touched. This is the condition the one-shot
	// `Exec` could not survive.
	s.db.SetMaxIdleConns(0)
	if _, err := s.db.ExecContext(ctx, `SELECT 1`); err != nil {
		t.Fatalf("cycle connection: %v", err)
	}

	assertEnforcing("after connection replacement")
}

// TestMigrateRestoresEnforcementAfterItsSuspensionWindow re-probes the migration
// window after the E3-1 change, rather than assuming the DSN default and the
// explicit toggles compose correctly.
//
// `Migrate` deliberately runs with `PRAGMA foreign_keys = OFF` so a table
// rebuild can drop and recreate a referenced parent, and restores it in a
// `defer`. With enforcement now defaulted per connection, that suspension must
// still take effect during migration and must still be lifted afterward — on a
// replacement connection as well, which is the case the old code could not hold.
func TestMigrateRestoresEnforcementAfterItsSuspensionWindow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "newf.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer s.Close()

	readPragma := func(stage string) int {
		t.Helper()
		var on int
		if err := s.db.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&on); err != nil {
			t.Fatalf("%s: read foreign_keys: %v", stage, err)
		}
		return on
	}

	if got := readPragma("before migrate"); got != 1 {
		t.Fatalf("before migrate: foreign_keys = %d, want 1", got)
	}
	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	if got := readPragma("after migrate"); got != 1 {
		t.Fatalf("after migrate: foreign_keys = %d, want 1 (the suspension window must be closed)", got)
	}

	// A cancelled migration must not leave enforcement suspended either. The
	// cancellation is asserted to actually fail first, so a silently-succeeding
	// injection cannot make this case pass vacuously.
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Migrate(cancelled); err == nil {
		t.Fatal("a cancelled migration must fail; it reported success")
	}
	if got := readPragma("after cancelled migrate"); got != 1 {
		t.Fatalf("after cancelled migrate: foreign_keys = %d, want 1", got)
	}

	// And it survives connection replacement after the window, which is the
	// property the DSN change buys.
	s.db.SetMaxIdleConns(0)
	if _, err := s.db.ExecContext(ctx, `SELECT 1`); err != nil {
		t.Fatalf("cycle connection: %v", err)
	}
	if got := readPragma("after replacement"); got != 1 {
		t.Fatalf("after replacement: foreign_keys = %d, want 1", got)
	}
}

// TestSQLiteDSNIdentifiesTheIntendedFile pins the path handling, because the
// obvious implementation of this fix is silently wrong rather than broken.
//
// Concatenating `"file:" + path` truncates at a `#` (it parses as a URI
// fragment), so the database is created at a DIFFERENT path than the caller
// asked for — a corruption-by-misdirection that no error surfaces. Rendering a
// RELATIVE path through url.URL.Path yields `file://first-segment/...`, making
// the first segment a URI host, which fails to open at all.
func TestSQLiteDSNIdentifiesTheIntendedFile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// A directory whose name contains both a space and a hash.
	base := t.TempDir()
	awkward := filepath.Join(base, "a dir#with hash")
	dbPath := filepath.Join(awkward, "newf.db")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open(%q) error = %v", dbPath, err)
	}
	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	s.Close()

	// Assert the intended FILE exists. Reopening the same path is not sufficient:
	// a wrong-path DSN is self-consistent, so both writes and reads would agree
	// with each other while landing somewhere the caller never named.
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("the intended database file does not exist (%v): the DSN wrote to a different path", err)
	}

	// And assert nothing was created at the truncation point. Naive
	// concatenation of "file:" + path stops at the '#', producing a file named
	// "a dir" beside the intended directory.
	truncated := filepath.Join(base, "a dir")
	if _, err := os.Stat(truncated); err == nil {
		t.Fatalf("a database was created at %q: the path was truncated at the '#' as a URI fragment", truncated)
	}

	// The reopened store must see the migrated schema at that same path.
	reopened, err := Open(dbPath)
	if err != nil {
		t.Fatalf("reopen error = %v", err)
	}
	defer reopened.Close()
	var count int
	if err := reopened.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("read schema_migrations from the intended path: %v", err)
	}
	if count == 0 {
		t.Fatalf("the database at %q has no migrations: the DSN wrote to a different file", dbPath)
	}
}
