# Getting Started

## Installation

```bash
go get github.com/GiGurra/cmder
```

## Basic Usage

The simplest way to run a command:

```go
import "github.com/GiGurra/cmder"

result := cmder.New("ls", "-la").Run(context.Background())

if result.Err != nil {
    fmt.Printf("Error: %v\n", result.Err)
} else {
    fmt.Printf("Output:\n%s", result.StdOut)
}
```

## Creating Commands

`cmder.New()` accepts the command and arguments:

```go
// Command with arguments
cmder.New("git", "status", "--short")

// Just the command
cmder.New("pwd")
```

Alternatively, use `cmder.NewA()` for explicit separation:

```go
cmder.NewA("git", "status", "--short")
```

## Builder Pattern

All configuration methods return a new `Spec`, enabling chaining:

```go
result := cmder.New("make", "build").
    WithWorkingDirectory("/path/to/project").
    WithAttemptTimeout(5 * time.Minute).
    WithVerbose(true).
    Run(ctx)
```

## Result Structure

The `Result` struct contains:

```go
type Result struct {
    StdOut   string  // Captured stdout
    StdErr   string  // Captured stderr
    Combined string  // Interleaved stdout + stderr
    Err      error   // Error if command failed
    Attempts int     // Number of attempts made
    ExitCode int     // Exit code (0 for success)
}
```

## Working Directory

Set where the command runs:

```go
result := cmder.New("npm", "install").
    WithWorkingDirectory("/path/to/project").
    Run(ctx)
```

## Standard Input

Provide input to the command:

```go
result := cmder.New("cat").
    WithStdIn(strings.NewReader("hello world")).
    Run(ctx)

// result.StdOut == "hello world"
```

## Verbose Mode

Enable logging of command execution:

```go
result := cmder.New("build.sh").
    WithVerbose(true).
    Run(ctx)
```

This logs the working directory, command, and retry information.

## Reusable Templates

Since `Spec` is immutable, you can create templates:

```go
// Create a template
gitCmd := cmder.New("git").
    WithAttemptTimeout(30 * time.Second).
    WithRetries(2)

// Use it for different operations
gitCmd.WithArgs("status").Run(ctx)
gitCmd.WithArgs("pull").Run(ctx)
gitCmd.WithArgs("push").Run(ctx)
```

## Next Steps

- [Timeouts](timeouts.md) - Configure timeout behavior
- [Retries](retries.md) - Set up retry strategies
- [Output Handling](output.md) - Control output capture and streaming
