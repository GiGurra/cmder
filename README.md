# cmder

`cmder` is a Go package designed to facilitate the execution of external commands with advanced features such as
retries, timeouts, and output collection. It provides a flexible and robust way to run commands, handle their
input/output, and manage their execution lifecycle.

## Features

- **Command Execution**: Run external commands with specified arguments.
- **Timeouts**: Set timeouts for individual attempts and total execution time.
- **Retries**: Retry commands on failure with customizable retry policies.
- **Output Handling**: Capture and manage standard output and error streams.
- **Verbose Logging**: Enable detailed logging for debugging purposes.
- **Working Directory**: Specify the working directory for the command execution.
- **Input Forwarding**: Forward standard input to the command.

## Installation

To install the package, use:

```sh
go get github.com/GiGurra/cmder
```

## Usage

Here's a basic example of how to use `cmder` to run a command:

```go
package main

import (
	"context"
	"fmt"
	"github.com/GiGurra/cmder"
	"time"
)

func main() {
	result := cmder.New("ls", "-la").
		WithAttemptTimeout(5 * time.Second).
		WithRetries(3).
		WithVerbose(true).
		Run(context.Background())

	if result.Err != nil {
		fmt.Printf("Command failed: %v\n", result.Err)
	} else {
		fmt.Printf("Command succeeded: %s\n", result.StdOut)
	}
}

```

## API

### Spec

Every command is defined by a `Spec` struct, which contains the command specification and options. The `Spec` struct
provides methods to configure the command, such as setting the application to run, arguments, timeouts, retries, and
output handling.

The `Spec` struct defines the command specification and options:

- `App`: The application to run.
- `Args`: Arguments for the command.
- `WorkingDirectory`: The working directory for the command.
- `AttemptTimeout`: Timeout for each attempt.
- `TotalTimeout`: Total timeout including retries.
- `ResetAttemptTimeoutOnOutput`: Reset attempt timeout on output.
- `Retries`: Number of retries before giving up.
- `RetryFilter`: Custom retry filter function.
- `StdIn`: Standard input for the command.
- `StdOut`: Standard output for the command.
- `StdErr`: Standard error for the command.
- `CollectAllOutput`: Whether to collect all output.
- `Verbose`: Enable verbose logging.

The `Spec` struct provides the following methods to configure the command:

- `New(appAndArgs ...string) Spec`: Create a new command specification.
- `WithTotalTimeout(timeout time.Duration) Spec`: Set total timeout.
- `WithStdOut(writer io.Writer) Spec`: Set standard output writer.
- `WithStdErr(writer io.Writer) Spec`: Set standard error writer.
- `WithResetAttemptTimeoutOnOutput(enabled bool) Spec`: Enable/disable resetting attempt timeout on output.
- `WithWorkingDirectory(wd string) Spec`: Set working directory.
- `WithCollectAllOutput(collect bool) Spec`: Enable/disable collecting all output.
- `WithStdIn(reader io.Reader) Spec`: Set standard input reader.
- `WithApp(app string) Spec`: Set application to run.
- `WithArgs(newArgs ...string) Spec`: Set command arguments.
- `WithExtraArgs(extraArgs ...string) Spec`: Append extra arguments.
- `WithRetryFilter(filter func(err error, isAttemptTimeout bool) bool) Spec`: Set retry filter.
- `WithRetries(n int) Spec`: Set number of retries.
- `WithVerbose(verbose bool) Spec`: Enable/disable verbose logging.
- `WithAttemptTimeout(timeout time.Duration) Spec`: Set attempt timeout.
- `Run(ctx context.Context) Result`: Run the command and return the result.

Each method returns a new `Spec` instance, allowing for method chaining to configure the command. You can also create
general templates that you re-use across multiple commands and modules of your program.

### Result

The `Result` struct contains the result of the command execution:

- `StdOut`: Captured standard output.
- `StdErr`: Captured standard error.
- `Combined`: Combined output of standard output and error.
- `Err`: Error encountered during execution.
- `Attempts`: Number of attempts made.
- `ExitCode`: Exit code of the command.

## Testing

The package includes tests to verify its functionality. To run the tests, use:

```sh
go test ./...
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

---

For more details, refer to the source code and comments within the package.
