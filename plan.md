# Implementation Plan

1. Define scope and UX contract
- Build a terminal UI around existing `gwt` command flows: `clone`, `checkout`, `list`, `remove`.
- Keep layout minimal: one main panel, one status/progress area, one footer for key hints.
- Standardize operation states for every action:
  - `running`
  - `success`
  - `failed`
- Decide one status model used everywhere:
  - `id`, `label`, `startedAt`, `endedAt`, `state`, `message`, `progressPercent?`, `isIndeterminate`.

2. Create visual design system (simple + beautiful)
- Use `lipgloss` styles with a small token set:
  - `Neutral` for idle/info
  - `Blue` for running
  - `Green` for success
  - `Red` for failed
- Keep typography clean with spacing and subtle borders (no visual clutter).
- Define reusable style helpers:
  - `StatusBadge(state)`
  - `Panel(title, content)`
  - `MutedText`, `ErrorText`, `SuccessText`
- Ensure color is not the only signal:
  - Include icons/text tags like `[RUNNING]`, `[OK]`, `[FAILED]`.

3. Add operation execution layer with progress events
- Wrap each command flow in an operation runner that emits events:
  - `OperationStarted`
  - `OperationProgress`
  - `OperationCompleted`
  - `OperationFailed`
- Split long tasks into steps and report progress:
  - `clone`: validating input, cloning, configuring bare repo, creating first worktree
  - `checkout`: resolving branch, creating worktree, upstream setup
  - `remove`: safety checks, remove worktree, delete branch
- For unknown duration tasks, use indeterminate spinner; switch to percent bar when measurable.

4. Build Bubble Tea model/update/view structure
- `Model` fields:
  - current command context
  - active operation
  - recent operation history
  - spinner/progress components from `bubbles`
  - viewport width/height for responsive layout
- `Update` handles:
  - keyboard input
  - operation events
  - window resize
  - periodic ticks for spinner/progress animation
- `View` renders:
  - header (repo/context)
  - main content (current command/result list)
  - status/progress strip (always visible)
  - footer (short key hints)

5. Implement progress indicators for long-running tasks
- Use `bubbles/spinner` for indeterminate states.
- Use `bubbles/progress` for determinate states with percentage and current step.
- Add elapsed time display for running tasks.
- On completion/failure, freeze final state briefly and show concise summary.

6. Error handling and resilience
- Normalize errors into user-facing messages with optional detail toggle.
- Keep failed operation history visible so users can inspect what failed.
- Ensure cancellation path (`ctrl+c`) updates state cleanly to failed/cancelled with clear messaging.

7. Command integration strategy
- Start by integrating TUI with `clone` and `checkout` (most complex/long-running).
- Add `remove` and `list` once status/progress pipeline is stable.
- Keep business logic separate from UI so commands can still run non-interactively.

8. Testing and validation
- Unit tests:
  - state transitions (`running -> success/failed`)
  - event-to-view-model mapping
  - style/state badge selection
- Integration tests:
  - simulated long-running command streams
  - failure mid-step and recovery messaging
- Manual UX checks:
  - narrow terminal widths
  - color contrast/readability
  - no flicker under rapid progress updates

9. Rollout phases
- Phase 1: design tokens + shared status model + operation event bus.
- Phase 2: TUI shell with running/success/failed visuals and spinner.
- Phase 3: determinate progress bars and step-level updates for `clone`/`checkout`.
- Phase 4: integrate `remove`/`list`, polish, and docs/screenshots in README.

10. Acceptance criteria
- Every operation clearly shows `running/success/failed` state with color + text.
- Long-running tasks always show a spinner or progress bar.
- Failures are understandable without reading stack traces.
- UI remains minimal, readable, and consistent across commands.
