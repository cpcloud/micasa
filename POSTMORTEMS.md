<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# Postmortems

Real examples of agent failure patterns in this repo. Read these before
attempting multi-iteration fixes.

---

## Cancellation bug: 14 fix commits for a 2-line root cause

**Feature**: ctrl+c cancellation of streaming LLM responses in the chat overlay.

**Symptom**: Pressing ctrl+c during SQL streaming produced a spurious "LLM
returned empty SQL" error, left a frozen spinner, and failed to show an
"Interrupted" indicator.

**What went wrong**: The agent (Claude 4.5 Sonnet) spent 14 commits adding
increasingly elaborate workarounds:

1. Added a `Cancelled` flag to chatState, checked it in multiple handlers.
2. Added special-case error suppression when the flag was set.
3. Added a second cancellation handler that duplicated the first.
4. Added flag-clearing logic in multiple places, creating ordering bugs.
5. Kept patching symptoms as each new flag interaction broke something else.

Each "fix" passed the specific scenario the agent was looking at but broke
another path. The test assertions checked internal state mutations rather
than observable behavior, so they kept passing even when the UI was broken.

**Root cause**: `waitForSQLChunk` and `waitForChunk` synthesized a
`Done: true` message when the stream channel closed (due to cancellation).
This fake "done" message had empty content, which downstream handlers
treated as a real completion with no SQL -- triggering the error.

**Actual fix**: Return `nil` (no Bubble Tea message) when the channel
closes instead of synthesizing a Done message. Two lines changed, zero
flags added. The message loop simply stops, and the cancellation handler
cleans up the UI state.

**Lessons**:

- When a fix doesn't work on the second try, the mental model of the bug
  is wrong. Stop patching and re-read the full code path.
- Flags and special cases are a smell. If you need a `Cancelled` bool to
  suppress errors, the errors shouldn't be generated in the first place.
- Test observable output (rendered UI), not internal state. The spinner
  was visibly frozen but all state-based tests passed.
- Concurrency bugs in message-passing systems are almost always about
  *what messages get sent*, not about *what flags are set when they arrive*.
