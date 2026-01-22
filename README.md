# cmder

[![CI Status](https://github.com/GiGurra/cmder/actions/workflows/ci.yml/badge.svg)](https://github.com/GiGurra/cmder/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GiGurra/cmder)](https://goreportcard.com/report/github.com/GiGurra/cmder)
[![Docs](https://img.shields.io/badge/docs-GitHub%20Pages-blue)](https://gigurra.github.io/cmder/)

Go package for executing external commands with retries, timeouts, and output collection.

## Features

- **Attempt timeouts** - Kill commands that run too long
- **Total timeouts** - Limit total time including retries
- **Automatic retries** - Retry on timeout with configurable filters
- **Output collection** - Capture stdout, stderr, and combined output
- **Reset timeout on output** - Keep alive commands that produce output
- **Fluent API** - Builder pattern for clean configuration

## Installation

```bash
go get github.com/GiGurra/cmder
```

## Quick Start

```go
result := cmder.New("ls", "-la").
    WithAttemptTimeout(5 * time.Second).
    WithRetries(3).
    Run(context.Background())

if result.Err != nil {
    fmt.Printf("Failed: %v\n", result.Err)
} else {
    fmt.Printf("Output: %s\n", result.StdOut)
}
```

## Common Patterns

### Command with timeout

```go
result := cmder.New("slow-command").
    WithAttemptTimeout(30 * time.Second).
    Run(ctx)
```

### Retry on failure

```go
result := cmder.New("flaky-service").
    WithRetries(5).
    WithAttemptTimeout(10 * time.Second).
    Run(ctx)
```

### Keep alive on output

```go
// Timeout resets each time command produces output
result := cmder.New("long-running-task").
    WithAttemptTimeout(5 * time.Second).
    WithResetAttemptTimeoutOnOutput(true).
    Run(ctx)
```

### Stream output while running

```go
result := cmder.New("build-script").
    WithStdOutErrForwarded().  // Print to terminal
    Run(ctx)
```

### Custom retry logic

```go
result := cmder.New("api-call").
    WithRetries(3).
    WithRetryFilter(func(err error, isTimeout bool) bool {
        return isTimeout  // Only retry timeouts
    }).
    Run(ctx)
```

## API

See the [API documentation](https://gigurra.github.io/cmder/api/spec/) for full details.

## License

MIT
