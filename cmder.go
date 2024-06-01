package cmder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/GiGurra/cmder/internal/util_ctx"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Spec struct {

	// Minimum input
	App  string
	Args []string

	// Extra options
	Cwd                         string
	AttemptTimeout              time.Duration
	TotalTimeout                time.Duration
	ResetAttemptTimeoutOnOutput bool
	Retries                     int
	RetryFilter                 func(err error, isAttemptTimeout bool) bool

	// Input/Output
	StdIn            io.Reader
	StdOut           io.Writer // if capturing output while running
	StdErr           io.Writer // if capturing output while running
	CollectAllOutput bool      // if running for a very long time, set this false to avoid OOM

	// debug functionality
	Verbose bool
}

type Result struct {
	StdOut   string
	StdErr   string
	Combined string
	Err      error
	Attempts int
	Code     int
}

func New(appAndArgs ...string) Spec {

	result := Spec{
		RetryFilter:      DefaultRetryFilter,
		CollectAllOutput: true,
	}

	if len(appAndArgs) > 0 {
		result.App = appAndArgs[0]
	}

	if len(appAndArgs) > 1 {
		result.Args = appAndArgs[1:]
	}

	return result
}

//goland:noinspection GoUnusedExportedFunction
func NewA(app string, args ...string) Spec {
	return New(append([]string{app}, args...)...)
}

func DefaultRetryFilter(err error, isAttemptTimeout bool) bool {
	return TimeoutRetryFilter(err, isAttemptTimeout)
}

// TimeoutRetryFilter is a simple retry policy that retries on timeouts only
func TimeoutRetryFilter(err error, isAttemptTimeout bool) bool {

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	if isAttemptTimeout {
		return true
	}

	return false
}

// WithTotalTimeout sets the total timeout for the command including retries
func (c Spec) WithTotalTimeout(timeout time.Duration) Spec {
	c.TotalTimeout = timeout
	return c
}

// WithResetAttemptTimeoutOnOutput resets the timeout if output is received from the command
func (c Spec) WithResetAttemptTimeoutOnOutput(enabled bool) Spec {
	c.ResetAttemptTimeoutOnOutput = enabled
	return c
}

// WithWD sets the working directory for the command
func (c Spec) WithWD(cwd string) Spec {
	c.Cwd = cwd
	return c
}

// WithCollectAllOutput sets whether to collect all output. Default is true. If false, you need to inject your own io.Writer
func (c Spec) WithCollectAllOutput(collect bool) Spec {
	c.CollectAllOutput = collect
	return c
}

// WithStdIn sets the standard input for the command
func (c Spec) WithStdIn(reader io.Reader) Spec {
	c.StdIn = reader
	return c
}

// WithApp sets the application to run
func (c Spec) WithApp(app string) Spec {
	c.App = app
	return c
}

// WithArgs sets the arguments for the command
func (c Spec) WithArgs(newArgs ...string) Spec {
	c.Args = newArgs
	return c
}

// WithExtraArgs appends the arguments to the command
func (c Spec) WithExtraArgs(extraArgs ...string) Spec {
	c.Args = append(c.Args, extraArgs...)
	return c
}

// WithRetryFilter sets the retry filter
func (c Spec) WithRetryFilter(filter func(err error, isAttemptTimeout bool) bool) Spec {
	c.RetryFilter = filter
	return c
}

// WithRetries sets the number of retries before giving up
func (c Spec) WithRetries(n int) Spec {
	c.Retries = n
	return c
}

// WithVerbose sets the verbose flag
func (c Spec) WithVerbose(verbose bool) Spec {
	c.Verbose = verbose
	return c
}

// WithAttemptTimeout sets the timeout for the command
// This is the total time the command can run, per attempt
func (c Spec) WithAttemptTimeout(timeout time.Duration) Spec {
	c.AttemptTimeout = timeout
	return c
}

func (c Spec) logBeforeRun() {
	if c.Verbose {
		if c.App == "sh" && len(c.Args) > 0 && c.Args[0] == "-c" {
			slog.Info(fmt.Sprintf("%s$ %s\n", c.Cwd, strings.Join(c.Args[1:], " ")))
		} else {
			slog.Info(fmt.Sprintf("%s$ %s %s\n", c.Cwd, c.App, strings.Join(c.Args, " ")))
		}
	}
}

