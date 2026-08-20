package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestStorePersistsAndBoundsRecentLogs(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "data", "gofm.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	for i, id := range []string{"one", "two", "three"} {
		if err := store.Append(context.Background(), RequestRecord{ID: id, StartedAt: time.Now().Add(time.Duration(i) * time.Second), Method: "GET", Path: "/api/test", Status: 200, Outcome: "success"}, 2); err != nil {
			t.Fatal(err)
		}
	}
	records, err := store.Recent(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("records = %d, want 2", len(records))
	}
}
