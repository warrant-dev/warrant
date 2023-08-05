package wookie_test

import (
	"context"
	"testing"

	"github.com/warrant-dev/warrant/pkg/wookie"
)

func TestBasicSerialization(t *testing.T) {
	ctx := wookie.WithLatest(context.Background())
	if !wookie.ContainsLatest(ctx) {
		t.Fatalf("expected ctx to contain 'latest' wookie")
	}

	ctx = context.Background()
	if wookie.ContainsLatest(ctx) {
		t.Fatalf("expected ctx to not contain 'latest' wookie")
	}
}
