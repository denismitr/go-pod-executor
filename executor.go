package podexecutor

import (
	"bytes"
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"time"
)

var (
	ErrInvalidRequest = fmt.Errorf("invalid command execute request")
)

type CommandExecutor struct {
	masterURL string
	rest      *k8sRest
}

type Output string

func NewCommandExecutor(masterURL string, kubeConfig string) (*CommandExecutor, error) {
	rst, err := newK8SRestClient(masterURL, kubeConfig)
	if err != nil {
		return nil, err
	}

	return &CommandExecutor{masterURL: masterURL, rest: rst}, nil
}

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
}

func (ce *CommandExecutor) Execute(ctx context.Context, r *Request) (Output, error) {
	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	request := ce.rest.client.
		Post().
		Namespace(r.Namespace).
		Resource("pods").
		Name(r.Pod).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: r.Container,
			Command:   r.Command,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec).Timeout(r.Timeout)

	exec, err := remotecommand.NewSPDYExecutor(ce.rest.config, "POST", request.URL())
	if err != nil {
		return "", fmt.Errorf("failed to instantiate SPDY executor: %w", err)
	}

	if streamErr := exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: buf,
		Stderr: errBuf,
	}); streamErr != nil {
		fullErr := fmt.Errorf(
			"failed executing command %+v on %s/%s in container %s: %s",
			r.Command,
			r.Namespace,
			r.Pod,
			r.Container,
			streamErr.Error(),
		)

		stdErr := errBuf.String()
		if stdErr != "" {
			fullErr = fmt.Errorf("%w, stderr: %s", fullErr, stdErr)
		}

		return "", fullErr
	}

	return Output(buf.String()), nil
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
