// Package cluster implements deterministic, inspectable, versioned clustering
// over canonical mechanism signatures (#11). It contains no SQL, Cobra, or
// provider concerns: it is a pure function of the signatures it is handed plus
// an explicit, caller-owned ComparisonProfile.
//
// The governing constraint (AGENTS.md, EPIC.md M3) is that the clustering layer
// must not hide behind embeddings. Grouping here is the connected-components of
// a linkage graph whose edges are exactly canon.CompareWithProfile verdicts of
// mechanism-near (optionally surface-qualified). Every grouping decision is
// therefore reproducible from persisted signatures and the recorded profile.
package cluster

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
)

// AlgoMechanismV1 is the versioned clustering algorithm identity persisted on a
// cluster run so a grouping is always replayable against the exact rule used.
const AlgoMechanismV1 = "cluster/v1"

// Params is the caller-owned, versioned configuration for a clustering pass.
// The Profile decides which axes are decisive (the comparator only measures),
// so clustering never inherits a hidden global definition of "mechanism".
type Params struct {
	Profile     canon.ComparisonProfile
	AlgoVersion string
}

// withDefaults returns a copy of p with unset fields defaulted deterministically.
func (p Params) withDefaults() Params {
	out := p
	if out.AlgoVersion == "" {
		out.AlgoVersion = AlgoMechanismV1
	}
	if out.Profile.Version == "" {
		// Clustering DELIBERATELY defaults to classify/v1: family identity is
		// persisted per cluster run, and re-clustering a corpus under the v2
		// missing-data contract would change family membership — a research
		// decision that requires a new protocol revision, never a silent default
		// change. Assessment-side callers use ProfileMechanismV2 explicitly.
		out.Profile = canon.ProfileMechanismV1()
	}
	return out
}

