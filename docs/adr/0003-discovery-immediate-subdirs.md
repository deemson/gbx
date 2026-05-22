# 3. Discovery scans immediate subdirectories of one root

Status: Accepted
Date: 2026-05-22

## Context

gbx needs to know which repos to list. Options range from a configured list, to
a recursive walk of a tree, to a flat scan of one directory.

## Decision

Scan the **immediate** subdirectories of a single root directory; each one that
is a git repository becomes a row, and non-repos are ignored. The root is a CLI
argument, defaulting to the current working directory. No config file, and no
recursive descent.

## Consequences

- Zero configuration: point gbx at a directory of checkouts and it works.
- Predictable, fast discovery — one level of `readdir`, no deep walk.
- Repos nested more than one level down are not found. Accepted for v1; a
  recursive or configured mode is explicitly deferred (see `DECISIONS.md`).