func (c Spec) WithStdOutErrForwarded() Spec {
	c.StdOut = os.Stdout
	c.StdErr = os.Stderr
	return c
}

func (c Spec) WithStdOutForwarded() Spec {
	c.StdOut = os.Stdout
	return c
}

func (c Spec) WithStdErrForwarded() Spec {
	c.StdErr = os.Stderr
	return c
}

func (c Spec) Run(ctx context.Context) (Result, error) {

	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}
	combinedBuffer := &bytes.Buffer{}
	attempts := 0

	// This channel is used to signal that the timeout should be reset
	resetChan := make(chan any, 1)
	defer close(resetChan)

	err := c.withRetries(ctx, resetChan, func(cmd *exec.Cmd) error {

		// Reset these each time, because they could internally
		attempts++
		cmd.Stdin = c.StdIn

		// create a writer that writes to buffer, but also sends a signal to reset the timeout
		combinedWriter := util_ctx.NewResetWriterCh(combinedBuffer, resetChan)

		if c.CollectAllOutput {
			stdoutBuffer = &bytes.Buffer{}
			stderrBuffer = &bytes.Buffer{}
			combinedBuffer = &bytes.Buffer{}

			if c.StdOut != nil {
				cmd.Stdout = io.MultiWriter(c.StdOut, stdoutBuffer, combinedWriter)
			} else {
				cmd.Stdout = io.MultiWriter(stdoutBuffer, combinedWriter)
			}
			if c.StdErr != nil {
				cmd.Stderr = io.MultiWriter(c.StdErr, stderrBuffer, combinedWriter)
			} else {
				cmd.Stderr = io.MultiWriter(stderrBuffer, combinedWriter)
			}
		} else {
			if c.StdOut != nil {
				cmd.Stdout = io.MultiWriter(c.StdOut, combinedWriter)
			} else {
				cmd.Stdout = combinedWriter
			}
			if c.StdErr != nil {
				cmd.Stderr = io.MultiWriter(c.StdErr, combinedWriter)
			} else {
				cmd.Stderr = combinedWriter
			}
		}

		return cmd.Run()
	})

	stdout := stdoutBuffer.String()
	stderr := stderrBuffer.String()
	combined := combinedBuffer.String()

	return Result{
		StdOut:   stdout,
		StdErr:   stderr,
		Combined: combined,
		Err:      err,
		Attempts: attempts,
	}, err
}

func (c Spec) withRetries(srcCtx context.Context, recvSignal <-chan any, processor func(cmd *exec.Cmd) error) error {

	c.logBeforeRun()

	ctx := srcCtx // needed so we don't cancel the parent context

	if c.TotalTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.TotalTimeout)
		defer cancel() // os effectively called after processor(cmd)
	}

	for i := 0; i <= c.Retries; i++ {

		ctx := ctx // needed so we don't cancel the parent context

		// Every retry needs its own timeout context
		isAttemptTimeout := false
		err := func() error {
			if c.AttemptTimeout > 0 {
				var cancel context.CancelFunc
				var resetTimeout util_ctx.ResetFunc
				ctx, cancel, resetTimeout = util_ctx.WithTimeoutAndReset(ctx, c.AttemptTimeout)
				defer cancel() // os effectively called after processor(cmd)
				go func() {
					for {
						select {
						case <-ctx.Done():
							isAttemptTimeout = true
							return
						case _, ok := <-recvSignal:
							if !ok {
								return
							}
							if c.ResetAttemptTimeoutOnOutput {
								resetTimeout()
							}
						}
					}
				}()
			} else {
				go func() {
					for {
						select {
						case _, ok := <-recvSignal:
							if !ok {
								return
							}
						}
					}
				}()
			}

			cmd := exec.CommandContext(ctx, c.App, c.Args...)
			cmd.Dir = c.Cwd

			return processor(cmd)

		}()

		// Check if context is timed out

		if err != nil {

			if c.RetryFilter(err, isAttemptTimeout) {
				if c.Verbose {
					slog.Warn(fmt.Sprintf("retrying %s, attempt %d/%d \n", c.App, i+1, c.Retries+1))
				}
				continue
			} else {
				return fmt.Errorf("error running cmd %s \n %s: %w", c.App, err.Error(), err)
			}
		}

		return nil

	}

	return fmt.Errorf("error running cmd %s \n %s: %w", c.App, "timeout", context.DeadlineExceeded)
}
