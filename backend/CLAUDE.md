# Backend Development

## Command: "run task XX"

When the user says **"run task XX"** (where XX is a number 01–10):

1. **Read the agent prompt:** `backend/AGENT.md`
2. **Read the backlog status:** `backend/status.md`
3. **Read the story file:** `backend/story-XX-*.md` (glob match on the number)
4. **Check prerequisites:** The story's Prerequisites section lists which stories must be done first. Cross-reference with `status.md` — if any prerequisite is not marked `done`, stop and tell the user which tasks must be completed first.
5. **Read the full spec** at `docs/server-mock.md` for any details the story references.
6. **Implement** everything in the story: production code, test code, in the exact file paths listed.
7. **Verify:** Run `go build ./...` and `go test ./... -v` from inside `backend/`.
8. **Update status:** Edit `backend/status.md` — change the task's status from `pending` to `done`.

## Rules

- Work entirely inside `backend/`. Never modify files outside it.
- Follow the coding standards in `AGENT.md` exactly.
- Do not add dependencies beyond what `AGENT.md` lists.
- Do not skip tests — every story includes test code that must be written and must pass.
- Do not modify code from completed stories unless the current story explicitly says to (e.g., Story 07 extends `app_handler.go`).
- After finishing, show the user the `go test` output and the updated `status.md`.
