package cmder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/GiGurra/cmder/internal/util"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"
)

type Spec struct {

	// Minimum input
	App  string
	Args []string

	// Extra options
	WorkingDirectory            string
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
	ExitCode int
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

// WithStdOut sets the standard output for the command
func (c Spec) WithStdOut(writer io.Writer) Spec {
	c.StdOut = writer
	return c
}

// WithStdErr sets the standard error for the command
func (c Spec) WithStdErr(writer io.Writer) Spec {
	c.StdErr = writer
	return c
}

// WithResetAttemptTimeoutOnOutput resets the timeout if output is received from the command
func (c Spec) WithResetAttemptTimeoutOnOutput(enabled bool) Spec {
	c.ResetAttemptTimeoutOnOutput = enabled
	return c
}

// WithWorkingDirectory sets the working directory for the command
func (c Spec) WithWorkingDirectory(wd string) Spec {
	c.WorkingDirectory = wd
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

// WithStdInForwarded sets the standard input to os.Stdin
func (c Spec) WithStdInForwarded() Spec {
	c.StdIn = os.Stdin
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
		slog.Info(fmt.Sprintf("%s$ %s %s\n", c.WorkingDirectory, c.App, strings.Join(c.Args, " ")))
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

func (c Spec) Run(ctx context.Context) Result {

	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}
	combinedBuffer := &bytes.Buffer{}
	attempts := 0
	exitCode := 0

	err := c.withRetries(ctx, func(cmd *exec.Cmd, aliveChannel chan any) error {

		exitCode = 0

		// Reset these each time, because they could internally
		attempts++
		cmd.Stdin = c.StdIn

		// create a writer that writes to buffer, but also sends a signal to reset the timeout
		sfw := util.NewSignalForwarderWriter(aliveChannel)

		// The sfw is used to reset the attempt timeout, when used
		stdOutTargets := []io.Writer{sfw}
		stdErrTargets := []io.Writer{sfw}

		// If we are collecting all output, we need to write to the corresponding buffers
		if c.CollectAllOutput {
			stdoutBuffer = &bytes.Buffer{}
			stderrBuffer = &bytes.Buffer{}
			combinedBuffer = &bytes.Buffer{}
			stdOutTargets = append(stdOutTargets, stdoutBuffer, combinedBuffer)
			stdErrTargets = append(stdErrTargets, stderrBuffer, combinedBuffer)
		}

		// If we are capturing output, we need to write to the corresponding writers
		if c.StdOut != nil {
			stdOutTargets = append(stdOutTargets, c.StdOut)
		}
		if c.StdErr != nil {
			stdErrTargets = append(stdErrTargets, c.StdErr)
		}

		// Set the writers
		cmd.Stdout = io.MultiWriter(stdOutTargets...)
		cmd.Stderr = io.MultiWriter(stdErrTargets...)

		err := cmd.Run() // waits internally

		if err != nil {
			if cmd.ProcessState != nil {
				exitCode = cmd.ProcessState.ExitCode()
			} else {
				exitCode = -1
			}
		}

		return err

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
		ExitCode: exitCode,
	}
}

func executeAfterDuration(ctx context.Context, duration time.Duration, task func()) {
	go func() {
		select {
		case <-time.After(duration):
			task()
		case <-ctx.Done():
		}
	}()
}

func toPtr[T any](x T) *T {
	return &x
}

func (c Spec) withRetries(parentCtx context.Context, processor func(cmd *exec.Cmd, aliveSignal chan any) error) error {

	c.logBeforeRun()

	jobCtx, cancelJobCtx := context.WithCancel(parentCtx)
	defer cancelJobCtx()

	jobTimedOut := atomic.Bool{}

	if c.TotalTimeout > 0 {
		executeAfterDuration(jobCtx, c.TotalTimeout, func() {
			jobTimedOut.Store(true)
			cancelJobCtx()
		})
	}

	for i := 0; i <= c.Retries; i++ {

		attemptTimedOut := atomic.Bool{}
		attemptDeadline := atomic.Pointer[time.Time]{}
		attemptDeadline.Store(toPtr(time.Now().Add(c.AttemptTimeout)))

		err := func() error {

			aliveSignal := make(chan any, 10)
			defer close(aliveSignal)

			attemptCtx, cancelAttemptCtx := context.WithCancel(jobCtx)
			defer cancelAttemptCtx()

			if c.AttemptTimeout > 0 {
				var checkTimeoutFunc func()
				checkTimeoutFunc = func() {
					curDeadline := attemptDeadline.Load()
					if time.Now().After(*curDeadline) {
						attemptTimedOut.Store(true)
						cancelAttemptCtx()
					} else {
						executeAfterDuration(attemptCtx, curDeadline.Sub(time.Now())+1*time.Millisecond, func() {
							checkTimeoutFunc()
						})
					}
				}
				checkTimeoutFunc()
			}

			go func() {
				for {
					select {
					case <-aliveSignal:
						if c.ResetAttemptTimeoutOnOutput {
							attemptDeadline.Store(toPtr(time.Now().Add(c.AttemptTimeout)))
						}
					case <-attemptCtx.Done():
						return
					}
				}
			}()

			cmd := exec.CommandContext(attemptCtx, c.App, c.Args...)
			cmd.Dir = c.WorkingDirectory

			return processor(cmd, aliveSignal)

		}()

		if err != nil {
			if c.RetryFilter(err, attemptTimedOut.Load()) {
				if c.Verbose {
					slog.Warn(fmt.Sprintf("retrying %s, attempt %d/%d \n", c.App, i+1, c.Retries+1))
				}
				continue
			} else {
				if jobTimedOut.Load() {
					return fmt.Errorf("error running cmd %s \n %s: %w", c.App, "timeout", context.DeadlineExceeded)
				} else {
					return fmt.Errorf("error running cmd %s \n %s: %w", c.App, err.Error(), err)
				}
			}
		}

		return nil

	}

	return fmt.Errorf("error running cmd %s \n %s: %w", c.App, "timeout and max retries exceeded", context.DeadlineExceeded)
}
