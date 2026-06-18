package zepstore

import (
	"context"
	"errors"
	"strings"
	"testing"

	zep "github.com/getzep/zep-go/v3"
)

func TestIsNotFound(t *testing.T) {
	if !isNotFound(&zep.NotFoundError{}) {
		t.Fatal("zep.NotFoundError should be recognized")
	}
	if isNotFound(errors.New("network failed")) {
		t.Fatal("plain error should not be recognized as not found")
	}
}

func TestDeleteGraphItemRejectsUnknownKind(t *testing.T) {
	store := New("test-key")
	err := store.DeleteGraphItem(context.Background(), "user", "uuid")
	if err == nil {
		t.Fatal("expected error for unknown graph item kind")
	}
	if !strings.Contains(err.Error(), "unknown graph item kind") {
		t.Fatalf("unexpected error: %v", err)
	}
}
