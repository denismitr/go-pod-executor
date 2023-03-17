package podexecutor_test

import (
	"context"
	"github.com/denismitr/podexecutor"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestCommandExecutor_Execute(t *testing.T) {
	masterURL := os.Getenv("K8S_MASTER_URL")
	kubeConfig := os.Getenv("K8S_CONFIG")

	if masterURL == "" {
		t.Fatalf("master url cannot be empty")
	}

	if kubeConfig == "" {
		t.Fatalf("kube config cannot be empty")
	}

	t.Run("integration testing of executing date in nginx", func(t *testing.T) {

		// todo: return also an error
		executor := podexecutor.NewCommandExecutor(masterURL, kubeConfig)

		output, err := executor.Execute(
			context.TODO(),
			"nginx",
			"nginx",
			"executor",
			[]string{"ls", "-a"},
		)
		if err != nil {
			t.Errorf("executor returned an error: %s", err.Error())
		}

		if output == "" {
			t.Errorf("error: executor output shpuld not be empty")
		}

		files := outputToSlice(output)
		want := []string{
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

		sort.Strings(want)
		sort.Strings(files)
		if !reflect.DeepEqual(files, want) {
			t.Errorf("error: executor output expected to be %+v, got %+v", want, files)
		}
	})
}

func outputToSlice(out podexecutor.Output) []string {
	slice := strings.Split(string(out), "\n")
	result := make([]string, 0, len(slice))
	for i := range slice {
		str := strings.TrimSpace(slice[i])
		str = strings.ReplaceAll(str, " ", "")
		if str == "" || str == " " {
			continue
		}

		result = append(result, str)
	}
	return result
}
