# Org-wide decisions (frostyard/core ADRs)

Conventions this repository follows that are decided at the org level are
recorded as ADRs in
[frostyard/core](https://github.com/frostyard/core/tree/main/docs/adr).
Although mwb is in the bketelsen account, it adopts these conventions under
[ADR-0001](adr/0001-adopt-the-frostyard-repository-baseline.md).
These are binding requirements, not a claim that every repository setting
or baseline item is already configured.

- [ADR-0012 — svu versioning and rolling dev prerelease](https://github.com/frostyard/core/blob/main/docs/adr/0012-svu-versioning-and-rolling-dev-prerelease.md) — `svu` derives versions and `make bump` tags releases; mwb has no rolling `dev` snapshot.
- [ADR-0018 — Org-wide agent instruction and knowledge surfaces](https://github.com/frostyard/core/blob/main/docs/adr/0018-org-wide-agent-instruction-and-knowledge-surfaces.md) — `AGENTS.md` is the canonical instruction surface; durable corrections belong in `.memory/corrections/`.
- [ADR-0019 — Governance as code and risk tiers](https://github.com/frostyard/core/blob/main/docs/adr/0019-governance-as-code-and-risk-tiers.md) — the PR template requires a risk tier and verification evidence.
- [ADR-0021 — SHA-pinned actions and least-privilege CI](https://github.com/frostyard/core/blob/main/docs/adr/0021-sha-pinned-actions-and-least-privilege-ci.md) — actions must use commit SHA pins, minimal permissions, and checkouts with `persist-credentials: false`.
- [ADR-0025 — Consolidate repository docs into docs/](https://github.com/frostyard/core/blob/main/docs/adr/0025-consolidate-repository-docs-into-docs.md) — maintainer docs use the 4 categories and the [documentation index](README.md).
- [ADR-0029 — ACMM conformance via canonical aliases](https://github.com/frostyard/core/blob/main/docs/adr/0029-acmm-conformance-via-canonical-aliases.md) — agent-specific instruction files must be relative symlinks to `AGENTS.md`, not independent copies.
- [ADR-0040 — Publish the repository settings contract](https://github.com/frostyard/core/blob/main/docs/adr/0040-publish-the-repository-settings-contract.md) — repository settings follow the contract; [SECURITY.md](../SECURITY.md) directs reports to GitHub private vulnerability reporting.
- [ADR-0043 — Pin repository tools in mise and name the verify gate](https://github.com/frostyard/core/blob/main/docs/adr/0043-pin-repository-tools-in-mise-and-name-the-verify-gate.md) — repository tools must be pinned in `mise.toml` and `mise.lock`; `make verify` must change no file.
- [ADR-0044 — Expose the make gate triad in every repository](https://github.com/frostyard/core/blob/main/docs/adr/0044-expose-the-make-gate-triad-in-every-repository.md) — expose `make verify`, `make check`, and `make ci`, with `ci` running `verify`.
- [ADR-0053 — Name the svu config .svu.yml](https://github.com/frostyard/core/blob/main/docs/adr/0053-name-the-svu-config-svu-yml.md) — the version configuration must use `.svu.yml`, not `.svu.yaml`.
- [ADR-0068 — Label pull requests that touch high-risk paths](https://github.com/frostyard/core/blob/main/docs/adr/0068-label-pull-requests-that-touch-high-risk-paths.md) — risk paths must cover release, deployment, and privileged installation surfaces; matching PRs retain `risk:high`.
- [ADR-0070 — Build tool docs sites with MaterialX on GitHub Pages](https://github.com/frostyard/core/blob/main/docs/adr/0070-build-tool-docs-sites-with-materialx-on-github-pages.md) — `site/` uses ProperDocs, MaterialX, hashed requirements, strict PR builds, and deployment from `main`.
- [ADR-0071 — Hold every code repository to one baseline](https://github.com/frostyard/core/blob/main/docs/adr/0071-hold-every-code-repository-to-one-baseline.md) — mwb adopts rows B1 to B19, with the account and publication exceptions in [ADR-0001](adr/0001-adopt-the-frostyard-repository-baseline.md).
- [ADR-0072 — Write every README to one standard](https://github.com/frostyard/core/blob/main/docs/adr/0072-write-every-readme-to-one-standard.md) — the root README follows the `tool` kind and rules R1 to R12.
- [ADR-0073 — Let CI run the make ci targets as separate jobs](https://github.com/frostyard/core/blob/main/docs/adr/0073-let-ci-run-the-make-ci-targets-as-separate-jobs.md) — separate CI jobs must collectively run every prerequisite and recipe command of `make ci`.
- [ADR-0075 — Batch routine pull requests](https://github.com/frostyard/core/blob/main/docs/adr/0075-batch-routine-pull-requests.md) — group routine dependency updates per ecosystem and batch routine maintenance work.
- [ADR-0076 — Require status checks on every default branch](https://github.com/frostyard/core/blob/main/docs/adr/0076-require-status-checks-on-every-default-branch.md) — the default branch must require the acceptance checks, with stable names and no PR path filter that leaves them pending.
- [ADR-0077 — Bind ready-for-human to the head commit](https://github.com/frostyard/core/blob/main/docs/adr/0077-bind-ready-for-human-to-the-head-commit.md) — readiness applies only to the checked PR head; a changed head or failing CI invalidates it.
- [ADR-0078 — Sort pull requests by levels L1 to L4](https://github.com/frostyard/core/blob/main/docs/adr/0078-sort-pull-requests-by-levels-l1-to-l4.md) — risk-path entries must declare levels for PR triage; L1 and L2 retain `risk:high`.
- [ADR-0081 — Store each correction in a file of its own](https://github.com/frostyard/core/blob/main/docs/adr/0081-store-each-correction-in-a-file-of-its-own.md) — each correction is a canonical JSON file in [`.memory/corrections/`](../.memory/README.md), never removed or rewritten except to fill an empty `promoted_to` once.
- [ADR-0083 — Run the CI gate locally as make ci](https://github.com/frostyard/core/blob/main/docs/adr/0083-run-the-ci-gate-locally-as-make-ci.md) — run `make ci` before a push; local CI must include PR checks except checks that need network access or credentials when they run.
- [ADR-0084 — Accept a tool's own site or a frostyard.org guide as its docs home](https://github.com/frostyard/core/blob/main/docs/adr/0084-accept-a-tools-own-site-or-a-frostyard-org-guide-as-its-docs-home.md) — mwb chooses its [own user site](https://bketelsen.github.io/mwb/), linked from the root README.

## Not applicable

- [ADR-0009 — Single artifact origin repository.frostyard.org](https://github.com/frostyard/core/blob/main/docs/adr/0009-single-artifact-origin-repository-frostyard-org.md) — mwb is a bketelsen repository; it publishes on its own GitHub releases and through the `bketelsen/homebrew-tap` cask, not through repository.frostyard.org.
- [ADR-0011 — Frostyard-prefixed package names](https://github.com/frostyard/core/blob/main/docs/adr/0011-frostyard-prefixed-package-names.md) — the bketelsen repository keeps the package name `mwb`.
- [ADR-0055 — Publish Debian packages through the apt-publisher](https://github.com/frostyard/core/blob/main/docs/adr/0055-publish-debian-packages-through-the-apt-publisher.md) — mwb publishes its deb/rpm packages on its own GitHub releases, not through repository.frostyard.org.
- [ADR-0042 — Adopt a merge queue on the default branch](https://github.com/frostyard/core/blob/main/docs/adr/0042-adopt-a-merge-queue-on-the-default-branch.md) — a merge queue needs an organization-owned repository; mwb is in the bketelsen account.

Core's organization authority declares frostyard repositories. mwb has no
core declaration and receives no core skills-sync PRs.

When changing behavior covered by a binding core ADR, update or supersede
the ADR in frostyard/core first, then change this repository in the same effort.
