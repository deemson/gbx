package gitexec

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

type LineFunc func(string) error

func Command(ctx context.Context, path string, args ...string) *exec.Cmd {
	fullArgs := append([]string{"-C", path}, args...)
	return exec.CommandContext(ctx, "git", fullArgs...)
}

func RunCmd(cmd *exec.Cmd) (Result, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	res := Result{
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		ExitCode: cmd.ProcessState.ExitCode(),
	}
	if runErr != nil {
		return res, fmt.Errorf(
			"%s: %w: stdout=`%s` stderr=`%s`",
			strings.Join(cmd.Args, " "), runErr, res.Stdout, res.Stderr,
		)
	}
	return res, nil
}

func Run(ctx context.Context, path string, args ...string) (Result, error) {
	return RunCmd(Command(ctx, path, args...))
}

func StreamCmd(cmd *exec.Cmd, onStdout, onStderr LineFunc) (int, error) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return -1, fmt.Errorf("%s: stdout pipe: %w", strings.Join(cmd.Args, " "), err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return -1, fmt.Errorf("%s: stderr pipe: %w", strings.Join(cmd.Args, " "), err)
	}
	if err := cmd.Start(); err != nil {
		return -1, fmt.Errorf("%s: start: %w", strings.Join(cmd.Args, " "), err)
	}

	var (
		mu     sync.Mutex
		cbErr  error
		killed bool
	)
	fail := func(e error) {
		mu.Lock()
		defer mu.Unlock()
		if cbErr == nil {
			cbErr = e
		}
		if !killed {
			killed = true
			_ = cmd.Process.Kill()
		}
	}
	failed := func() bool {
		mu.Lock()
		defer mu.Unlock()
		return cbErr != nil
	}

	var wg sync.WaitGroup
	scan := func(r io.Reader, fn LineFunc) {
		defer wg.Done()
		if fn == nil {
			_, _ = io.Copy(io.Discard, r)
			return
		}
		s := bufio.NewScanner(r)
		s.Buffer(make([]byte, 64*1024), 1024*1024)
		for s.Scan() {
			if failed() {
				_, _ = io.Copy(io.Discard, r)
				return
			}
			if err := fn(s.Text()); err != nil {
				fail(err)
				_, _ = io.Copy(io.Discard, r)
				return
			}
		}
	}
	wg.Add(2)
	go scan(stdoutPipe, onStdout)
	go scan(stderrPipe, onStderr)
	wg.Wait()
	runErr := cmd.Wait()
	exitCode := cmd.ProcessState.ExitCode()

	if cbErr != nil {
		return exitCode, fmt.Errorf("%s: %w", strings.Join(cmd.Args, " "), cbErr)
	}
	if runErr != nil {
		return exitCode, fmt.Errorf("%s: %w", strings.Join(cmd.Args, " "), runErr)
	}
	return exitCode, nil
}

func Stream(ctx context.Context, path string, onStdout, onStderr LineFunc, args ...string) (int, error) {
	return StreamCmd(Command(ctx, path, args...), onStdout, onStderr)
}
