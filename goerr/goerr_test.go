package goerr_test

import (
	"errors"
	"testing"

	"github.com/l00pss/helpme/goerr"
)

func TestNewGoErr(t *testing.T) {
	err := errors.New("test error")

	t.Run("runtime error", func(t *testing.T) {
		goErr := goerr.WrapRuntimeErr(err)
		if goErr == nil {
			t.Fatal("expected non-nil GoErr")
		}
		if !goErr.IsRuntime() {
			t.Error("expected runtime error")
		}
		if goErr.Error() != "test error" {
			t.Errorf("expected 'test error', got '%s'", goErr.Error())
		}
	})

	t.Run("non-runtime error", func(t *testing.T) {
		goErr := goerr.WrapNonRuntimeErr(err)
		if goErr == nil {
			t.Fatal("expected non-nil GoErr")
		}
		if goErr.IsRuntime() {
			t.Error("expected non-runtime error")
		}
		if goErr.Error() != "test error" {
			t.Errorf("expected 'test error', got '%s'", goErr.Error())
		}
	})
}

func TestGoErr_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	goErr := goerr.WrapRuntimeErr(originalErr)

	unwrapped := goErr.Unwrap()
	if unwrapped != originalErr {
		t.Error("unwrapped error does not match original")
	}
	if unwrapped.Error() != "original error" {
		t.Errorf("expected 'original error', got '%s'", unwrapped.Error())
	}
}

func TestIsGoErr(t *testing.T) {
	t.Run("is GoErr", func(t *testing.T) {
		err := errors.New("test")
		goErr := goerr.WrapRuntimeErr(err)
		if !goerr.IsGoErr(goErr) {
			t.Error("expected IsGoErr to return true")
		}
	})

	t.Run("is not GoErr", func(t *testing.T) {
		err := errors.New("test")
		if goerr.IsGoErr(err) {
			t.Error("expected IsGoErr to return false")
		}
	})

	t.Run("nil error", func(t *testing.T) {
		if goerr.IsGoErr(nil) {
			t.Error("expected IsGoErr to return false for nil")
		}
	})
}

func TestGoErr_Error(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected string
	}{
		{"simple message", "test error", "test error"},
		{"empty message", "", ""},
		{"special chars", "error: test @#$", "error: test @#$"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			goErr := goerr.WrapRuntimeErr(err)
			if goErr.Error() != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, goErr.Error())
			}
		})
	}
}

func TestGoErr_IsRuntime(t *testing.T) {
	err := errors.New("test")

	runtimeErr := goerr.WrapRuntimeErr(err)
	if !runtimeErr.IsRuntime() {
		t.Error("WrapRuntimeErr should create runtime error")
	}

	nonRuntimeErr := goerr.WrapNonRuntimeErr(err)
	if nonRuntimeErr.IsRuntime() {
		t.Error("WrapNonRuntimeErr should create non-runtime error")
	}
}
