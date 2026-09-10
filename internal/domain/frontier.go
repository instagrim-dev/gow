package domain

// Ordinal is a coarse, deliberately non-numeric magnitude band used by frontier
// scoring. The system records ordinal, component-wise judgments rather than
// fabricated floats (AGENTS.md: "Do not fabricate fake precision"). The bands are
// totally ordered low < medium < high.
type Ordinal string

const (
	OrdinalUnknown Ordinal = "unknown"
	OrdinalLow     Ordinal = "low"
	OrdinalMedium  Ordinal = "medium"
	OrdinalHigh    Ordinal = "high"
)

// Valid reports whether o is a defined ordinal band. "unknown" is valid and
// distinct from an empty/absent value.
func (o Ordinal) Valid() bool {
	switch o {
	case OrdinalUnknown, OrdinalLow, OrdinalMedium, OrdinalHigh:
		return true
	default:
		return false
	}
}

// Rank maps an ordinal to a total order for deterministic comparison. Higher is
// stronger; unknown sorts below low so an unscored component never outranks a
// scored one.
func (o Ordinal) Rank() int {
	switch o {
	case OrdinalHigh:
		return 3
	case OrdinalMedium:
		return 2
	case OrdinalLow:
		return 1
	default: // OrdinalUnknown or invalid
		return 0
	}
}
