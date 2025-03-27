package k8s

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"syscall"

	"github.com/deviceinsight/kafkactl/v5/internal/output"
)

// Runner interface for shell commands
type Runner interface {
	ExecuteAndReturn(cmd string, args []string) ([]byte, error)
	Execute(cmd string, args []string) error
}

type ShellRunner struct {
	Dir string
}

func (shell *ShellRunner) ExecuteAndReturn(binary string, args []string) ([]byte, error) {
	cmd := exec.Command(binary, args...)
	cmd.Dir = shell.Dir
	cmd.Env = os.Environ()

	if cmd.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	if cmd.Stderr != nil {
		return nil, errors.New("exec: Stderr already set")
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var combined bytes.Buffer
	cmd.Stdout = io.MultiWriter(&stdout, &combined)
	cmd.Stderr = io.MultiWriter(&stderr, &combined)
	err := cmd.Run()

	if err != nil {
		var exitError *exec.ExitError
		switch {
		case errors.As(err, &exitError):
			// Propagate any non-zero exit status from the external command
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			exitStatus := waitStatus.ExitStatus()
			err = newExitError(cmd.Path, cmd.Args, exitStatus, exitError, stderr.String(), combined.String())
		default:
			return nil, fmt.Errorf("unexpected error: %w", err)
		}
	}

	return stdout.Bytes(), err
}

// Execute a shell command
func (shell *ShellRunner) Execute(binary string, args []string) error {
	cmd := exec.Command(binary, args...)
	cmd.Dir = shell.Dir
	cmd.Env = os.Environ()

	// get stdOut of cmd
	stdoutIn, _ := cmd.StdoutPipe()

	// stdin, stderr directly mapped from outside
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var writeErr error
	wg.Add(1)
	go func() {
		writeErr = filterOutput(output.IoStreams.Out, stdoutIn)
		wg.Done()
	}()

	wg.Wait()

	if writeErr != nil {
		return fmt.Errorf("unable to write std out: %w", writeErr)
	}

	err = cmd.Wait()

	if err != nil {
		var exitError *exec.ExitError
		switch {
		case errors.As(err, &exitError):
			// Propagate any non-zero exit status from the external command
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			exitStatus := waitStatus.ExitStatus()
			err = newExitError(cmd.Path, cmd.Args, exitStatus, exitError, "", "")
		default:
			return fmt.Errorf("unexpected error: %w", err)
		}
	}

	return err
}

func filterOutput(w io.Writer, r io.Reader) error {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			data := buf[:n]
			data = filterData(data)
			_, err := w.Write(data)
			if err != nil {
				return err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return err
		}
	}
}

func filterData(data []byte) []byte {
	// filter unwanted stuff from kubectl output
	var podDeletedRegex = regexp.MustCompile(`pod "kafkactl-.+" deleted\n`)
	return podDeletedRegex.ReplaceAll(data, []byte(""))
}
