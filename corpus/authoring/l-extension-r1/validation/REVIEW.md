# Outstanding review obligation — semantic support

Structural validation passed. **That is not the same as the claims being
supported**, and this file records what a human reviewer still owes before this
package is used for anything beyond inspection.

The validator checks that every support row has a well-formed locator pointing at
a paragraph that exists. It cannot check that the paragraph says what the
annotation claims. Recording an unsupported claim as a linter success is exactly
the failure mode `sha1-07` documents in the source domain: a claim assembled from
an unverified input inherits that input's fragility.

## 1. Per-record obligation

For each of the 12 records, confirm that each `support` row's cited paragraph
actually states the annotated field, and that `support_kind` is honest —
`explicit` only where the source states it, `inferred` where the authoring agent
derived it.

Records with **`inferred`** rows carrying the most weight, listed because they
are where a reviewer should start:

| Record | Inferred field | What to check |
|---|---|---|
| `sha1-01` | `mechanism.representations`, `outcome.boundary_statement` | The rotation-in-message-expansion mechanism is **not** attributed to Chabaud–Joux by their abstract. Confirm the note keeps this as inferred and does not assert it as their claim. |
| `sha1-02` | `mechanism.breaks` | That the paper's contribution breaks the one-trial-per-path-search cost model is the agent's framing, not the authors' wording. |
| `sha1-03` | `mechanism.breaks`, `outcome.class` | `partial_success` is a scoring decision against this corpus's objective, not a label the paper applies to itself. |
| `sha1-05` | `outcome.class` | Same, and confirm no absolute complexity figure has crept in — only the relative factor of 32 is verified. |
| `sha1-06` | `mechanism.operators`, `outcome.class` | Check the two-version divergence is represented without reconciliation. |
| `sha1-08` | `mechanism.operators`, `mechanism.breaks` | Confirm the note claims nothing about SHA-1 SAT difficulty, since SHA-1 was never attempted. |
| `sha1-09`, `sha1-10`, `sha1-12` | `outcome.class` | All three are scoring decisions relative to the declared objective. |
| `sha1-11` | `mechanism.operators` | "Industrialization rather than new cryptanalysis" is the agent's characterization. Verify it against the paper's own account of its contribution. |

## 2. Source-level limitations to carry into the review

Each is recorded in `manifest.json` `unresolved_items` and must not be silently
resolved:

1. **`src-chabaud-joux-1998`** — read scope is **abstract only**; full text was
   behind a publisher wall. Any claim about the paper's body is unverified.
2. **`src-joux-peyrin-2007`** — technical claims come from the ECRYPT Hash
   Workshop version, **not** the CRYPTO 2007 proceedings version. No absolute
   complexity figure is verified for this work.
3. **`src-mcdonald-hawkes-pieprzyk-2009`** — the withdrawal is verified verbatim;
   the **Eurocrypt 2009 rump-session announcement venue is not** and is not
   asserted anywhere.
4. **`src-manuel-2008`** — ePrint and journal versions reach **different
   conclusions**. Both are recorded. A reviewer must not collapse them.
5. **`src-shattered-2017`** — paper says 100 GPU years, vendor announcement says
   110. Both recorded, unreconciled.
6. **`src-leurent-peyrin-2020`** — paper estimates ~$45,000; project page reports
   ~$75,000 actually spent. Both recorded, unreconciled.
7. **`shattered.io` is a repurposed domain.** Any *prose* claim about the project
   must be cited from the Wayback capture. The PDFs at `shattered.io/static/`
   remain genuine — hash-verified locally.

## 3. Corrections already applied (verify they held)

Five factual corrections were applied during source verification. Confirm none
has leaked back in:

1. **`src-biham-chen-2004`** gives 65-round **SHA-0** collisions, *not* reduced
   SHA-1 collisions. Reduced-SHA-1 is Biham–Chen–Joux–Carribault–Lemuet–Jalby,
   EUROCRYPT 2005, not in this corpus.
2. **The 2^63 announcement** is **Wang, Andrew Yao & Frances Yao** — *not*
   Wang–Yin–Yu. Retained as register context only; it has no Work record because
   no primary account of its mechanism exists.
3. **Nossum 2012 is a preimage thesis**, not a failed SAT collision attempt. The
   SAT negative result is **Mironov–Zhang, SAT 2006** (`sha1-08`). Nossum is
   retained as an explicit exclusion in the register.
4. **SHA-256 digests of the SHAttered PDFs are locally computed**, not published
   on any authoritative page.
5. **The 2^61 figure in Chabaud–Joux** is scoped by the abstract to the
   **compression function**, not the full hash.

## 4. Verdict discipline

Record unsupported claims as **defects**, not as acceptable imprecision. If a
support row cannot be substantiated, the correct actions are: weaken
`support_kind` to `inferred`, remove the row, or mark the field unresolved —
never invent a locator that appears to support it.

If the review finds defects, the package gets a **new `content_revision`** with
regenerated hashes. Hashes must not be rewritten to make an unchanged claimed
revision pass; a validator regression enforces this.
