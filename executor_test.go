package podexecutor_test

import (
	"bytes"
	"context"
	"github.com/denismitr/podexecutor"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

var nginxFolders = []string{
	".",
	"..",
	"bin",
	"boot",
	"dev",
	"docker-entrypoint.d",
	"docker-entrypoint.sh",
	"etc",
	"home",
	"lib",
	"media",
	"mnt",
	"opt",
	"proc",
	"product_uuid",
	"srv",
	"var",
	"root",
	"run",
	"sbin",
	"tmp",
	"usr",
	"sys",
}

func TestCommandExecutor_Execute(t *testing.T) {
	masterURL := os.Getenv("K8S_MASTER_URL")
	kubeConfig := os.Getenv("K8S_CONFIG")

	if masterURL == "" {
		t.Fatalf("master url cannot be empty")
	}

	if kubeConfig == "" {
		t.Fatalf("kube config cannot be empty")
	}

	t.Run("integration testing with executing ls in nginx", func(t *testing.T) {
		// todo: return also an error
		executor, err := podexecutor.NewCommandExecutor(masterURL, kubeConfig)
		if err != nil {
			t.Fatalf("executor constructor should not have returned an error: %s", err.Error())
		}

		result, err := executor.Execute(context.TODO(), &podexecutor.Request{
			Pod:       "nginx",
			Namespace: "executor",
			Container: "nginx",
			Command:   []string{"ls", "-a"},
		})
		if err != nil {
			t.Fatalf("executor returned an error: %s", err.Error())
		}

		output := result.Output()
		if output == "" {
			t.Errorf("error: executor output should not be empty")
		}

		folders := outputToSlice(output)
		want := nginxFolders
		sort.Strings(want)

		if !reflect.DeepEqual(folders, want) {
			t.Errorf("error: executor output expected to be %+v, got %+v", want, folders)
		}
	})

	t.Run("integration test with executing ls in nginx and using custom writer", func(t *testing.T) {
		// todo: return also an error
		executor, err := podexecutor.NewCommandExecutor(masterURL, kubeConfig)
		if err != nil {
			t.Fatalf("executor constructor should not have returned an error: %s", err.Error())
		}

		w := &bytes.Buffer{}
		result, err := executor.Execute(context.TODO(), &podexecutor.Request{
			Pod:       "nginx",
			Namespace: "executor",
			Container: "nginx",
			Command:   []string{"ls", "-a"},
			Stdout:    w,
		})
		if err != nil {
			t.Fatalf("executor returned an error: %s", err.Error())
		}

		output := result.Output()
		if output == "" {
			t.Errorf("error: executor output should not be empty")
		}

		folders := outputToSlice(output)
		want := nginxFolders
		sort.Strings(want)

		if !reflect.DeepEqual(folders, want) {
			t.Errorf("error: executor output expected to be %+v, got %+v", want, folders)
		}

		writerOutput, _ := io.ReadAll(w)
		foldersFromCustomWriter := outputToSlice(string(writerOutput))
		if !reflect.DeepEqual(foldersFromCustomWriter, want) {
			t.Errorf("error: executor output expected to be %+v, got %+v", want, foldersFromCustomWriter)
		}
	})

	t.Run("handle non existent shell interpretation command", func(t *testing.T) {
		// todo: return also an error
		executor, err := podexecutor.NewCommandExecutor(masterURL, kubeConfig)
		if err != nil {
			t.Fatalf("executor constructor should not have returned an error: %s", err.Error())
		}

		result, err := executor.Execute(context.TODO(), &podexecutor.Request{
			Pod:       "nginx",
			Namespace: "executor",
			Container: "nginx",
			Command:   []string{"sh", "-c", `"non existent command"`},
		})
		if err == nil {
			t.Fatal("execute method should have returned an error")
		}

		errMsg := strings.TrimSpace(err.Error())
		if !strings.Contains(errMsg, `failed executing command [sh -c "non existent command"] on executor/nginx in container nginx: command terminated with exit code 127, stderr: sh: 1: non existent command: not found`) {
			t.Errorf("error message is invalid: %s", errMsg)
		}

		if result != nil {
			t.Errorf("error: executor output should be empty")
		}
	})
}

func outputToSlice(out string) []string {
	slice := strings.Split(out, "\n")
	result := make([]string, 0, len(slice))
	for i := range slice {
		str := strings.TrimSpace(slice[i])
		str = strings.ReplaceAll(str, " ", "")
		if str == "" || str == " " {
			continue
		}

		result = append(result, str)
	}
	sort.Strings(result)
	return result
}