// ThresholdsHash is a deterministic sha256 over the parameters that affect the
// grouping (algorithm + the profile's decisive axis selection). It is persisted
// so a run under a different decisive set is a distinct, non-colliding run.
func (p Params) ThresholdsHash() string {
	p = p.withDefaults()
	body := struct {
		Algo            string             `json:"algo"`
		ProfileVersion  string             `json:"profile_version"`
		DecisiveSet     []domain.FieldKind `json:"decisive_set"`
		SurfaceField    domain.FieldKind   `json:"surface_field"`
		DecisivePosture bool               `json:"decisive_posture"`
		DecisiveOutcome bool               `json:"decisive_outcome"`
	}{
		Algo:            p.AlgoVersion,
		ProfileVersion:  p.Profile.Version,
		DecisiveSet:     sortedFieldKinds(p.Profile.DecisiveSetFields),
		SurfaceField:    p.Profile.SurfaceField,
		DecisivePosture: p.Profile.DecisivePosture,
		DecisiveOutcome: p.Profile.DecisiveOutcome,
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		panic("cluster: thresholds body marshal failed: " + err.Error())
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

// Member is one signature assigned to a cluster, with its redundancy flag.
type Member struct {
	SignatureID string
	MechanismID string
	Fingerprint string
	// Redundant is true when this member is surface-distinct+mechanism-near with
	// at least one other member of the same family: it adds no new mechanism, so
	// it is a redundant failure sample (KTD-5).
	Redundant bool
}

// Cluster is one mechanism family: a connected component of the linkage graph.
type Cluster struct {
	// Fingerprint is a deterministic, order-independent identity over the sorted
	// member fingerprints AND sorted member signature IDs plus the version tuple
	// (KTD-3). Signature IDs are included so ambiguity-separated singletons that
	// share a resolved-only fingerprint remain distinct (see clusterFingerprint).
	Fingerprint string
	// RepresentativeSignatureID is the member with the lexicographically smallest
	// signature fingerprint (KTD-3): a stable, provenance-independent choice.
	RepresentativeSignatureID string
	// Isolate marks a singleton that had zero comparable neighbors (every pair
	// was incomparable on a decisive field). It is surfaced, never merged (KTD-4).
	Isolate bool
	// IntraVariation summarizes internal spread as counts of pairwise
	// classifications among members (empty for singletons).
	IntraVariation IntraVariation
	Members        []Member
	// OutcomeClasses is the sorted set of distinct outcome classes carried by the
	// family's members. Outcome is deliberately NON-decisive for mechanism
	// identity (a family is one mechanism regardless of how its instances turned
	// out), but the family's outcome must be reported truthfully: a family that
	// contains both failure and partial-success instances is Mixed, never
	// compressed to whichever class its representative happened to own (KTD-9).
	OutcomeClasses []domain.OutcomeClass
	// Mixed is true when the family spans more than one distinct outcome class.
	Mixed bool
}

// PrimaryOutcome returns the family's single outcome class, or "mixed" when the
// family spans more than one. Empty only for an outcome-less family.
func (c Cluster) PrimaryOutcome() domain.OutcomeClass {
	if c.Mixed {
		return domain.OutcomeMixed
	}
	if len(c.OutcomeClasses) == 1 {
		return c.OutcomeClasses[0]
	}
	return ""
}

// IntraVariation is a deterministic summary of within-family pairwise verdicts.
type IntraVariation struct {
	Identical           int // identical fingerprints
	MechanismNear       int // mechanism-near / surface-distinct+mechanism-near
	SurfaceDistinctNear int // subset of the above that is surface-distinct
	IncomparablePairs   int
	// MechanismDistinct counts pairs INSIDE this family that compared as
	// mechanism-distinct. The coherence guard (see BuildClustering) makes this 0
	// by construction for cluster/v1: no two members of a family may be
	// mechanism-distinct. It is recorded so any future linkage rule that relaxes
	// the guard cannot silently hide an incoherent family in the summary (KTD-10).
	MechanismDistinct int
}

// Distance is a representative-vs-representative comparison between two clusters.
type Distance struct {
	AIndex         int
	BIndex         int
	Classification canon.Classification
}

// AxisCoverage counts distinct canonical values present on one axis across all
// non-isolate families, and whether it is under-sampled.
type AxisCoverage struct {
	Axis               string
	DistinctValueCount int
	UnderSampled       bool
}

// Coverage is the population-level diversity report (R3).
type Coverage struct {
	DistinctFamilyCount  int
	RedundantMemberCount int
	Axes                 []AxisCoverage
}

// Clustering is the full deterministic result of a pass.
type Clustering struct {
	AlgoVersion    string
	ProfileVersion string
	ThresholdsHash string
	// InputSetHash is a deterministic, order-independent sha256 over the exact
	// signatures that were clustered (each member's signature id + fingerprint).
	// It is part of cluster-run identity so that adding, removing, or changing a
	// signature and re-clustering produces a NEW run rather than colliding with a
	// stale run under the same version tuple. Without it, the recursive
	// failure -> atlas -> recluster loop cannot learn from newly added failures
	// (KTD-1). signature_count alone is insufficient: two different populations
	// of equal size would collide.
	InputSetHash      string
	SchemaVersion     string
	VocabularyVersion string
	Clusters          []Cluster
	Distances         []Distance
	Coverage          Coverage
	// Status is "degraded" when the discrimination-loss guard fired, else "clean".
	Status             string
	DiscriminationLoss []canon.DiscriminationLoss
}

const (
	StatusClean    = "clean"
	StatusDegraded = "degraded"
)

// minAxisCoverage is the versioned threshold below which an axis is reported as
// under-sampled. One distinct value across all families means the corpus barely
// exercises that axis. Part of cluster/v1's thresholds_hash via AlgoVersion.
const minAxisCoverage = 2

// BuildClustering groups signatures into mechanism families deterministically.
// It is order-independent: inputs are sorted by fingerprint before any pairwise
// work, so permuting the input yields byte-identical clusters, representatives,
// and cluster fingerprints.
//
// A pair (a,b) is linked iff canon.CompareWithProfile(a,b,profile) classifies as
// mechanism-near or surface-distinct+mechanism-near AND no profile-decisive field
// was incomparable. Incomparable-on-a-decisive-field never links (KTD-4).
func BuildClustering(signatures []canon.MechanismSignature, params Params) Clustering {
	params = params.withDefaults()

	// Deterministic order: sort by fingerprint. Ties broken by mechanism id.
	sigs := make([]canon.MechanismSignature, len(signatures))
	copy(sigs, signatures)
	fpOf := make([]string, len(sigs))
	for i := range sigs {
		fpOf[i] = canon.Fingerprint(sigs[i])
	}
	order := make([]int, len(sigs))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		if fpOf[order[i]] != fpOf[order[j]] {
			return fpOf[order[i]] < fpOf[order[j]]
		}
		return sigs[order[i]].MechanismID < sigs[order[j]].MechanismID
	})
	ordered := make([]canon.MechanismSignature, len(sigs))
	orderedFP := make([]string, len(sigs))
	for newIdx, oldIdx := range order {
		ordered[newIdx] = sigs[oldIdx]
		orderedFP[newIdx] = fpOf[oldIdx]
	}

	n := len(ordered)
	result := Clustering{
		AlgoVersion:    params.AlgoVersion,
		ProfileVersion: params.Profile.Version,
		ThresholdsHash: params.ThresholdsHash(),
		InputSetHash:   inputSetHash(ordered, orderedFP),
		Status:         StatusClean,
	}
	if n > 0 {
		result.SchemaVersion = ordered[0].SchemaVersion
		result.VocabularyVersion = ordered[0].VocabularyVersion
	}

	// Pairwise verdicts up front so linkage, coherence, and variation all read
	// the same recorded classification.
	verdict := make(map[[2]int]canon.Classification, n*n/2)
	comparableNeighbor := make([]bool, n)
	// distinctPair[i][j] (i<j) records a mechanism-distinct verdict, used by the
	// coherence guard to forbid merges that would put a distinct pair in one
	// family.
	distinctPair := make(map[[2]int]bool, n)
	linkCand := make([][2]int, 0, n)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			cmp := canon.CompareWithProfile(ordered[i], ordered[j], params.Profile)
			verdict[[2]int{i, j}] = cmp.Classification
			if isDistinct(cmp.Classification) {
				distinctPair[[2]int{i, j}] = true
			}
			if incomparableOnDecisive(cmp, params.Profile) {
				continue
			}
			comparableNeighbor[i] = true
			comparableNeighbor[j] = true
			if isLink(cmp.Classification) {
				linkCand = append(linkCand, [2]int{i, j})
			}
		}
	}

	// Coherence-guarded union (KTD-2). The underlying mechanism-near relation is
	// NOT transitive (Jaccard/ordinal): A~B and B~C does not imply A~C. Plain
	// single-linkage connected components would therefore manufacture a family
	// {A,B,C} even when A is mechanism-distinct from C. We enforce the family
	// coherence invariant by construction:
	//
	//     forall a,b in C:  not mechanism-distinct(a,b)
	//
	// A near-edge (i,j) is applied only if no member of i's current component is
	// mechanism-distinct from any member of j's current component. Edges are
	// processed in deterministic (fingerprint) order, and the guard is symmetric
	// over whole components, so the result is order-independent for a fixed input.
	uf := newUnionFind(n)
	members := make([][]int, n) // component members keyed by current root
	for i := 0; i < n; i++ {
		members[i] = []int{i}
	}
	for _, e := range linkCand {
		ri, rj := uf.find(e[0]), uf.find(e[1])
		if ri == rj {
			continue
		}
		if componentsConflict(members[ri], members[rj], distinctPair) {
			// Merging would violate family coherence; skip this edge. The pair
			// remains near, but the components stay separate (KTD-2).
			continue
		}
		uf.union(e[0], e[1])
		root := uf.find(e[0])
		other := ri
		if root == ri {
			other = rj
		}
		members[root] = append(members[root], members[other]...)
		members[other] = nil
	}

	// Group by union-find root, preserving deterministic member order.
	groups := map[int][]int{}
	var rootOrder []int
	for i := 0; i < n; i++ {
		r := uf.find(i)
		if _, seen := groups[r]; !seen {
			rootOrder = append(rootOrder, r)
		}
		groups[r] = append(groups[r], i)
	}

	for _, root := range rootOrder {
		idxs := groups[root]
		c := Cluster{}
		// Members already in fingerprint order (i ascending == fp ascending).
		for _, i := range idxs {
			c.Members = append(c.Members, Member{
				SignatureID: signatureID(ordered[i]),
				MechanismID: ordered[i].MechanismID,
				Fingerprint: orderedFP[i],
			})
		}
		// Representative: smallest fingerprint == first member (already sorted).
		c.RepresentativeSignatureID = c.Members[0].SignatureID
		// Isolate: single member with no comparable neighbor at all.
		if len(idxs) == 1 && !comparableNeighbor[idxs[0]] {
			c.Isolate = true
		}
		// Intra-cluster variation + redundancy from stored pairwise verdicts.
		redundant := make([]bool, len(idxs))
		for a := 0; a < len(idxs); a++ {
			for b := a + 1; b < len(idxs); b++ {
				i, j := idxs[a], idxs[b]
				key := [2]int{i, j}
				if i > j {
					key = [2]int{j, i}
				}
				cl := verdict[key]
				switch cl {
				case canon.ClassSurfaceDistinctMechNear:
					c.IntraVariation.MechanismNear++
					c.IntraVariation.SurfaceDistinctNear++
					redundant[a] = true
					redundant[b] = true
				case canon.ClassMechanismNear:
					c.IntraVariation.MechanismNear++
					if orderedFP[i] == orderedFP[j] {
						c.IntraVariation.Identical++
					}
				case canon.ClassMechanismDistinct, canon.ClassSurfaceNearMechDistinct:
					// Should be impossible inside a coherent family; recorded so a
					// relaxed future linkage rule cannot hide the contradiction.
					c.IntraVariation.MechanismDistinct++
				case canon.ClassUnknown:
					c.IntraVariation.IncomparablePairs++
				}
			}
		}
		for a := range c.Members {
			c.Members[a].Redundant = redundant[a]
			if redundant[a] {
				result.Coverage.RedundantMemberCount++
			}
		}
		// Per-family outcome distribution (KTD-9): collect the distinct member
		// outcome classes rather than inheriting the representative's class.
		outcomeSeen := map[domain.OutcomeClass]struct{}{}
		for _, i := range idxs {
			oc := ordered[i].OutcomeClass
			if oc == "" {
				oc = domain.OutcomeUnknown
			}
			outcomeSeen[oc] = struct{}{}
		}
		for oc := range outcomeSeen {
			c.OutcomeClasses = append(c.OutcomeClasses, oc)
		}
		sort.Slice(c.OutcomeClasses, func(x, y int) bool { return c.OutcomeClasses[x] < c.OutcomeClasses[y] })
		c.Mixed = len(c.OutcomeClasses) > 1
		c.Fingerprint = clusterFingerprint(c.Members, result.AlgoVersion, result.ProfileVersion, result.SchemaVersion, result.VocabularyVersion, result.ThresholdsHash)
		result.Clusters = append(result.Clusters, c)
	}

	// Cluster order is stable: sort by cluster fingerprint so display/persist
	// ordinal is reproducible regardless of union-find traversal.
	sort.SliceStable(result.Clusters, func(i, j int) bool {
		return result.Clusters[i].Fingerprint < result.Clusters[j].Fingerprint
	})

	result.Coverage.DistinctFamilyCount = len(result.Clusters)
	result.Distances = representativeDistances(result.Clusters, ordered, params.Profile)
	result.Coverage.Axes = axisCoverage(ordered)

	// Discrimination-loss guard (KTD-8): reuse #9's system-owned invariant across
	// the whole population. If a lossy vocabulary erased every non-outcome
	// distinction between two mechanisms of different outcome classes, the run is
	// degraded (still persisted, explicitly flagged) rather than presenting a
	// clean family count.
	result.DiscriminationLoss = canon.AssertDiscriminationPreserved(ordered)
	if len(result.DiscriminationLoss) > 0 {
		result.Status = StatusDegraded
	}

	return result
}

