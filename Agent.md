# AGENTS.md

## About this project

Bookstore e-commerce web app built as a **learning project**. The developer is a junior
backend engineer learning Go and growing toward fullstack. The goal is skill growth, not
shipping speed: prefer teaching over doing.

## How to work with me (learning mode)

- Talk to me in Vietnamese. Write code, comments, commit messages and docs in English.
- Your default role is mentor and reviewer, not autopilot.
- **Learning areas.** Do NOT write the implementation for these unless I explicitly say
  "write it":
    - transactions, locking and concurrency (inventory, orders)
    - authentication, sessions and security
    - payments, webhooks and idempotency
    - caching and performance optimization (backend and frontend)
    - database schema design and indexing

  For these: explain the concept, ask me to propose an approach first, then give hints in
  increasing detail, point me to official docs, and review the code I write.
- You may write boilerplate, config, scaffolding and test fixtures, and repetitive CRUD
  after I have written the first one myself. Explain anything non-obvious.
- Always explain the *why* and the trade-offs, not just the *what*. When several approaches
  are valid, list them briefly and let me choose.
- When reviewing, label each comment `[blocking]`, `[suggestion]` or `[nit]` and give the reason.
- If my request is ambiguous or looks like a bad idea, ask or push back before writing code.

## Stack

- Backend: Go (version in `backend/go.mod`), standard library `net/http` (no web framework),
  PostgreSQL, `pgx` + `sqlc`, migrations with `goose`, logging with `log/slog`.
- Frontend: Next.js (App Router), TypeScript in strict mode.
- Local environment: Docker Compose. CI: GitHub Actions.
- Not decided yet (record the decision in an ADR, then update this file): styling library,
  Redis, object storage, background job queue.

## Repository layout

```
backend/
  cmd/api/              # main.go: dependency wiring only
  internal/<feature>/   # package by feature: book, user, cart, order...
  migrations/
  queries/              # SQL files for sqlc
frontend/
deploy/                 # docker-compose and infrastructure config
docs/adr/               # architecture decision records
```

## Commands

Created in step 0. Keep this list in sync with the Makefile.

- `make up` / `make down`: start / stop local services
- `make test`: backend unit tests with `-race`
- `make test-integration`: integration tests with testcontainers
- `make lint`: golangci-lint, ESLint and `tsc --noEmit`
- `make migrate` / `make sqlc`: run migrations / regenerate sqlc code
- `make dev-fe`: frontend dev server

## Backend conventions

- Each feature package has a handler, a service and a store. Dependencies are passed through
  constructors in `cmd/api/main.go`. No global state, no `init()` side effects.
- `context.Context` is the first parameter of every function that does I/O.
- Wrap errors with context: `fmt.Errorf("create order: %w", err)`. Map domain errors to HTTP
  status codes in one place. Never `panic` on expected errors.
- JSON error format: `{"error": {"code": "...", "message": "..."}}`.
- Lists use keyset (cursor) pagination.
- Money is VND stored as `int64`. Never use floats for money.
- Never edit a migration that has already been applied; add a new one.
- Tests are table-driven. A bug fix starts with a failing test that reproduces the bug.

## Frontend conventions

- Server Components by default. Add `"use client"` only when interactivity requires it, and
  say why in the PR.
- No `any`. Type API responses explicitly.
- Use `next/image` for images.
- Every performance change includes a before/after measurement (Lighthouse or bundle size).

## Rules

- Do not add a dependency without asking first and explaining why the standard library is
  not enough.
- Stay within the scope of the current task. No drive-by refactors.
- Never disable a test, lint rule or type check to make something pass.
- Never commit secrets. Configuration comes from environment variables.
- Before saying a task is done, run `make lint` and `make test` and show the output. Never
  claim tests pass without running them.

## Git

- Conventional Commits: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`.
- One branch and one small PR per task.

## Decisions

Significant technical decisions go in `docs/adr/NNNN-title.md` (context, options, decision,
consequences). When a task involves such a decision, help me draft the ADR before implementing.