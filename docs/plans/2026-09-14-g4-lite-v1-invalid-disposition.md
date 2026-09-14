# G4-lite V1 invalid disposition

**Status: retained invalid; no regrade or in-place retry.** The operator-recorded
private V1 dispatch completed all 72 cells but failed frozen-controller identity
conformance during independent grading. Its reported completion arithmetic was
also saturated at H0=24, H1=24, and HG=24, leaving no possible HG-over-H1
completion margin on that batch. The private artifacts and grade are not
imported into this repository, so their detailed adjudication remains
operator-reported rather than repository-verified.

The identity failure has no retrospective repair. It does not establish which
algorithm changed, because the old contract bound arbitrary snapshot-file bytes
without defining their relationship to the runner's decision-snapshot hashes.
The saturated completion counts are a ceiling diagnostic, not evidence against
the shaping policy's incremental value. Neither observation earns a shaping
claim, funding decision, or successor-pack authority.

The V1 raw episodes, answers, execution receipt, and result grid remain in the
separate custody location. They must not be reused as an untouched holdout.
Any later evaluation needs a successor design: preserve an open sensitivity
calibration, export and preflight the actual H0/H1/HG runtime identities from
the pinned executable before protected authoring, then obtain fresh custody and
execution authorization.
