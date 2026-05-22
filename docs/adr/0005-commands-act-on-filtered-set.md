# 5. Commands act on the filtered set; no multi-select

Status: Accepted
Date: 2026-05-22

## Context

To run a command across several repos, the user must express *which* repos. The
common pattern is marking/multi-select (toggle rows, then act). That adds
selection state, a marking binding, and a visual column.

## Decision

A command acts on **the repos currently matching the filter**. There is **no
marking and no multi-select**. Clearing the filter targets all repos; narrowing
it narrows the target set.

The drill-in (ADR 0009) is the one place that needs a single repo, so it — and
only it — uses a cursor over the filtered list. The cursor selects *what to
inspect*, not a target set for commands.

## Consequences

- "Select the repos" and "filter the view" are the same act — less state, fewer
  keys.
- Running a command is `filter → ctrl+<key>`; the targets are exactly what is on
  screen, which makes the blast radius visible before acting (reinforces 0006).
- You cannot act on an arbitrary, non-contiguous hand-picked subset in one go;
  you express it as a filter instead. Accepted.
