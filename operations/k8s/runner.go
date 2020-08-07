package k8s

import (
	"bytes"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sync"
	"syscall"
)

// Runner interface for shell commands
type Runner interface {
	ExecuteAndReturn(cmd string, args []string) ([]byte, error)
	Execute(cmd string, args []string) error
}

type ShellRunner struct {
	Dir string
}

func (shell ShellRunner) ExecuteAndReturn(binary string, args []string) ([]byte, error) {
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
		switch ee := err.(type) {
		case *exec.ExitError:
			// Propagate any non-zero exit status from the external command
			waitStatus := ee.Sys().(syscall.WaitStatus)
			exitStatus := waitStatus.ExitStatus()
			err = newExitError(cmd.Path, cmd.Args, exitStatus, ee, stderr.String(), combined.String())
		default:
			output.Fail(errors.Wrap(err, "unexpected error"))
		}
	}

	return stdout.Bytes(), err
}

// Execute a shell command
func (shell ShellRunner) Execute(binary string, args []string) error {
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
		output.Fail(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := filterOutput(os.Stdout, stdoutIn); err != nil {
			output.Fail(errors.Wrap(err, "unable to write std out"))
		}
		wg.Done()
	}()

	wg.Wait()

	err = cmd.Wait()

	if err != nil {
		switch ee := err.(type) {
		case *exec.ExitError:
			// Propagate any non-zero exit status from the external command
			waitStatus := ee.Sys().(syscall.WaitStatus)
			exitStatus := waitStatus.ExitStatus()
			err = newExitError(cmd.Path, cmd.Args, exitStatus, ee, "", "")
		default:
			output.Fail(errors.Wrap(err, "unexpected error"))
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
