# Domain model (`newf` v0)

## Boundary rule

Source evidence is immutable and never overwritten. Model/provider interpretation is stored in separate revisioned records linked by provenance. Generated claims cannot be promoted to source evidence without an explicit independent evidence record.

## Core entities

- **Problem**: research target and scope.
- **Source**: bibliographic or dataset reference.
- **Evidence**: immutable extracted/quoted factual unit from a source.
- **Approach**: normalized attempt descriptor tied to a normalization revision.
- **Mechanism**: typed mechanism representation for an approach.
- **Outcome**: failure / partial failure / partial success / success for an approach.
- **FailureBoundary**: normalized boundary encountered by failed/partial attempts.
- **MechanismCluster**: mechanistic family grouping over normalized approaches.
- **CandidateInvariant**: candidate failure invariant with lifecycle state.
- **InvariantChallenge**: challenge operation and result against an invariant.
- **FrontierProposal**: generated proposal targeting one or more invariants.
- **Evaluation**: proposal/experiment judgment and metric components.
- **EvaluationRun**: evaluation mode/budget/cutoff/baseline context for a batch of evaluations.
- **SuccessInvariant**: recurring structure in partial-success/success boundary crossing.
- **Run / provenance**: command execution, providers, prompts/schemas, config hashes.

## Ownership boundaries

- `EvidenceRecord` ownership: immutable factual record only (quoted/measured/externalized).
- `Approach`/`Mechanism`/`Outcome` ownership: derived interpretation owned by normalization revision.
- `CandidateInvariant` and `SuccessInvariant` ownership: hypothesis layer, always challengeable/revisionable.
- `Evaluation` ownership: judgment layer tied to explicit evaluator process and baseline.

## Relationships (high level)

- `Problem 1--* Source`
- `Source 1--* EvidenceRecord (immutable)`
- `NormalizationRevision 1--* Approach 1--1 Mechanism`
- `Approach 1--1 Outcome`, `Outcome *--* FailureBoundary`
- `ClusterRevision 1--* MechanismCluster`, `MechanismCluster *--* Approach`
- `InvariantRevision 1--* CandidateInvariant`, `CandidateInvariant *--* MechanismCluster`
- `CandidateInvariant 1--* InvariantChallenge`
- `FrontierGenerationRun 1--* FrontierProposal`, `FrontierProposal *--* CandidateInvariant`
- `EvaluationRun 1--* Evaluation`, `Evaluation -> FrontierProposal`
- `SuccessInvariantRevision 1--* SuccessInvariant`, linked to `Evaluation` and possibly `CandidateInvariant`
- `Run 1--* RunEvent`, and each derived record points to originating run + provider call(s)

## Invariant lifecycle

`proposed -> challenged -> surviving | weakened | split | merged | falsified | established`

- `split`: one parent invariant to many children invariants.
- `merged`: many parent invariants to one child invariant.
- `established`: requires independent external evidence class stronger than model output.

## Go type skeletons

