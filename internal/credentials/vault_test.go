package credentials

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVaultEncryptsAndReloadsCredentials(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOFM_TEST_VAULT_KEY", base64.StdEncoding.EncodeToString(key))
	path := filepath.Join(t.TempDir(), "secrets", "credentials.enc")
	vault, err := Open(path, "GOFM_TEST_VAULT_KEY")
	if err != nil {
		t.Fatal(err)
	}
	if err := vault.Put("primary", Entry{Username: "fm-user", Password: "fm-secret"}); err != nil {
		t.Fatal(err)
	}
	if err := vault.Put("secondary", Entry{Username: "fm-user-2", Password: "fm-secret-2"}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "fm-user") || strings.Contains(string(raw), "fm-secret") {
		t.Fatal("vault contains plaintext credentials")
	}
	reloaded, err := Open(path, "GOFM_TEST_VAULT_KEY")
	if err != nil {
		t.Fatal(err)
	}
	credential, err := reloaded.Credentials(context.Background(), "primary")
	if err != nil {
		t.Fatal(err)
	}
	if credential.Username != "fm-user" || credential.Password != "fm-secret" {
		t.Fatal("reloaded credential did not match")
	}
	if len(reloaded.Names()) != 2 {
		t.Fatalf("reloaded names = %#v", reloaded.Names())
	}
}
