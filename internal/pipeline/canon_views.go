package pipeline

import (
	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/store"
)

func termView(t store.TermRecord) TermView {
	return TermView{
		VocabularyVersion: t.VocabularyVersion,
		CanonicalID:       t.CanonicalID,
		FieldKind:         t.FieldKind,
		Description:       t.Description,
		ParentCanonicalID: t.ParentCanonicalID,
		Aliases:           t.Aliases,
	}
}

func termViews(terms []store.TermRecord) []TermView {
	out := make([]TermView, 0, len(terms))
	for _, t := range terms {
		out = append(out, termView(t))
	}
	return out
}

func signatureView(rec store.SignatureRecord) SignatureView {
	view := SignatureView{
		ID:                rec.ID,
		MechanismID:       rec.MechanismID,
		SchemaVersion:     rec.SchemaVersion,
		VocabularyVersion: rec.VocabularyVersion,
		Fingerprint:       rec.Fingerprint,
		OutcomeClass:      rec.OutcomeClass,
		Posture:           rec.Posture,
		CreatedAt:         rec.CreatedAt,
	}
	for _, c := range rec.FieldClaims {
		view.FieldClaims = append(view.FieldClaims, SignatureFieldClaimView{
			FieldKind:          c.FieldKind,
			SurfaceLabel:       c.SurfaceLabel,
			ResolutionState:    c.ResolutionState,
			CanonicalID:        c.CanonicalID,
			ClaimStatus:        c.ClaimStatus,
			SupportSnapshotID:  c.SupportSnapshotID,
			SupportLocator:     c.SupportLocator,
			Confidence:         c.Confidence,
			ClassifierContract: c.ClassifierContract,
		})
	}
	for _, b := range rec.Boundaries {
		view.Boundaries = append(view.Boundaries, SignatureBoundaryView{
			SurfaceLabel:    b.SurfaceLabel,
			ResolutionState: b.ResolutionState,
			CanonicalID:     b.CanonicalID,
			Relation:        b.Relation,
		})
	}
	return view
}

func comparisonView(cmp canon.Comparison) ComparisonView {
	view := ComparisonView{
		WeightsVersion:  cmp.WeightsVersion,
		ClassifyVersion: cmp.ClassifyVersion,
		Classification:  string(cmp.Classification),
		OutcomeEqual:    cmp.OutcomeEqual,
		Posture: PostureComparisonView{
			LocalityEqual:     cmp.Posture.LocalityEqual,
			ConstructionEqual: cmp.Posture.ConstructionEqual,
			UncertaintyEqual:  cmp.Posture.UncertaintyEqual,
		},
	}
	for _, f := range cmp.Fields {
		view.Fields = append(view.Fields, ComparisonFieldView{
			FieldKind:    string(f.FieldKind),
			OverlapCount: f.OverlapCount,
			UnionCount:   f.UnionCount,
			Jaccard:      f.Jaccard,
			Ordinal:      string(f.Ordinal),
			Incomparable: f.Incomparable,
		})
	}
	return view
}
