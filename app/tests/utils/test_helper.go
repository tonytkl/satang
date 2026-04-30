package test_utils

import (
	"net/http"
	"testing"
)

func AssertHeader(t *testing.T, request *http.Request, key, want string) {
	t.Helper()

	if got := request.Header.Get(key); got != want {
		t.Fatalf("header %s = %q, want %q", key, got, want)
	}
}

func AssertEqual(t *testing.T, got, want any) {
	t.Helper()

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
