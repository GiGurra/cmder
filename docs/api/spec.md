# Spec

The `Spec` struct defines a command specification. All methods return a new `Spec` for chaining.

## Creating a Spec

```go
// Command with arguments
spec := cmder.New("ls", "-la", "/tmp")

// Alternative: explicit app and args
spec := cmder.NewA("ls", "-la", "/tmp")
```

## Fields

| Field | Type | Description |
|-------|------|-------------|
| `App` | `string` | The application to run |
| `Args` | `[]string` | Command arguments |
| `WorkingDirectory` | `string` | Working directory |
| `AttemptTimeout` | `time.Duration` | Timeout per attempt |
| `TotalTimeout` | `time.Duration` | Total timeout including retries |
| `ResetAttemptTimeoutOnOutput` | `bool` | Reset timeout on output |
| `Retries` | `int` | Number of retry attempts |
| `RetryFilter` | `func(error, bool) bool` | Custom retry logic |
| `StdIn` | `io.Reader` | Standard input |
| `StdOut` | `io.Writer` | Additional stdout writer |
| `StdErr` | `io.Writer` | Additional stderr writer |
| `CollectAllOutput` | `bool` | Collect output in result (default: true) |
| `Verbose` | `bool` | Enable verbose logging |

## Builder Methods

### Command Configuration

```go
WithApp(app string) Spec
```
Set the application to run.

```go
WithArgs(args ...string) Spec
```
Set command arguments (replaces existing).

```go
WithExtraArgs(args ...string) Spec
```
Append additional arguments.

```go
WithCmd(app string, args ...string) Spec
```
Set both application and arguments.

```go
WithWorkingDirectory(wd string) Spec
```
Set the working directory.

### Timeout Configuration

```go
WithAttemptTimeout(timeout time.Duration) Spec
```
Set timeout for each attempt.

```go
WithTotalTimeout(timeout time.Duration) Spec
```
Set total timeout including all retries.

```go
WithResetAttemptTimeoutOnOutput(enabled bool) Spec
```
Reset attempt timeout when output is received.

### Retry Configuration

```go
WithRetries(n int) Spec
```
Set number of retry attempts.

```go
WithRetryFilter(filter func(err error, isAttemptTimeout bool) bool) Spec
```
Set custom retry filter function.

### I/O Configuration

```go
WithStdIn(reader io.Reader) Spec
```
Set standard input source.

```go
WithStdInForwarded() Spec
```
Connect stdin to `os.Stdin`.

```go
WithStdOut(writer io.Writer) Spec
```
Add stdout destination.

```go
WithStdErr(writer io.Writer) Spec
```
Add stderr destination.

```go
WithStdOutForwarded() Spec
```
Forward stdout to `os.Stdout`.

```go
WithStdErrForwarded() Spec
```
Forward stderr to `os.Stderr`.

```go
WithStdOutErrForwarded() Spec
```
Forward both stdout and stderr.

```go
WithCollectAllOutput(collect bool) Spec
```
Enable/disable output collection in result.

### Other

```go
WithVerbose(verbose bool) Spec
```
Enable verbose logging.

## Running

```go
Run(ctx context.Context) Result
```
Execute the command and return the result.

## Default Retry Filters

```go
func DefaultRetryFilter(err error, isAttemptTimeout bool) bool
```
The default filter (retries on timeouts).

```go
func TimeoutRetryFilter(err error, isAttemptTimeout bool) bool
```
Retry only on timeouts.

## Example

```go
spec := cmder.New("curl").
    WithArgs("-s", "https://api.example.com/data").
    WithAttemptTimeout(10 * time.Second).
    WithRetries(3).
    WithRetryFilter(func(err error, isTimeout bool) bool {
        return isTimeout || strings.Contains(err.Error(), "connection")
    })

result := spec.Run(context.Background())
```
