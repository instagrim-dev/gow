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
	// member fingerprints plus the version tuple (KTD-3).
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
}

// IntraVariation is a deterministic summary of within-family pairwise verdicts.
type IntraVariation struct {
	Identical           int // identical fingerprints
	MechanismNear       int // mechanism-near / surface-distinct+mechanism-near
	SurfaceDistinctNear int // subset of the above that is surface-distinct
	IncomparablePairs   int
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
	AlgoVersion       string
	ProfileVersion    string
	ThresholdsHash    string
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
		Status:         StatusClean,
	}
	if n > 0 {
		result.SchemaVersion = ordered[0].SchemaVersion
		result.VocabularyVersion = ordered[0].VocabularyVersion
	}

	// Union-find over the deterministic order.
	uf := newUnionFind(n)
	// linked[i][j] records the pairwise verdict for coverage/variation reuse.
	verdict := make(map[[2]int]canon.Classification, n*n/2)
	comparableNeighbor := make([]bool, n)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			cmp := canon.CompareWithProfile(ordered[i], ordered[j], params.Profile)
			verdict[[2]int{i, j}] = cmp.Classification
			if incomparableOnDecisive(cmp, params.Profile) {
				continue
			}
			comparableNeighbor[i] = true
			comparableNeighbor[j] = true
			if isLink(cmp.Classification) {
				uf.union(i, j)
			}
		}
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
// sorted member fingerprints plus the version tuple (KTD-3).
func clusterFingerprint(members []Member, algo, profile, schema, vocab, thresholds string) string {
	fps := make([]string, 0, len(members))
	for _, m := range members {
		fps = append(fps, m.Fingerprint)
	}
	sort.Strings(fps)
	body := struct {
		Algo       string   `json:"algo"`
		Profile    string   `json:"profile"`
		Schema     string   `json:"schema"`
		Vocabulary string   `json:"vocab"`
		Thresholds string   `json:"thresholds"`
		Members    []string `json:"members"`
	}{algo, profile, schema, vocab, thresholds, fps}
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
