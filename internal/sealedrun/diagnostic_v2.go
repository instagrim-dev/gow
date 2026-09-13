package sealedrun

import (
	"fmt"

	"github.com/instagrim-dev/newf/internal/screen"
	"github.com/instagrim-dev/newf/internal/shape"
)

// RunDiagnosticV2 re-runs a pack with the current failure-aware selector
// (shape.ControllerVersionV2) in the HG arm. ADAPTATION-REUSE WARNING,
// enforced in the label: when the pack under test is the one whose
// failures motivated the selector's design, this is a development
// diagnostic (the controller was fit to this pack's observed defects); it
// can never confirm the selector's value. Confirmation requires a pack
// authored after the selector froze. The outcome label carries both the
// controller version and the taint suffix, so no reader can mistake
// either the grade or which procedure produced it — record 019's run-3
// numbers belong to shape-selector/1, and a re-run under /2 must not
// borrow that label.
func RunDiagnosticV2(p Pack, budget int) (screen.Outcome, []EpisodeTrace, error) {
	if p.Label == "" {
		return screen.Outcome{}, nil, fmt.Errorf("a pack must carry its evidence-tier label")
	}
	return runWithSelector(p, budget, hgV2, p.Label+"+"+shape.ControllerVersionV2+"-adaptation-reuse-diagnostic")
}
