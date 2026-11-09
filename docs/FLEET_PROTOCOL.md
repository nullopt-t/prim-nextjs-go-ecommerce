# Fleet Multi-Agent Protocol & Completion Contract

When working as part of a multi-agent fleet across Herdr panes and workspaces, every agent MUST follow this exact contract from turn 1.

## 1. The Strict Completion Handshake (MANDATORY)

Never finish work silently. When your task, audit, or code modification is complete:

1. **Verify All Quality Gates**:
   - For Backend: `cd backend && go test ./...` and `go vet ./...` MUST be clean.
   - For Frontend: `cd frontend && npx tsc --noEmit` and `npx vitest run` MUST be clean (0 errors, 100% tests passing).
   - If UI text was changed/added: confirm both `frontend/messages/en.json` AND `frontend/messages/ar.json` have identical keys.

2. **Explicitly Notify the Lead Agent**:
   Do NOT simply stop writing. Use the Herdr CLI to report your exact completion to the Lead orchestrator (`w9:p2`):
   ```bash
   herdr pane send-text w9:p2 "[<your-agent-name>] TASK DONE: <concise summary of findings, fixes, and verification state>"
   herdr pane send-keys w9:p2 "Enter"
   ```
   Or write your final response beginning with `[<your-agent-name>] TASK DONE: ...` so the orchestrator hook immediately catches it.

## 2. Engineering Discipline & Privacy Rules

- **Zero AI Attribution / Privacy Contract**:
  - Never add `Co-authored-by`, `Signed-off-by`, or assistant metadata to git commits or code.
  - All commits belong exclusively to `Amr Ali`.
  - Never add bot watermarks, emojis, or AI comments in code or markdown.

- **Ponytail Discipline**:
  - The smallest, cleanest diff that completely solves the problem.
  - No unnecessary abstractions, no sprawling refactors, no premature generalizations.
  - Fix the exact bug, add the unit test to prove the fix, and stop.

- **Logical Bugs over Cosmetic Checks**:
  - Focus heavily on bugs that return HTTP 200 but have corrupted domain logic:
    - Minor-unit currency rounding drift (cents/piastres vs display floats).
    - Unintended state retention (e.g. invalid coupon not clearing applied discount).
    - Concurrency & race conditions on stock reservations / inventory limits.
    - SSR hydration mismatches and URL query parameter drops.
    - Cumulative limit bypasses (e.g. requesting returns for more items than purchased).
