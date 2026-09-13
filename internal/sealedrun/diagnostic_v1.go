package sealedrun

import (
	"fmt"

	"github.com/instagrim-dev/newf/internal/screen"
)

// RunDiagnosticV1 re-runs a pack with the v1 selector (shape-selector/1)
// in the HG arm. ADAPTATION-REUSE WARNING, enforced in the label: when
// the pack under test is the one whose failures motivated v1, this is a
// development diagnostic (the controller was fit to this pack's observed
// defects); it can never confirm v1's value. Confirmation requires a
// pack authored after v1 froze. The outcome label carries the suffix so
// no reader can mistake the grade.
func RunDiagnosticV1(p Pack, budget int) (screen.Outcome, []EpisodeTrace, error) {
	if p.Label == "" {
		return screen.Outcome{}, nil, fmt.Errorf("a pack must carry its evidence-tier label")
	}
	return runWithSelector(p, budget, hgV1, p.Label+"+v1-adaptation-reuse-diagnostic")
}
