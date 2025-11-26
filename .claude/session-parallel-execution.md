# Session: Parallel Validator Execution

Date: 2024-11-26

## What Was Implemented

Added parallel validator execution with category-specific worker pools.

## Key Patterns Discovered

### ValidatorCategory System

Validators categorized by workload type:

- **CategoryCPU** (default): Pure computation (regex, parsing)
- **CategoryIO**: External processes (shellcheck, terraform, actionlint)
- **CategoryGit**: Git operations - serialized to avoid index lock contention

### Executor Pattern

`internal/dispatcher/executor.go`:

- **SequentialExecutor**: Default, runs validators in order
- **ParallelExecutor**: Semaphore-based pools per category

Pool sizes: CPU=NumCPU, IO=NumCPU*2, Git=1

### Adding Category to Validators

```go
func (*MyValidator) Category() validator.ValidatorCategory {
    return validator.CategoryIO
}
```

## Testing Concurrent Code

### Race Detection

```bash
go test -race ./...
```

### Patterns Used

1. `golang.org/x/sync/semaphore` for bounded concurrency
2. `sync.Mutex` to protect shared result slices
3. `context.WithTimeout` to detect deadlocks in tests
4. `context.WithCancel` to test early termination

### Deadlock Test Pattern

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
result := executor.Execute(ctx, hookCtx, validators)
// Would timeout if deadlocked
```

## Go 1.22+ Integer Range

```go
// Use this
for i := range 10 { ... }

// Not this
for i := 0; i < 10; i++ { ... }
```

## Resources

- [Go Race Detector](https://go.dev/doc/articles/race_detector)
- [uber-go/goleak](https://github.com/uber-go/goleak)
