# 0001 — Adopt the frostyard repository baseline

- **Status:** Accepted
- **Date:** 2026-11-17

## Context

mwb is a Go client for Microsoft PowerToys Mouse Without Borders, hosted in
the bketelsen account. It ships Linux and macOS binaries, Linux deb/rpm
packages, and a cask in `bketelsen/homebrew-tap`.

Core's repository baseline defines a common instruction surface,
corrections inbox, documentation structure, tool and CI gates, and release
and review conventions. Its organization authority declares frostyard
repositories, not repositories in personal accounts.

## Decision

mwb follows rows B1 to B19 of core's
[repository baseline](https://github.com/frostyard/core/blob/main/docs/specs/repository-baseline.md),
although it is in the bketelsen account. The
[binding core ADR index](../org-adrs.md) records the requirements.

The account and publication exceptions are:

- **No core declaration and no skills sync.** Core's organization authority
  declares frostyard repositories. mwb remains outside that declaration and
  its skills-sync distribution.
- **No merge queue.** Core ADR-0042 requires an organization-owned repository.
  mwb is in the bketelsen account.
- **Keep the `mwb` package name and existing publication routes.** Core
  ADR-0009, ADR-0011, and ADR-0055 do not apply. The bketelsen repository
  publishes to its own GitHub releases and the `bketelsen/homebrew-tap` cask,
  not through repository.frostyard.org.

The user documentation home is the ProperDocs and MaterialX site in
[`site/`](../../site/), hosted on GitHub Pages at
https://bketelsen.github.io/mwb/. Maintainer documentation uses core's
4-category `docs/` shape. The 6 finished plans stay as historical records.

Adoption defines the requirements. It does not assert that every baseline
file or repository setting is already in place.

## Consequences

- Contributors use the same documentation, correction, CI, and review
  conventions as frostyard repositories.
- Baseline checks can identify gaps without inventing a separate mwb standard.
- mwb must maintain its own implementation of the baseline and its site.
  It does not receive organization declaration or skills-sync changes.
- Publication stays compatible with the `mwb` package name and existing tap.
- Repository settings, including Pages and required checks, still need
  configuration by an account with the required permissions.

## Alternatives considered

- **Keep unrelated repository conventions:** this leaves avoidable differences
  in contributor instructions, verification, and documentation.
- **Move mwb to frostyard:** an account move is not needed to adopt the
  baseline and would change repository and publication addresses.
- **Adopt every organization publication rule:** this would change the
  package namespace and artifact origin without a client requirement.

## References

- Shapes: [documentation index](../README.md), [user site](../../site/),
  [corrections inbox](../../.memory/README.md)
- Contracts: [core repository baseline](https://github.com/frostyard/core/blob/main/docs/specs/repository-baseline.md),
  [core README spec](https://github.com/frostyard/core/blob/main/docs/specs/readme.md)
- Builds on: [core ADR-0071](https://github.com/frostyard/core/blob/main/docs/adr/0071-hold-every-code-repository-to-one-baseline.md),
  [core ADR-0025](https://github.com/frostyard/core/blob/main/docs/adr/0025-consolidate-repository-docs-into-docs.md),
  [core ADR-0070](https://github.com/frostyard/core/blob/main/docs/adr/0070-build-tool-docs-sites-with-materialx-on-github-pages.md)
- Exceptions: [core ADR-0009](https://github.com/frostyard/core/blob/main/docs/adr/0009-single-artifact-origin-repository-frostyard-org.md),
  [core ADR-0011](https://github.com/frostyard/core/blob/main/docs/adr/0011-frostyard-prefixed-package-names.md),
  [core ADR-0042](https://github.com/frostyard/core/blob/main/docs/adr/0042-adopt-a-merge-queue-on-the-default-branch.md),
  [core ADR-0055](https://github.com/frostyard/core/blob/main/docs/adr/0055-publish-debian-packages-through-the-apt-publisher.md)
