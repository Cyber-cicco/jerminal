package utils

import (
	"runtime/debug"
	"testing"
)


func FatalExpectedActual[T comparable](expected, actual T, t *testing.T) {
	if expected != actual {
		debug.PrintStack()
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

func FatalError(err error, t *testing.T) {
	if err != nil {
		debug.PrintStack()
		t.Fatalf("Expected no error, but got %v", err)
	}
}

func FatalNoError(err error, details any, t *testing.T) {
	if err == nil {
		debug.PrintStack()
		t.Fatalf("Expected an error, but got nothing instead.\nDetails : %v", details)
	}
}