// incomparableOnDecisive reports whether any profile-decisive set field was
// incomparable (a non-resolved claim on either side). Such a pair cannot be
// asserted to agree and therefore must not create a link (KTD-4).
func incomparableOnDecisive(cmp canon.Comparison, profile canon.ComparisonProfile) bool {
	decisive := map[domain.FieldKind]struct{}{}
	for _, k := range profile.DecisiveSetFields {
		decisive[k] = struct{}{}
	}
	for _, f := range cmp.Fields {
		if _, ok := decisive[f.FieldKind]; ok && f.Incomparable {
			return true
		}
	}
	if _, ok := decisive[domain.FieldBoundary]; ok && cmp.Boundary.Incomparable {
		return true
	}
	return false
}

func isLink(c canon.Classification) bool {
	return c == canon.ClassMechanismNear || c == canon.ClassSurfaceDistinctMechNear
}

// isDistinct reports whether the classification asserts the two mechanisms are
// distinct (either purely, or surface-near but mechanism-distinct). Such a pair
// must never share a family (family coherence, KTD-2).
func isDistinct(c canon.Classification) bool {
	return c == canon.ClassMechanismDistinct || c == canon.ClassSurfaceNearMechDistinct
}

// componentsConflict reports whether merging two components would place a
// mechanism-distinct pair in one family. It checks every cross pair against the
// recorded distinct set (symmetric; indices normalized to i<j).
func componentsConflict(a, b []int, distinctPair map[[2]int]bool) bool {
	for _, i := range a {
		for _, j := range b {
			lo, hi := i, j
			if lo > hi {
				lo, hi = hi, lo
			}
			if distinctPair[[2]int{lo, hi}] {
				return true
			}
		}
	}
	return false
}

