# UI Specification

## 1. Design goals

- Minimal, focused, status-first UI.
- Color communicates state: running, success, failure.
- Clear step-by-step operation timeline.
- No fullscreen/alternate-screen behavior; output remains in normal terminal scrollback.
- Graceful fallback when not attached to an interactive TTY.

## 2. Libraries used

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — terminal UI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) — UI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — terminal styling

## 3. Runtime mode

The operation UI (`clone`, `checkout`, `remove`) only renders the Bubble Tea interface when terminal is interactive:

- `stdin` and `stdout` must be character devices.
- `/dev/tty` must be openable.

If non-interactive (CI, redirected output, script), run operations without TUI but keep same internal event flow.

## 4. Color and style tokens

Using Lip Gloss color codes:

- `running`: foreground `39`, bold.
- `success`: foreground `42`, bold.
- `failed`: foreground `196`, bold.
- `muted`: foreground `245`.
- `header`: foreground `75`, bold.
- panel border: rounded border, color `240`.
- panel padding: horizontal `1`.

Status badge style:

- base: bold + horizontal padding `1`.
- `RUNNING`: background `39`, foreground `0`, label ` RUNNING `.
- `OK`: background `42`, foreground `0`, label ` OK `.
- `FAILED`: background `196`, foreground `15`, label ` FAILED `.
- `IDLE` (fallback): background `245`, foreground `0`, label ` IDLE `.

## 5. Panel layout

Primary operations panel structure (top to bottom):

1. Status line: `[BADGE]  <status message>`
2. Optional error detail line (failure only): `Error: <full error text>` in failed style
3. Blank line
4. Activity indicator line
5. Elapsed line (`Elapsed: <duration>` in muted style)
6. Step history lines

Footer line under panel:

- `Press q to quit` (muted)

Panel width rule:

- `max(40, min(90, terminalWidth - 4))`

History line width rule (inside panel):

- `max(20, min(70, terminalWidth - 12))`

## 6. Activity indicator behavior

While running:

- show spinner (Bubble `spinner.Dot`) + `Working...` in running style.

On success:

- show `✓ Done` in success style.

On failure:

- show `✗ Failed` in failed style.

## 7. Step history formatting

Each history row:

- Left: `<icon> <message>`
- Dot leader fills up to target width
- Right state label aligned by dot padding

State mapping:

- Running: icon `•`, label `running`, running style
- Success: icon `✓`, label `done`, success style
- Failed: icon `✗`, label `failed`, failed style

Rules:

- Keep only the last 8 entries.
- Ignore empty messages.
- De-duplicate exact consecutive `(message, state)` rows.
- When moving to a new step, finalize previous running step to success.
- On operation completion, finalize current running step to success.
- On failure, finalize current running step to failed.
- Do not append raw error details as a history row.

## 8. Failure presentation rules

On failure:

- Header badge is `FAILED`.
- Status message is generic: `Operation failed`.
- Full error text appears once as a dedicated error detail line.
- Do not print additional duplicate error/usage text after panel.

Command-level behavior to support this:

- Cobra `SilenceUsage: true`
- Cobra `SilenceErrors: true`
- Wrap UI errors as "already displayed" and suppress extra stderr print.

## 9. Animation and pacing

Global flag:

- `--ui-step-delay <duration>`

Behavior:

- Artificial sleep is inserted after each emitted operation event (`started`, `progress`, `completed`, `failed`).
- Useful for demo/review of state transitions.

## 10. Key handling

- `q` or `Ctrl+C` exits UI.
- Exit while running maps to cancellation error.

## 11. Typography and symbols

Use Unicode symbols to match feel:

- Success: `✓`
- Failure: `✗`
- Running/Bare bullet: `•`

Avoid ASCII fallbacks unless terminal/font cannot render Unicode.

## 12. Implementation checklist for replication

1. Create state palette and badge renderer with exact colors.
2. Build bordered panel container with same width math.
3. Implement event-driven state machine (`started/progress/completed/failed`).
4. Add spinner-driven running indicator and terminal elapsed timer.
5. Implement history list with dot leaders and row finalization rules.
6. Add failure detail line and suppress duplicate post-panel errors.
7. Add non-interactive fallback path.
8. Add `--ui-step-delay` and apply delay in event emitter.
9. Implement styled static list panel with `--all` behavior.
