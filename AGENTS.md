# AGENTS.md — Division of Labor

This document defines the roles and responsibilities of each agent involved in implementing hashtray-Go.

---

## Claude Code (Planner & Orchestrator)

**Role**: Architect, planner, and quality gate.

### Responsibilities

1. **Plan** — Analyse requirements, read the existing Python source, and produce a detailed implementation plan with file-by-file specifications. Get user approval before handing off to Codex.
2. **Orchestrate** — Break the plan into discrete implementation tasks and delegate each to Codex.
3. **Verify** — After Codex completes implementation, run `make test` (or `go test ./...`) to confirm correctness.
4. **Request Review** — After tests pass, invoke Codex for adversarial review.
5. **Iterate** — If adversarial review surfaces issues, fix or re-delegate, re-test, and re-review until clean.
6. **Commit** — Only commit after tests pass and adversarial review is clean.

### Rules

- Never skip the planning step.
- Never commit code that has not passed tests AND adversarial review.
- Keep plans concise and actionable — one plan per implementation phase.

---

## Codex (Implementer & Adversarial Reviewer)

**Role**: Code writer and critical reviewer.

### Implementation Mode

- Receive a task specification from Claude Code.
- Write idiomatic Go code following the conventions in CLAUDE.md.
- Include unit tests for every exported function.
- Keep changes focused on the task — no drive-by refactors.

### Adversarial Review Mode

- Review the full diff or specified files with a critical eye.
- Check for: correctness bugs, edge cases, security issues (injection, path traversal), performance problems, deviation from the Python original's behaviour, missing tests.
- Report issues as a numbered list with file, line, severity (critical/major/minor), and suggested fix.
- If no issues found, explicitly state "No issues found."

---

## Workflow Sequence

```
1. Claude Code  →  creates implementation plan  →  user approves
2. Claude Code  →  delegates task to Codex      →  Codex implements
3. Claude Code  →  runs `make test`             →  fix if failing
4. Claude Code  →  delegates adversarial review →  Codex reviews
5. If issues    →  fix → re-test → re-review (loop back to step 3)
6. If clean     →  Claude Code commits
```

---

## Escalation

- If Codex is unavailable, Claude Code performs both implementation and self-review but still follows the plan → implement → test → review cycle.
- If a task is ambiguous, Claude Code asks the user rather than guessing.
