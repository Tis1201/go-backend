# 1. Use Gin as the HTTP framework

## Status

Accepted, 2026-11-01. Replaces the earlier plan to stay on the standard library.

## Context

The project's goal is to learn to build a production-grade Go API. The owner wants that
experience with Gin, the framework that appears most often in Go job postings around
them.

At decision time the API has two endpoints (`/healthz`, `/readyz`) on the standard
library router. Upcoming steps need route groups, middleware, JSON body binding and
input validation, which the standard library leaves to us to build.

## Options

- **Standard library `net/http`** (Go 1.22+ patterns). No dependencies, every Go library
  works with it. We would write binding, validation and route grouping ourselves.
- **Chi.** Small router on top of `net/http`; handlers stay `http.Handler`, so leaving it
  later is cheap. No binding or validation.
- **Gin.** Router, `*gin.Context`, middleware chain, binding and validation through
  `go-playground/validator`. The most used Go web framework. Handlers depend on
  `*gin.Context`, and it adds about 29 modules to `go.mod`.
- **Fiber.** Express-like API, but built on fasthttp instead of `net/http`, so it cannot
  use `httptest`, standard middleware or OpenTelemetry's `net/http` instrumentation.
  Rejected.

## Decision

Use Gin, served through our own `http.Server`.

`*gin.Engine` implements `http.Handler`, so `httpapi.NewHandler` keeps returning
`http.Handler`, `main` keeps the server with its timeouts, and tests keep using
`httptest`.

## Consequences

Good:

- Binding, validation, route groups and a middleware chain without writing them.
- Skills that transfer directly to Gin codebases at work.
- The existing tests ran unchanged against the Gin version and caught two behavior
  differences (405 handling, `Content-Type` charset). They keep protecting the API
  contract.

Costs, and how we handle them (enforced in the "Gin rules" section of `AGENTS.md`):

- Gin's defaults are demo-friendly, not production-safe: `r.Run()` has no timeouts,
  wrong methods return 404, every proxy is trusted, `gin.Default()` logs in Gin's own
  text format, and `c` does not carry request cancellation. We build with `gin.New()`
  and set each of these explicitly.
- Handlers are tied to `*gin.Context`. To keep a later move possible, only handlers and
  the router import Gin; services and stores below the handlers do not.
- More dependencies to keep up to date.

Revisit if Gin stops being maintained, or if its dependency weight or defaults cause
real problems.
