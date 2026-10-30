package main

import (
	"os"
	"path/filepath"
	"testing"

	"gofm-server/internal/api"
)

func TestInitConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.local.json")
	if err := initConfig(path); err != nil {
		t.Fatal(err)
	}
	c, err := api.LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Token) != 64 || c.Address != "127.0.0.1:8080" {
		t.Fatal("unexpected generated config")
	}
	before, _ := os.ReadFile(path)
	if err := initConfig(path); err == nil {
		t.Fatal("existing config overwritten")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("existing config changed")
	}
}
