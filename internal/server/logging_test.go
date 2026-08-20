package server

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"testing"
)

func TestEncryptedLogStoreRoundTripsAndIsBounded(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOFM_TEST_LOG_KEY", base64.StdEncoding.EncodeToString(key))
	path := t.TempDir() + string(os.PathSeparator) + "requests.enc"
	config := LogConfig{MaxEntries: 2, File: path, EncryptionKeyEnv: "GOFM_TEST_LOG_KEY"}
	store, err := newLogStore(config)
	if err != nil {
		t.Fatal(err)
	}
	store.add(LogEntry{ID: "one", Path: "/one"})
	store.add(LogEntry{ID: "two", Path: "/two"})
	store.add(LogEntry{ID: "three", Path: "/three"})
	reloaded, err := newLogStore(config)
	if err != nil {
		t.Fatal(err)
	}
	entries := reloaded.list()
	if len(entries) != 2 || entries[0].ID != "three" || entries[1].ID != "two" {
		t.Fatalf("unexpected entries: %#v", entries)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" || string(data) == "/three" {
		t.Fatal("encrypted log file was not written")
	}
}
