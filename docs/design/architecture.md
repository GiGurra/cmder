# Architecture

How cmder works internally.

## Overview

```
┌─────────────────────────────────────────────────────────────┐
│                          Spec                               │
│  App, Args, Timeouts, Retries, I/O config                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        Run(ctx)                             │
│  Main entry point                                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     withRetries()                           │
│  Retry loop with total timeout                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Attempt execution                         │
│  exec.CommandContext + attempt timeout                      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        Result                               │
│  StdOut, StdErr, Combined, Err, Attempts, ExitCode          │
└─────────────────────────────────────────────────────────────┘
```

## Key Components

### Spec (Immutable Configuration)

The `Spec` struct holds all configuration. Builder methods return new instances:

```go
func (c Spec) WithRetries(n int) Spec {
    c.Retries = n  // Copies struct
    return c       // Returns copy
}
```

This enables:
- Safe reuse of templates
- Thread-safe configuration
- Method chaining

### Timeout Management

Two timeout layers:

1. **Total timeout** - Wraps the entire retry loop
2. **Attempt timeout** - Per-execution limit

```go
// Total timeout context
jobCtx, cancelJobCtx := context.WithCancel(parentCtx)
if c.TotalTimeout > 0 {
    executeAfterDuration(jobCtx, c.TotalTimeout, func() {
        jobTimedOut.Store(true)
        cancelJobCtx()
    })
}

// Attempt timeout (inside retry loop)
attemptCtx, cancelAttemptCtx := context.WithCancel(jobCtx)
```

### Reset on Output

The "keep alive" feature works via a signal channel:

```go
// Writer that signals on write
sfw := util.NewSignalForwarderWriter(aliveChannel)

// Monitor goroutine
go func() {
    for {
        select {
        case <-aliveSignal:
            if c.ResetAttemptTimeoutOnOutput {
                attemptDeadline.Store(toPtr(time.Now().Add(c.AttemptTimeout)))
            }
        case <-attemptCtx.Done():
            return
        }
    }
}()
```

Each write to stdout/stderr sends a signal, resetting the deadline.

### Output Collection

Output goes through `io.MultiWriter`:

```go
stdOutTargets := []io.Writer{sfw}  // Signal forwarder

if c.CollectAllOutput {
    stdOutTargets = append(stdOutTargets, stdoutBuffer, combinedBuffer)
}

if c.StdOut != nil {
    stdOutTargets = append(stdOutTargets, c.StdOut)
}

cmd.Stdout = io.MultiWriter(stdOutTargets...)
```

This allows simultaneous:
- Timeout reset signaling
- Buffer collection
- User-provided writers

### Retry Logic

The retry loop:

```go
for i := 0; i <= c.Retries; i++ {
    err := attemptExecution()

    if err != nil {
        if c.RetryFilter(err, attemptTimedOut.Load()) {
            continue  // Retry
        }
        return err  // Give up
    }
    return nil  // Success
}
return maxRetriesError
```

## Concurrency

cmder is designed for concurrent use:

- `Spec` is immutable (safe to share)
- Each `Run()` call is independent
- Internal state uses `atomic` operations

```go
// Safe: shared spec
baseSpec := cmder.New("cmd").WithRetries(3)

// Concurrent runs
go baseSpec.WithArgs("arg1").Run(ctx)
go baseSpec.WithArgs("arg2").Run(ctx)
```

## Error Handling

Errors are wrapped with context:

```go
return fmt.Errorf("error running cmd %s \n %s: %w", c.App, err.Error(), err)
```

This preserves:
- Original error (for `errors.Is`)
- Command name (for debugging)
- Error message (for logging)

## Dependencies

cmder uses only standard library:
- `os/exec` - Command execution
- `context` - Cancellation
- `sync/atomic` - Thread-safe state
- `io` - Output handling
