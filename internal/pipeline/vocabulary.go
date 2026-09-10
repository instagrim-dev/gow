package pipeline

import (
	"context"
	"fmt"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/store"
)

// VocabularyListInput requests the persisted vocabularies (and optionally the
// terms of one version).
type VocabularyListInput struct {
	DBPath     string
	Version    string // optional: list terms of this version
	FieldKind  string // optional filter
	JSONOutput bool
}

// VocabularyShowInput requests one term across versions.
type VocabularyShowInput struct {
	DBPath      string
	CanonicalID string
	Version     string // optional; when empty, search all versions
	JSONOutput  bool
}

// VocabularyResolveInput resolves a candidate label deterministically.
type VocabularyResolveInput struct {
	DBPath     string
	Label      string
	Field      string
	Version    string // optional; default latest seeded
	Novel      bool
	JSONOutput bool
}

// seedVocabularies persists the in-repo canonical vocabularies + rubric records
// idempotently. Called after migrate so read/resolve always have data offline.
func (a *App) seedVocabularies(ctx context.Context, repoStore problemStore) error {
	now := a.now().Format(timeLayout)
	for _, v := range canon.SeededVocabularies() {
		input := store.VocabularySeedInput{
			Version:   v.Version(),
			Notes:     "seeded from in-repo definition",
			CreatedAt: now,
			Rejected:  v.RejectedKeys(),
		}
		for _, term := range v.Terms("") {
			aliases := make([]string, 0, len(term.Aliases))
			for _, alias := range term.Aliases {
				// Persist the normalized alias key so the stored uniqueness key
				// (version, field_kind, alias_normalized) matches exactly what
				// the in-memory resolver looks up. Normalize is idempotent.
				aliases = append(aliases, canon.Normalize(alias))
			}
			input.Terms = append(input.Terms, store.TermRecord{
				VocabularyVersion: v.Version(),
				CanonicalID:       string(term.CanonicalID),
				FieldKind:         string(term.FieldKind),
				Description:       term.Description,
				ParentCanonicalID: string(term.Parent),
				Aliases:           aliases,
			})
		}
		if err := repoStore.SeedVocabulary(ctx, input); err != nil {
			return err
		}
	}
	return nil
}

// ListVocabulary returns vocabulary versions, or the terms of a single version.
func (a *App) ListVocabulary(ctx context.Context, input VocabularyListInput) (VocabularyListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return VocabularyListResponse{}, err
	}
	defer repoStore.Close()

	resp := VocabularyListResponse{OK: true, Command: "vocabulary list", Store: dbPath}

	if input.Version != "" {
		terms, err := repoStore.ListTerms(ctx, input.Version, input.FieldKind)
		if err != nil {
			return VocabularyListResponse{}, err
		}
		resp.Version = input.Version
		resp.Terms = termViews(terms)
		return resp, nil
	}

	vocabs, err := repoStore.ListVocabularies(ctx)
	if err != nil {
		return VocabularyListResponse{}, err
	}
	for _, v := range vocabs {
		resp.Vocabularies = append(resp.Vocabularies, VocabularyView{
			Version:   v.Version,
			CreatedAt: v.CreatedAt,
			Notes:     v.Notes,
		})
	}
	return resp, nil
}

// ShowVocabularyTerm returns one term (aliases, parent, containing versions).
func (a *App) ShowVocabularyTerm(ctx context.Context, input VocabularyShowInput) (VocabularyShowResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return VocabularyShowResponse{}, err
	}
	defer repoStore.Close()

	versions := []string{}
	if input.Version != "" {
		versions = append(versions, input.Version)
	} else {
		all, err := repoStore.ListVocabularies(ctx)
		if err != nil {
			return VocabularyShowResponse{}, err
		}
		for _, v := range all {
			versions = append(versions, v.Version)
		}
	}

	resp := VocabularyShowResponse{OK: true, Command: "vocabulary show", Store: dbPath, CanonicalID: input.CanonicalID}
	for _, version := range versions {
		term, err := repoStore.GetTerm(ctx, version, input.CanonicalID)
		if err != nil {
			continue // term not present in this version
		}
		resp.Found = true
		resp.Terms = append(resp.Terms, termView(term))
	}
	if !resp.Found {
		return VocabularyShowResponse{}, fmt.Errorf("%w: canonical id %s", store.ErrNotFound, input.CanonicalID)
	}
	return resp, nil
}

// ResolveVocabulary resolves a candidate label against a persisted vocabulary
// version using the pure canon.Resolve. It never coerces to the nearest term.
func (a *App) ResolveVocabulary(ctx context.Context, input VocabularyResolveInput) (VocabularyResolveResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return VocabularyResolveResponse{}, err
	}
	defer repoStore.Close()

	fieldKind := domain.FieldKind(input.Field)
	if !fieldKind.Valid() {
		return VocabularyResolveResponse{}, fmt.Errorf("invalid field kind %q", input.Field)
	}

	version := input.Version
	if version == "" {
		version = canon.VocabularyMechanismV1
	}

	vocab, err := a.loadVocabulary(ctx, repoStore, version)
	if err != nil {
		return VocabularyResolveResponse{}, err
	}

	res := vocab.Resolve(fieldKind, input.Label, input.Novel)
	view := VocabularyResolveResponse{
		OK:          true,
		Command:     "vocabulary resolve",
		Store:       dbPath,
		Version:     version,
		Field:       input.Field,
		SurfaceKey:  res.SurfaceKey,
		State:       string(res.State),
		CanonicalID: string(res.CanonicalID),
	}
	for _, c := range res.Candidates {
		view.Candidates = append(view.Candidates, string(c))
	}
	return view, nil
}

// loadVocabulary rebuilds a canon.Vocabulary from persisted rows so resolution
// runs against exactly the persisted version (satisfies reproducibility).
func (a *App) loadVocabulary(ctx context.Context, repoStore problemStore, version string) (*canon.Vocabulary, error) {
	terms, err := repoStore.ListTerms(ctx, version, "")
	if err != nil {
		return nil, err
	}
	defs := make([]canon.TermDef, 0, len(terms))
	for _, t := range terms {
		defs = append(defs, canon.TermDef{
			CanonicalID: domain.CanonicalID(t.CanonicalID),
			FieldKind:   domain.FieldKind(t.FieldKind),
			Description: t.Description,
			Parent:      domain.CanonicalID(t.ParentCanonicalID),
			Aliases:     t.Aliases,
		})
	}
	rejected, err := repoStore.ListRejected(ctx, version)
	if err != nil {
		return nil, err
	}
	return canon.BuildVocabularyWithRejected(version, defs, rejected)
}