// inputSetHash is a deterministic, order-independent sha256 over the exact
// clustered population: each member's signature id + resolved fingerprint. It is
// part of cluster-run identity so re-clustering after the population changes
// yields a new run (KTD-1).
func inputSetHash(ordered []canon.MechanismSignature, orderedFP []string) string {
	if len(ordered) == 0 {
		return ""
	}
	entries := make([]string, len(ordered))
	for i := range ordered {
		entries[i] = signatureID(ordered[i]) + "|" + orderedFP[i]
	}
	sort.Strings(entries)
	encoded, err := json.Marshal(entries)
	if err != nil {
		panic("cluster: input set hash marshal failed: " + err.Error())
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

// representativeDistances compares each pair of cluster representatives so the
// inter-cluster distance is reported over canonical structure (R2).
func representativeDistances(clusters []Cluster, ordered []canon.MechanismSignature, profile canon.ComparisonProfile) []Distance {
	bySigID := map[string]canon.MechanismSignature{}
	for _, s := range ordered {
		bySigID[signatureID(s)] = s
	}
	var out []Distance
	for i := 0; i < len(clusters); i++ {
		for j := i + 1; j < len(clusters); j++ {
			a := bySigID[clusters[i].RepresentativeSignatureID]
			b := bySigID[clusters[j].RepresentativeSignatureID]
			cmp := canon.CompareWithProfile(a, b, profile)
			out = append(out, Distance{AIndex: i, BIndex: j, Classification: cmp.Classification})
		}
	}
	return out
}

// axisCoverage counts distinct resolved canonical IDs per set field and distinct
// enum values per posture axis + outcome, flagging under-sampled axes (KTD-6).
func axisCoverage(sigs []canon.MechanismSignature) []AxisCoverage {
	setAxes := []struct {
		name   string
		getter func(canon.MechanismSignature) []canon.FieldClaim
	}{
		{"representation", func(s canon.MechanismSignature) []canon.FieldClaim { return s.Representations }},
		{"operator", func(s canon.MechanismSignature) []canon.FieldClaim { return s.Operators }},
		{"assumption", func(s canon.MechanismSignature) []canon.FieldClaim { return s.Assumptions }},
		{"preserves", func(s canon.MechanismSignature) []canon.FieldClaim { return s.Preserves }},
		{"breaks", func(s canon.MechanismSignature) []canon.FieldClaim { return s.Breaks }},
		{"auxiliary_object", func(s canon.MechanismSignature) []canon.FieldClaim { return s.AuxiliaryObjects }},
	}
	var out []AxisCoverage
	for _, ax := range setAxes {
		seen := map[domain.CanonicalID]struct{}{}
		for _, s := range sigs {
			for _, c := range ax.getter(s) {
				if c.State == domain.ResolutionResolved && c.CanonicalID != "" {
					seen[c.CanonicalID] = struct{}{}
				}
			}
		}
		out = append(out, AxisCoverage{
			Axis:               ax.name,
			DistinctValueCount: len(seen),
			UnderSampled:       len(seen) < minAxisCoverage,
		})
	}
	// Posture axes + outcome as distinct-enum coverage.
	postureAxes := []struct {
		name  string
		value func(canon.MechanismSignature) string
	}{
		{"posture.locality", func(s canon.MechanismSignature) string { return string(s.Posture.Locality) }},
		{"posture.construction", func(s canon.MechanismSignature) string { return string(s.Posture.Construction) }},
		{"posture.uncertainty", func(s canon.MechanismSignature) string { return string(s.Posture.Uncertainty) }},
		{"outcome", func(s canon.MechanismSignature) string { return string(s.OutcomeClass) }},
	}
	for _, ax := range postureAxes {
		seen := map[string]struct{}{}
		for _, s := range sigs {
			if v := ax.value(s); v != "" {
				seen[v] = struct{}{}
			}
		}
		out = append(out, AxisCoverage{
			Axis:               ax.name,
			DistinctValueCount: len(seen),
			UnderSampled:       len(seen) < minAxisCoverage,
		})
	}
	return out
}

// clusterFingerprint is a deterministic, order-independent identity over the
// cluster's members plus the version tuple (KTD-3).
//
// Identity is keyed on BOTH the sorted member signature fingerprints AND the
// sorted member signature IDs. The signature IDs are required for injectivity:
// a signature fingerprint excludes non-resolved (ambiguous/unknown) claims, so
// two signatures with identical resolved structure but a differing ambiguous
// decisive claim hash to the SAME fingerprint yet are incomparable-on-decisive
// and therefore land in DISTINCT singleton clusters. Keying on fingerprints
// alone would give those distinct clusters the same identity and collide under
// UNIQUE(cluster_run_id, cluster_fingerprint) at persist. The signature ID set
// is provenance-stable (one persisted msig_ row per member) and order-sorted,
// so it restores injectivity without breaking determinism or order-independence.
func clusterFingerprint(members []Member, algo, profile, schema, vocab, thresholds string) string {
	fps := make([]string, 0, len(members))
	sigIDs := make([]string, 0, len(members))
	for _, m := range members {
		fps = append(fps, m.Fingerprint)
		sigIDs = append(sigIDs, m.SignatureID)
	}
	sort.Strings(fps)
	sort.Strings(sigIDs)
	body := struct {
		Algo         string   `json:"algo"`
		Profile      string   `json:"profile"`
		Schema       string   `json:"schema"`
		Vocabulary   string   `json:"vocab"`
		Thresholds   string   `json:"thresholds"`
		Members      []string `json:"members"`
		MemberSigIDs []string `json:"member_sig_ids"`
	}{algo, profile, schema, vocab, thresholds, fps, sigIDs}
	encoded, err := json.Marshal(body)
	if err != nil {
		panic("cluster: cluster fingerprint marshal failed: " + err.Error())
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

// signatureID returns the persisted signature id when the signature carries one
// via its MechanismID-derived rehydration. The pipeline sets a SignatureID on
// the canon signature through the store record; when clustering pure in-memory
// fixtures the mechanism id is used as a stable stand-in.
func signatureID(s canon.MechanismSignature) string {
	if s.SignatureID != "" {
		return s.SignatureID
	}
	return s.MechanismID
}

func sortedFieldKinds(in []domain.FieldKind) []domain.FieldKind {
	out := make([]domain.FieldKind, len(in))
	copy(out, in)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// unionFind is a tiny deterministic disjoint-set with path halving and
// union-by-min-index so roots are stable across runs.
type unionFind struct{ parent []int }

func newUnionFind(n int) *unionFind {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return &unionFind{parent: p}
}

func (u *unionFind) find(x int) int {
	for u.parent[x] != x {
		u.parent[x] = u.parent[u.parent[x]]
		x = u.parent[x]
	}
	return x
}

func (u *unionFind) union(a, b int) {
	ra, rb := u.find(a), u.find(b)
	if ra == rb {
		return
	}
	// Attach larger root index under the smaller so the canonical root is stable.
	if ra < rb {
		u.parent[rb] = ra
	} else {
		u.parent[ra] = rb
	}
}
