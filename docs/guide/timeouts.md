# Timeouts

cmder supports two types of timeouts: attempt timeout and total timeout.

## Attempt Timeout

Limits how long a single execution can run:

```go
result := cmder.New("slow-query").
    WithAttemptTimeout(30 * time.Second).
    Run(ctx)
```

If the command exceeds this duration, it's killed and (by default) retried.

## Total Timeout

Limits total time including all retry attempts:

```go
result := cmder.New("flaky-service").
    WithTotalTimeout(2 * time.Minute).
    WithRetries(5).
    Run(ctx)
```

Even with 5 retries, the entire operation stops after 2 minutes.

## Both Timeouts

Combine for precise control:

```go
result := cmder.New("api-call").
    WithAttemptTimeout(10 * time.Second).  // Each attempt: max 10s
    WithTotalTimeout(45 * time.Second).    // All attempts: max 45s
    WithRetries(5).
    Run(ctx)
```

## Reset on Output

For long-running commands that produce periodic output:

```go
result := cmder.New("build.sh").
    WithAttemptTimeout(30 * time.Second).
    WithResetAttemptTimeoutOnOutput(true).
    Run(ctx)
```

The timeout resets each time the command writes to stdout or stderr. This is useful for:

- Build scripts that log progress
- Data processing that reports status
- Any command that "checks in" periodically

### Example: Build Script

A build script that takes 5 minutes but logs every 10 seconds:

```go
// Without reset: would timeout after 30s
// With reset: succeeds because timeout resets on each log line
result := cmder.New("make", "all").
    WithAttemptTimeout(30 * time.Second).
    WithResetAttemptTimeoutOnOutput(true).
    Run(ctx)
```

## Context Cancellation

The parent context can also cancel:

```go
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
defer cancel()

result := cmder.New("long-task").Run(ctx)
```

If the context is cancelled, the command is killed immediately.

## Timeout Behavior

When a timeout occurs:

1. The command process is killed
2. The retry filter is consulted
3. If retry filter returns `true` and retries remain, retry
4. Otherwise, return error with `context.DeadlineExceeded`

## No Timeout

Without any timeout, commands run until completion:

```go
// Runs forever if the command hangs
result := cmder.New("potentially-hanging-command").Run(ctx)
```

!!! warning
    Always set timeouts for production code to prevent hanging processes.
