# Note: Go performance improvements for gbx

The slowdown with many repositories is likely a combination of unbounded Git subprocess fan-out and repeated full-list rendering. Startup currently uses roughly one discovery process plus four metadata processes per repository: `rev-parse`, status, diff, describe, and branch listing.

## Suggested improvements, in priority order

### 1. Bound Git concurrency

`tea.Batch` currently launches discovery and metadata commands concurrently without a limit. Add one shared scheduler or semaphore for every typed Git operation.

Suggested starting limits:

- local reads: 8 workers;
- fetch/pull: 4 workers.

Benchmark limits of 4, 8, and 16. Bounded concurrency should reduce process storms and disk contention while still keeping the machine busy.

### 2. Reduce startup subprocesses

- Check for a `.git` directory or file before calling `git.Open`; still validate candidates through the existing Git wrapper.
- Load local branch lists only when the switch prompt opens.
- Load status first because it provides almost all primary row information.
- Defer diff and describe until after the first screen is usable.
- Prioritize visible repositories before off-screen repositories.

This should preserve eventual row content while improving perceived startup time.

### 3. Coalesce load messages

Each repository currently produces four completion messages and therefore up to four model updates and renders. Prefer a staged pipeline:

```text
discovered
    ↓
status loaded → first useful row
    ↓
diff + describe loaded → one supplementary update
```

Load branch lists separately and lazily. This reduces message and redraw volume substantially.

### 4. Render only visible repository rows

`listContent()` currently formats every matching row and clips afterward. Instead:

1. compute which visual rows intersect the viewport;
2. render only those repository rows;
3. add headings and scroll markers as required.

This matters particularly while command results and spinner ticks are arriving.

### 5. Cache derived display state

Avoid recomputing these on every render:

- column widths;
- branch color rankings;
- parsed filter terms;
- matched/ranked indexes;
- visual-line mappings.

Invalidate each cache only when its inputs change. Spinner ticks, for example, should not recalculate column widths or filter rankings.

### 6. Avoid repeated full-list searches

Methods such as `setStatusRef`, `loadDoneRef`, and `repoByRef` linearly search `m.repos`. Maintain a `map[repoRef]int` and update it whenever repository order changes.

Also calculate `matchedIndexes()` once per model transition or render rather than repeatedly through `clampView`, `cursorIndex`, `visualLineCount`, and `listContent`.

### 7. Sort discovery results once

`addRepoTo` sorts the complete repository slice after each discovered repository. Either:

- collect results per section and sort once when discovery completes; or
- insert each result directly at its sorted position.

Collecting and sorting once is simplest, although it changes how rows stream into view.

### 8. Reduce spinner redraw cost

The shared spinner causes repeated full-screen model rendering while anything is busy. Options include:

- lowering its tick frequency;
- using a static busy glyph;
- animating only while a visible row is busy;
- relying on visible-row-only rendering to make ticks cheap.

### 9. Refresh only affected metadata

The current post-command path reloads status, diff, describe, and branches every time. A narrower policy could be:

| Operation | Refresh |
| --- | --- |
| Fetch | status; possibly describe |
| Pull | status, diff, describe, branches |
| Switch/create branch | all |
| Arbitrary configured action | all |
| Manual refresh | all |

Fetch does not normally change local branches or the working-tree diff.

### 10. Add cancellation and timeouts

Replace `context.Background()` with contexts owned by the running program:

- cancel outstanding work when gbx exits;
- optionally time out hung Git operations;
- prevent stale results from an older load cycle overwriting newer state.

## Recommended implementation sequence

1. Add startup benchmarks and instrumentation for 10, 50, and 200 repositories.
2. Bound Git concurrency.
3. Lazy-load branches and stage status before diff/describe.
4. Render only visible rows.
5. Cache display calculations and add the `repoRef` index.
6. Coalesce load results and narrow post-command refreshes.

The first four changes should provide most of the improvement without changing Git semantics or rewriting the application.
