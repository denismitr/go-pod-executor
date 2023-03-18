package podexecutor

import (
	"fmt"
	"io"
	"time"
)

type Request struct {
	// Pod is the name of the pod where command should be executed
	// required field
	Pod string

	// Namespace is the kubernetes namespace where command should be executed
	// by default is set to `default` namespace
	Namespace string

	// Container is the container name where command should be executed
	Container string

	// Command is the command that is supposed to be executed
	Command []string

	// Timeout is the timeout for the command execution
	// defaults to 1 minute
	Timeout time.Duration

	// Stdout optional writer to duplicate the output of stdout stream to, note that result
	// of the execution will contain the buffer with the stdout output anyway
	Stdout io.Writer

	// Stderr optional writer to duplicate the output of stderr stream to, note that result
	// of the execution will contain the buffer with the stderr output anyway
	Stderr io.Writer
}

func (r *Request) applyDefaults() {
	if r.Namespace == "" {
		r.Namespace = ""
	}

	if r.Timeout == 0 {
		r.Timeout = 1 * time.Minute
	}
}

func (r *Request) validate() error {
	if r.Pod == "" {
		return fmt.Errorf("%w: pod name must be specified", ErrInvalidRequest)
	}

	if len(r.Command) == 0 {
		return fmt.Errorf("%w: command slice should not be empty", ErrInvalidRequest)
	}

	return nil
}
