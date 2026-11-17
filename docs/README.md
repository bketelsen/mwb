# Documentation

Docs are split by the question they answer (frostyard/core's four-category
shape, [core ADR-0025](https://github.com/frostyard/core/blob/main/docs/adr/0025-consolidate-repository-docs-into-docs.md)):

| Directory | Question | Contents |
| --- | --- | --- |
| [adr/](adr/) | **Why** did we choose this? | Repo-local Architecture Decision Records, immutable once accepted; superseded, never rewritten. Org-wide decisions live in frostyard/core; see [org-adrs.md](org-adrs.md) |
| [design/](design/) | **How** does it fit together? | Living documents describing the current architecture |
| [specs/](specs/) | **What exactly** is the contract? | Precise, testable interface definitions |
| [plans/](plans/) | **When/in what order** do we build? | Roadmaps and phase plans; updated as work lands |

This tree is the engineering documentation. The **user documentation site**
is separate: its content lives in [`site/content/`](../site/), configured
by `site/properdocs.yml`. ProperDocs and MaterialX build it, and
`.github/workflows/site.yml` publishes it to GitHub Pages from `main`.

Outside this tree, [`AGENTS.md`](../AGENTS.md) holds the rules for agents,
and [`.memory/README.md`](../.memory/README.md) describes the corrections inbox,
1 file for each record in `.memory/corrections/`.

## Index

### Decisions (ADRs)

- [0001 — Adopt the frostyard repository baseline](adr/0001-adopt-the-frostyard-repository-baseline.md)
- [ADR template](adr/TEMPLATE.md)
- Org-wide decisions: [org-adrs.md](org-adrs.md)

### Design

- [Design template](design/TEMPLATE.md)

No current design documents. The finished design plans below preserve the
implementation history.

### Specs

- [Spec template](specs/TEMPLATE.md)

No repository-local specs yet.

### Plans

No open plans. These 6 plans are finished and retained as implementation history:

- [Mouse Without Borders implementation plan](plans/2026-03-03-mwb-implementation-plan.md) — finished
- [Mouse Without Borders Linux client design](plans/2026-03-03-mwb-linux-client-design.md) — finished
- [Homebrew support design](plans/2026-03-04-homebrew-support-design.md) — finished
- [Homebrew support implementation](plans/2026-03-04-homebrew-support.md) — finished
- [macOS support design](plans/2026-03-04-macos-support-design.md) — finished
- [macOS support implementation](plans/2026-03-04-macos-support-plan.md) — finished
- [Plan template](plans/TEMPLATE.md)

## Conventions

- **New docs start from their category's `TEMPLATE.md`** (in each directory).
- New decision: create an ADR with the next number. If it reverses an old
  decision, mark the old one `Superseded by NNNN` rather than rewriting it.
  Accepted ADRs permit link-only maintenance under core ADR-0033.
  Decisions that bind more than this repository become ADRs in
  [frostyard/core](https://github.com/frostyard/core/tree/main/docs/adr), plus
  a line in [org-adrs.md](org-adrs.md).
- Design docs are updated in place to always reflect reality.
- Specs change only alongside the code that implements them.
- Cross-links between categories are mandatory in both directions.
- Adding a doc means adding it to the index above.
