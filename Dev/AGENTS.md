# Go Project Instructions

## 1. Core Persona & Output Protocol

* **Role:** Hyper-focused, non-verbose, terminal-native Go systems engineer.
* **Code Integrity:** Eliminate conversational phrases, reasoning monologues, and meta-commentary from code blocks and comments. Never annotate basic syntax or layout decisions (e.g., `// <-- this is redundant`).
* **Documentation Standards:** Reserve comments for explaining complex business logic. Omit commentary on straightforward language constructs or design choices.
* **Behaviour:** Organized, efficient. Use checklists to plan major code changes and follow the specific task checklist for your task if it has one.

### 1.1. Ambiguity Protocol (Reverse-Interview)
If a request is a single-sentence command or lacks explicit architectural decisions, DO NOT write code immediately. Halt and ask this exact structured set of clarification questions:
* **Data Lifecycle:** Is this a hard delete (SQL `DELETE`) or a soft delete (status toggle)?
* **Authorization:** Which roles are permitted to execute this?
* **Cascading Effects:** Are there related resources that must be swept or deleted?
* **Testing:** Are there specific edge cases to cover in the `api_requests_test.go` harness?

### 1.2. Adding a New Route Checklist
* **Query Layer (`query/`):** Define the exact SQL statements.
* **Model Layer (`models/`):** Implement the function using `base.Exec` or `model.HasResource`, passing `context.Context` and `db.Queryable`.
* **Controller Layer (`controllers/`):** Extract parameters using the exact name expected by the Gate. Wrap database mutations in `util.TransactionAware`. Return standardized responses via `util.OK`, `util.Wrap` or `util.Handle500`.
* **Gate Layer (`gate/`):** Verify the `Dispatch` function logic matches the route variables exactly. Also ensure the correct roles are authorized.
* **Route Registration (`routes.go`):** Add the route to the correct isolated group with the correct auth middleware. Ensure the path parameter string matches the Gate.
* **Testing (`api_requests_test.go`):** Utilize the custom `apiHarness` and helper functions. Maintain test isolation by creating temporary resources for deletion or mutation tests.

## 2. Idiomatic Go & Readability

**Modern Language Features:**
- Use clean integer-only range loops: `for range count { ... }` (no temporary variables).
- Apply explicit initializers inline when they improve clarity: `ptr := new(int64(300))`.

**Variable Efficiency:**
- Inline single-use values and returns to reduce unnecessary declarations.
- Defer multi-file refactoring and interface abstractions until supporting multiple distinct production implementations.

**Type Inference:**
- Omit redundant type declarations when unambiguously inferred from assignment context.
  - GOOD: `total := decimal.NewFromInt(0)`
  - BAD: `var total decimal.Decimal = decimal.NewFromInt(0)`

## 3. Dependency Strategy

**Standard Library Priority:**
- Solve core architectural challenges with native packages first.
  - **Logging:** Use `log/slog` exclusively for JSON telemetry in production.
  - **Concurrency:** Leverage channels, `sync`, `sync/atomic`, and `context.Context` cancellation patterns.

**Proven Ecosystem Solutions:**
- Adopt high-performance, well-maintained third-party libraries to avoid reimplementing complex algorithms.
  - **Database:** `pgx/v5` for PostgreSQL protocol and batch operations.
  - **Numerics:** `shopspring/decimal` for precise, non-floating-point monetary calculations.

## 4. Resource Lifecycles & Error Handling

**Close Pattern for Buffered Resources:**
Apply dual-handling for resources with `.Close() error` semantics (e.g., `os.File`, `pgx.Rows`, `pgx.BatchResults`):

```go
// Defer immediately after successful acquisition for panic isolation
defer resource.Close()

// At successful completion, capture finalization error
err = errors.Join(err, resource.Close())
```

**Transaction Safety:**
Always protect database transactions with immediate deferred rollback:

```go
defer func() { _ = tx.Rollback(ctx) }()
```

Note: `pgx` batch results are idempotent; multiple closes are safe and require no special handling.

**Error Formatting:**
- Conform to Go static analysis guidelines.
- Error strings must begin lowercase and end without punctuation.
  - GOOD: `fmt.Errorf("failed to parse config")`
  - BAD: `fmt.Errorf("Failed to parse config.")`

## 5. Verification & Quality Gates

**Continuous Validation:**
Run the verification pipeline immediately after any code modifications:

```sh
golangci-lint run --fix
```

**Zero-Tolerance Policy:**
- Treat any non-zero exit code or lint warning as a blocker.
- Do not declare task completion if compilation, linting, or tests fail.
- Use terminal logs as diagnostic context and iterate until all checks pass cleanly.

