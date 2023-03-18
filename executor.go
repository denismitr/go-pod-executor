package podexecutor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

var (
	ErrInvalidRequest = fmt.Errorf("invalid command execute request")
)

type CommandExecutor struct {
	masterURL string
	rest      *k8sRest
}

type Result struct {
	buf    *bytes.Buffer
	errBuf *bytes.Buffer
}

func newResult() *Result {
	return &Result{
		buf:    &bytes.Buffer{},
		errBuf: &bytes.Buffer{},
	}
}

func (r Result) Output() string {
	return r.buf.String()
}

func (r Result) ErrOutput() string {
	return r.errBuf.String()
}

func NewCommandExecutor(masterURL string, kubeConfig string) (*CommandExecutor, error) {
	rst, err := newK8SRestClient(masterURL, kubeConfig)
	if err != nil {
		return nil, err
	}

	return &CommandExecutor{masterURL: masterURL, rest: rst}, nil
}

func (ce *CommandExecutor) Execute(
	ctx context.Context,
	req *Request,
) (*Result, error) {
	req.applyDefaults()
	if err := req.validate(); err != nil {
		return nil, err
	}

	exec, err := ce.prepareRequestExecutor(req)
	if err != nil {
		return nil, err
	}

	result := newResult()
	streamOpts := newStreamOptions(req, result)
	if streamErr := exec.StreamWithContext(ctx, streamOpts); streamErr != nil {
		fullErr := fmt.Errorf(
			"failed executing command %+v on %s/%s in container %s: %s",
			req.Command,
			req.Namespace,
			req.Pod,
			req.Container,
			streamErr.Error(),
		)

		stdErr := result.errBuf.String()
		if stdErr != "" {
			fullErr = fmt.Errorf("%w, stderr: %s", fullErr, stdErr)
		}

		return nil, fullErr
	}

	return result, nil
}

func (ce *CommandExecutor) prepareRequestExecutor(req *Request) (remotecommand.Executor, error) {
	request := ce.rest.client.
		Post().
		Namespace(req.Namespace).
		Resource("pods").
		Name(req.Pod).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: req.Container,
			Command:   req.Command,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec).Timeout(req.Timeout)

	exec, err := remotecommand.NewSPDYExecutor(ce.rest.config, "POST", request.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request executor: %w", err)
	}
	return exec, nil
}

func newStreamOptions(req *Request, r *Result) remotecommand.StreamOptions {
	opts := remotecommand.StreamOptions{}
	if req.Stdout != nil {
		opts.Stdout = io.MultiWriter(r.buf, req.Stdout)
	} else {
		opts.Stdout = r.buf
	}
	if req.Stderr != nil {
		opts.Stderr = io.MultiWriter(r.errBuf, req.Stderr)
	} else {
		opts.Stderr = r.errBuf
	}
	return opts
}
