# Retries

cmder supports automatic retries with configurable strategies.

## Basic Retries

```go
result := cmder.New("flaky-command").
    WithRetries(3).
    WithAttemptTimeout(5 * time.Second).
    Run(ctx)
```

This allows up to 4 total attempts (1 initial + 3 retries).

## Default Behavior

By default, cmder only retries on timeouts:

```go
// Default retry filter
func DefaultRetryFilter(err error, isAttemptTimeout bool) bool {
    return TimeoutRetryFilter(err, isAttemptTimeout)
}

func TimeoutRetryFilter(err error, isAttemptTimeout bool) bool {
    if errors.Is(err, context.DeadlineExceeded) {
        return true
    }
    if isAttemptTimeout {
        return true
    }
    return false
}
```

This means:
- Command exits with non-zero: **not retried**
- Command times out: **retried**
- Context cancelled: **retried**

## Custom Retry Filter

Control exactly when to retry:

```go
result := cmder.New("api-call").
    WithRetries(5).
    WithRetryFilter(func(err error, isTimeout bool) bool {
        // Retry on timeouts
        if isTimeout {
            return true
        }
        // Retry on specific error messages
        if strings.Contains(err.Error(), "connection refused") {
            return true
        }
        // Don't retry other errors
        return false
    }).
    Run(ctx)
```

## Always Retry

Retry on any failure:

```go
result := cmder.New("unreliable-script").
    WithRetries(3).
    WithRetryFilter(func(err error, isTimeout bool) bool {
        return true  // Always retry
    }).
    Run(ctx)
```

## Never Retry

Disable retries entirely:

```go
result := cmder.New("one-shot").
    WithRetries(0).  // or just don't set it
    Run(ctx)
```

## Retry with Total Timeout

Limit total retry time:

```go
result := cmder.New("flaky-service").
    WithRetries(10).
    WithAttemptTimeout(5 * time.Second).
    WithTotalTimeout(30 * time.Second).
    Run(ctx)
```

Even with 10 retries configured, stops after 30 seconds total.

## Checking Attempts

The result tells you how many attempts were made:

```go
result := cmder.New("flaky").
    WithRetries(5).
    Run(ctx)

if result.Err != nil {
    log.Printf("Failed after %d attempts: %v", result.Attempts, result.Err)
}
```

## Retry Patterns

### Retry Transient Failures

```go
WithRetryFilter(func(err error, isTimeout bool) bool {
    msg := err.Error()
    transient := []string{
        "connection refused",
        "temporary failure",
        "try again",
    }
    for _, t := range transient {
        if strings.Contains(strings.ToLower(msg), t) {
            return true
        }
    }
    return isTimeout
})
```

### Retry Based on Exit Code

```go
// Note: exit code is in result, not directly available in filter
// For exit-code-based retry, check result.ExitCode after Run()
```

## Verbose Retry Logging

See retry activity:

```go
result := cmder.New("flaky").
    WithRetries(3).
    WithVerbose(true).
    Run(ctx)
```

Logs each retry attempt.
