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

type CommandExecutor struct {
	masterURL  string
	kubeConfig string
}

type Output string

// todo: initialize rest object here
func NewCommandExecutor(masterURL string, kubeConfig string) *CommandExecutor {
	return &CommandExecutor{masterURL: masterURL, kubeConfig: kubeConfig}
}

func (ce *CommandExecutor) Execute(
	ctx context.Context,
	podName, containerName, namespace string,
	command []string,
) (Output, error) {
	rest, err := newK8SRestClient(ce.masterURL, ce.kubeConfig)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	request := rest.client.
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   command,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec).Timeout(60 * time.Second)

	exec, err := remotecommand.NewSPDYExecutor(rest.config, "POST", request.URL())
	if err != nil {
		return "", fmt.Errorf("failed to instantiate SPDY executor: %w", err)
	}

	if streamErr := exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: buf,
		Stderr: errBuf,
	}); streamErr != nil {
		return "", fmt.Errorf(
			"failed executing command %+v on %s/%s request URL: %s, err: %s, bufErr: %s",
			command,
			namespace,
			podName,
			request.URL().String(),
			streamErr.Error(),
			errBuf.String()+buf.String(),
		)
	}

	return Output(buf.String()), nil
}
