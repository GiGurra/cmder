# Output Handling

cmder provides flexible control over command output.

## Default: Collect All Output

By default, all output is captured:

```go
result := cmder.New("ls", "-la").Run(ctx)

fmt.Println(result.StdOut)   // Standard output
fmt.Println(result.StdErr)   // Standard error
fmt.Println(result.Combined) // Both interleaved
```

## Stream to Terminal

Forward output to the terminal while running:

```go
result := cmder.New("build.sh").
    WithStdOutErrForwarded().
    Run(ctx)
```

Or individually:

```go
// Just stdout
cmder.New("cmd").WithStdOutForwarded()

// Just stderr
cmder.New("cmd").WithStdErrForwarded()

// Both
cmder.New("cmd").WithStdOutErrForwarded()
```

!!! note
    Forwarded output is still collected in `result.StdOut`, etc.

## Custom Writers

Send output to custom `io.Writer`:

```go
var logBuffer bytes.Buffer

result := cmder.New("command").
    WithStdOut(&logBuffer).
    WithStdErr(os.Stderr).
    Run(ctx)
```

## Disable Collection

For long-running commands that produce lots of output:

```go
result := cmder.New("generate-data").
    WithCollectAllOutput(false).
    WithStdOutErrForwarded().  // Still see it
    Run(ctx)

// result.StdOut is empty
// result.StdErr is empty
// result.Combined is empty
```

This prevents memory issues when output is large.

## Standard Input

Provide input to the command:

```go
// From string
result := cmder.New("grep", "pattern").
    WithStdIn(strings.NewReader("line1\npattern here\nline3")).
    Run(ctx)

// From file
file, _ := os.Open("input.txt")
defer file.Close()

result := cmder.New("process").
    WithStdIn(file).
    Run(ctx)
```

## Forward stdin

Connect to terminal stdin for interactive commands:

```go
result := cmder.New("interactive-tool").
    WithStdInForwarded().
    WithStdOutErrForwarded().
    Run(ctx)
```

## Multiple Output Destinations

Output goes to all configured destinations:

```go
var capture bytes.Buffer

result := cmder.New("build").
    WithStdOut(&capture).       // Capture
    WithStdOutForwarded().      // AND forward to terminal
    Run(ctx)

// capture.String() has stdout
// result.StdOut also has stdout
// Terminal also showed stdout
```

Wait, that's not quite right. Let me clarify:

```go
result := cmder.New("build").
    WithStdOut(os.Stdout).  // This replaces, doesn't add
    Run(ctx)
```

The `WithStdOut` sets an additional writer. Output goes to:
1. Internal buffer (for `result.StdOut`)
2. Your custom writer (if set)

## Exit Codes

The exit code is always available:

```go
result := cmder.New("exit-test").Run(ctx)

if result.ExitCode != 0 {
    log.Printf("Command failed with exit code: %d", result.ExitCode)
}
```

Exit codes:
- `0`: Success
- Non-zero: Failure
- `-1`: Process didn't start or state unavailable