```go
package domain

type ID string

type OutcomeClass string
const (
    OutcomeFailure       OutcomeClass = "failure"
    OutcomePartialFailure OutcomeClass = "partial_failure"
    OutcomePartialSuccess OutcomeClass = "partial_success"
    OutcomeSuccess        OutcomeClass = "success"
)

type InvariantState string
const (
    InvariantProposed    InvariantState = "proposed"
    InvariantChallenged  InvariantState = "challenged"
    InvariantSurviving   InvariantState = "surviving"
    InvariantWeakened    InvariantState = "weakened"
    InvariantSplit       InvariantState = "split"
    InvariantMerged      InvariantState = "merged"
    InvariantFalsified   InvariantState = "falsified"
    InvariantEstablished InvariantState = "established"
)

type Problem struct {
    ID          ID
    Slug        string
    Statement   string
    Description string
    CreatedAt   string
}

type Source struct {
    ID          ID
    ProblemID   ID
    Kind        string // literature|human|generated|experiment
    CanonicalRef string
    ContentHash string
    Metadata    SourceMetadata
}

type EvidenceRecord struct {
    ID          ID
    ProblemID   ID
    SourceID    ID
    EvidenceType string // theorem|proof_check|computation|experiment|argument|claim
    Snippet     string
    Locator     string // page/section/url fragment
    Strength    string
    CreatedAt   string
    Immutable   bool
}

type NormalizationRevision struct {
    ID          ID
    ProblemID   ID
    RunID       ID
    ParentID    *ID
    ConfigHash  string
    Status      string
}

type Approach struct {
    ID                      ID
    ProblemID               ID
    NormalizationRevisionID ID
    SourceID                ID
    SurfaceSummary          string
}

type MechanismAxis struct {
    AxisKey   string // e.g. scope, constructiveness
    ValueKey  string // e.g. local/global, constructive/existential
    VocabularyVersion string
}

type Mechanism struct {
    ID         ID
    ApproachID ID
    Representation []string
    Assumptions    []string
    Operators      []string
    Preserves      []string
    Axes           []MechanismAxis
}

type Outcome struct {
    ID         ID
    ApproachID ID
    Class      OutcomeClass
    Notes      string
}

type FailureBoundary struct {
    ID        ID
    ProblemID ID
    Label     string
}

type MechanismCluster struct {
    ID              ID
    ClusterRevisionID ID
    Label           string
    Rationale       string
}

type CandidateInvariant struct {
    ID                ID
    InvariantRevisionID ID
    Statement         string
    AbstractionLevel  string
    State             InvariantState
    ConfidenceOrdinal string
}

type InvariantChallenge struct {
    ID             ID
    CandidateInvariantID ID
    ChallengeType  string
    ResultState    InvariantState
    Summary        string
}

type FrontierProposal struct {
    ID                ID
    FrontierRunID     ID
    StructuralClaim   string
    NoveltyArgument   string
    CheapestFalsificationPath string
    ExpectedInfoGainOrdinal   string
    EvaluationCostOrdinal     string
    Result            string
}

type Evaluation struct {
    ID            ID
    EvaluationRunID ID
    FrontierProposalID *ID
    Verdict       string
    Confidence    string
}

type EvaluationRun struct {
    ID         ID
    ProblemID  ID
    Mode       string // proposal|holdout
    Baseline   string
    CutoffTime *string
}

type SuccessInvariant struct {
    ID                     ID
    SuccessInvariantRevisionID ID
    Statement              string
    BoundaryCrossing       string
    ConfidenceOrdinal      string
}

type Run struct {
    ID         ID
    ProblemID  ID
    Command    string
    Status     string
}
```

## Provider role interfaces (replaceable adapters)

```go
package provider

import "context"

type Normalizer interface {
    Normalize(ctx context.Context, in NormalizeInput) (NormalizeOutput, error)
}
type Clusterer interface {
    Cluster(ctx context.Context, in ClusterInput) (ClusterOutput, error)
}
type InvariantMiner interface {
    Mine(ctx context.Context, in MineInvariantInput) (MineInvariantOutput, error)
}
type InvariantCritic interface {
    Challenge(ctx context.Context, in ChallengeInput) (ChallengeOutput, error)
}
type FrontierGenerator interface {
    Generate(ctx context.Context, in GenerateInput) (GenerateOutput, error)
}
type Evaluator interface {
    Evaluate(ctx context.Context, in EvaluateInput) (EvaluateOutput, error)
}
type SuccessCompressor interface {
    Compress(ctx context.Context, in CompressInput) (CompressOutput, error)
}
```

- One concrete provider may implement multiple interfaces.
- Each call persists role-specific provider metadata so runs are reproducible.
- Inputs/outputs should be schema-validated structured objects, not free-form prose parsing.

## Mechanistic diversity vocabulary

- Axis definitions are versioned (`mechanism_axis_vocabulary` + `vocabulary_version` refs).
- Historical normalized records keep original axis key/value IDs to preserve comparability over time.
- New axes can be added without rewriting old mechanism rows.

## Provenance guarantees

- Every derived record references:
  - originating run
  - provider role + provider name/version
  - prompt/schema hash
  - parent revision (if applicable)
  - input evidence/record IDs used
- Promotion from generated claim to evidence requires explicit new `EvidenceRecord` with independent source and provenance.
