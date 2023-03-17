package podexecutor

import (
	"errors"
	"testing"
)

func TestRequest_validate(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		r := &Request{
			Pod:       "baz",
			Namespace: "foo",
			Container: "bar",
			Command:   []string{"ls"},
		}

		err := r.validate()
		if err != nil {
			t.Fatal("error should be nil")
		}
	})

	t.Run("no pod name", func(t *testing.T) {
		r := &Request{
			Pod:       "",
			Namespace: "foo",
			Container: "bar",
			Command:   []string{"ls"},
		}

		err := r.validate()
		if err == nil {
			t.Fatal("error should not be nil")
		}

		want, got := "invalid command execute request: pod name must be specified", err.Error()
		if want != got {
			t.Fatalf("expected error msg %s, got %s", want, got)
		}

		if !errors.Is(err, ErrInvalidRequest) {
			t.Fatal("expected to see ErrInvalidRequest error")
		}
	})

	t.Run("empty command", func(t *testing.T) {
		r := &Request{
			Pod:       "baz",
			Namespace: "foo",
			Container: "bar",
			Command:   []string{},
		}

		err := r.validate()
		if err == nil {
			t.Fatal("error should not be nil")
		}

		want, got := "invalid command execute request: command slice should not be empty", err.Error()
		if want != got {
			t.Fatalf("expected error msg %s, got %s", want, got)
		}

		if !errors.Is(err, ErrInvalidRequest) {
			t.Fatal("expected to see ErrInvalidRequest error")
		}
	})
}
