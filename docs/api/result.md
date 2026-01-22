# Result

The `Result` struct contains the outcome of running a command.

## Structure

```go
type Result struct {
    StdOut   string  // Captured standard output
    StdErr   string  // Captured standard error
    Combined string  // Interleaved stdout and stderr
    Err      error   // Error if command failed
    Attempts int     // Number of attempts made
    ExitCode int     // Exit code of the command
}
```

## Fields

### StdOut

Captured standard output.

```go
result := cmder.New("echo", "hello").Run(ctx)
fmt.Println(result.StdOut)  // "hello\n"
```

Empty if `WithCollectAllOutput(false)` was set.

### StdErr

Captured standard error.

```go
result := cmder.New("ls", "/nonexistent").Run(ctx)
fmt.Println(result.StdErr)  // "ls: /nonexistent: No such file..."
```

Empty if `WithCollectAllOutput(false)` was set.

### Combined

Interleaved stdout and stderr in order received.

```go
result := cmder.New("sh", "-c", "echo out; echo err >&2; echo out2").Run(ctx)
fmt.Println(result.Combined)
// out
// err
// out2
```

Useful when order matters.

### Err

Error if the command failed.

```go
result := cmder.New("false").Run(ctx)  // 'false' always exits 1

if result.Err != nil {
    fmt.Printf("Command failed: %v\n", result.Err)
}
```

Common error types:
- `context.DeadlineExceeded` - Timeout
- Wrapped exec errors - Command not found, etc.

### Attempts

Number of execution attempts.

```go
result := cmder.New("flaky").
    WithRetries(5).
    WithRetryFilter(func(err error, isTimeout bool) bool { return true }).
    Run(ctx)

fmt.Printf("Took %d attempts\n", result.Attempts)
// 1 = success on first try
// 2-6 = needed retries
```

### ExitCode

The command's exit code.

```go
result := cmder.New("sh", "-c", "exit 42").Run(ctx)
fmt.Println(result.ExitCode)  // 42
```

Values:
- `0` - Success
- `1-255` - Command-defined failure code
- `-1` - Process didn't start or state unavailable

## Usage Patterns

### Check Success

```go
result := cmder.New("make", "test").Run(ctx)

if result.Err != nil {
    log.Fatalf("Tests failed: %v\n%s", result.Err, result.Combined)
}
```

### Handle Specific Exit Codes

```go
result := cmder.New("grep", "pattern", "file.txt").Run(ctx)

switch result.ExitCode {
case 0:
    fmt.Printf("Found: %s", result.StdOut)
case 1:
    fmt.Println("Pattern not found")
case 2:
    fmt.Printf("Error: %s", result.StdErr)
}
```

### Log Failures with Context

```go
result := cmder.New("deploy.sh").
    WithRetries(3).
    Run(ctx)

if result.Err != nil {
    log.Printf("Deploy failed after %d attempts\n", result.Attempts)
    log.Printf("Exit code: %d\n", result.ExitCode)
    log.Printf("Stdout:\n%s", result.StdOut)
    log.Printf("Stderr:\n%s", result.StdErr)
}
```

### Parse Output

```go
result := cmder.New("git", "rev-parse", "HEAD").Run(ctx)

if result.Err == nil {
    commitHash := strings.TrimSpace(result.StdOut)
    fmt.Printf("Current commit: %s\n", commitHash)
}
```
