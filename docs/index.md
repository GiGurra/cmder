# cmder

Go package for executing external commands with retries, timeouts, and output collection.

## Why cmder?

Go's `os/exec` is powerful but basic. Real-world command execution often needs:

- **Timeouts** - Commands that hang shouldn't hang forever
- **Retries** - Flaky operations benefit from retry logic
- **Output capture** - You often need stdout, stderr, or both
- **Keep-alive** - Long-running commands that produce output shouldn't timeout

cmder wraps `os/exec` with these capabilities while keeping a clean, fluent API.

## Quick Example

```go
result := cmder.New("curl", "-s", "https://api.example.com/health").
    WithAttemptTimeout(5 * time.Second).
    WithRetries(3).
    Run(context.Background())

if result.Err != nil {
    log.Printf("Health check failed after %d attempts: %v", result.Attempts, result.Err)
} else {
    log.Printf("Service healthy: %s", result.StdOut)
}
```

## Features

| Feature | Description |
|---------|-------------|
| Attempt timeout | Kill a single attempt after duration |
| Total timeout | Limit total time including all retries |
| Retries | Automatic retry with configurable filter |
| Output collection | Capture stdout, stderr, combined |
| Timeout reset | Reset timeout when command produces output |
| Fluent API | Builder pattern for clean configuration |

## Installation

```bash
go get github.com/GiGurra/cmder
```

## Next Steps

- [Getting Started](guide/getting-started.md) - Basic usage patterns
- [Timeouts](guide/timeouts.md) - Timeout configuration
- [Retries](guide/retries.md) - Retry strategies
- [API Reference](api/spec.md) - Full API documentation
