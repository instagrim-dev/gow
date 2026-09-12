// Package pipeline: operator-authored claim forms (v39, issue #21). Sample
// recurrence, transformation invariance, and obstruction are distinct
// propositions; the claim form fixes a candidate's proposition shape —
// predicate (mined) + quantifier + scope + claim role — explicitly, so
// refutation semantics are authored rather than inferred from measured
// coverage. Authoring `universal` does not strengthen evidence: it makes the
// claim MORE falsifiable (one in-scope counterexample refutes it) and records
// who fixed its shape and why.
package pipeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/store"
)

// AuthorClaimInput drives `newf invariant claim`.
type AuthorClaimInput struct {
	DBPath      string
	InvariantID string
	// Quantifier: universal | recurrent | existential.
	Quantifier string
	// ClaimRole: regularity | obstruction | enabling_condition |
	// boundary_hypothesis.
	ClaimRole string
	// Scope is the authored statement of the population/domain the claim is
	// about (e.g. "failure families of cluster run mcr_...; no claim beyond
	// the recorded corpus").
	Scope string
	// Note is the operator's basis, recorded verbatim.
	Note       string
	JSONOutput bool
}

// ClaimFormView is one authored claim form.
type ClaimFormView struct {
	ID          string `json:"id"`
	InvariantID string `json:"invariant_id"`
	Quantifier  string `json:"quantifier"`
	ClaimRole   string `json:"claim_role"`
	Scope       string `json:"scope"`
	AuthoredBy  string `json:"authored_by"`
	Basis       string `json:"basis"`
	CreatedAt   string `json:"created_at"`
}

// AuthorClaimResponse is returned by `newf invariant claim`.
type AuthorClaimResponse struct {
	OK      bool          `json:"ok"`
	Command string        `json:"command"`
	Store   string        `json:"store"`
	Claim   ClaimFormView `json:"claim"`
}

func claimFormView(row store.InvariantClaimFormRow) ClaimFormView {
	return ClaimFormView{
		ID: row.ID, InvariantID: row.InvariantID, Quantifier: row.Quantifier,
		ClaimRole: row.ClaimRole, Scope: row.Scope, AuthoredBy: row.AuthoredBy,
		Basis: row.Basis, CreatedAt: row.CreatedAt,
	}
}

var validQuantifiers = map[string]struct{}{"universal": {}, "recurrent": {}, "existential": {}}
var validClaimRoles = map[string]struct{}{"regularity": {}, "obstruction": {}, "enabling_condition": {}, "boundary_hypothesis": {}}

// AuthorInvariantClaim records one operator-authored claim form for a
// candidate invariant. Forms are append-only; the latest governs. This is an
// authoring act, not a verification: it changes what would refute the claim,
// never how well-supported it is.
func (a *App) AuthorInvariantClaim(ctx context.Context, input AuthorClaimInput) (AuthorClaimResponse, error) {
	if _, ok := validQuantifiers[input.Quantifier]; !ok {
		return AuthorClaimResponse{}, fmt.Errorf("quantifier must be universal, recurrent, or existential; got %q", input.Quantifier)
	}
	if _, ok := validClaimRoles[input.ClaimRole]; !ok {
		return AuthorClaimResponse{}, fmt.Errorf("claim role must be regularity, obstruction, enabling_condition, or boundary_hypothesis; got %q", input.ClaimRole)
	}
	if strings.TrimSpace(input.Scope) == "" {
		return AuthorClaimResponse{}, fmt.Errorf("--scope is required: a quantifier without an authored scope is exactly the unstated universal domain this record exists to prevent")
	}
	if strings.TrimSpace(input.Note) == "" {
		return AuthorClaimResponse{}, fmt.Errorf("--note is required: an authored claim form records its basis")
	}

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return AuthorClaimResponse{}, err
	}
	defer repoStore.Close()

	now := a.now()
	row := store.InvariantClaimFormRow{
		ID:          domain.NewInvariantClaimFormID(now),
		InvariantID: input.InvariantID,
		Quantifier:  input.Quantifier,
		ClaimRole:   input.ClaimRole,
		Scope:       strings.TrimSpace(input.Scope),
		AuthoredBy:  "operator",
		Basis:       strings.TrimSpace(input.Note),
		CreatedAt:   now.Format(timeLayout),
	}
	if err := repoStore.PersistInvariantClaimForm(ctx, row); err != nil {
		return AuthorClaimResponse{}, err
	}
	return AuthorClaimResponse{OK: true, Command: "invariant claim", Store: dbPath, Claim: claimFormView(row)}, nil
}
