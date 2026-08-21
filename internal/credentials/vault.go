package credentials

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"gofm-server/internal/filemaker"
)

type Entry struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
}

type Vault struct {
	mu      sync.RWMutex
	file    string
	box     cipher.AEAD
	entries map[string]Entry
}

func Open(file, keyEnv string) (*Vault, error) {
	encoded := strings.TrimSpace(os.Getenv(keyEnv))
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("credential encryption key must be base64-encoded 32 bytes from %s", keyEnv)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create credential cipher: %w", err)
	}
	box, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create credential AEAD: %w", err)
	}
	vault := &Vault{file: file, box: box, entries: make(map[string]Entry)}
	if err := vault.load(); err != nil {
		return nil, err
	}
	return vault, nil
}

func (v *Vault) Put(name string, entry Entry) error {
	name = strings.TrimSpace(name)
	if name == "" || strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("credential name must be non-empty and cannot contain slashes")
	}
	if entry.Username == "" || (entry.Password == "" && entry.APIKey == "") {
		return fmt.Errorf("username and a password or API key are required")
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	v.entries[name] = entry
	return v.persistLocked()
}

func (v *Vault) Credentials(_ context.Context, name string) (filemaker.Credential, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	entry, ok := v.entries[name]
	if !ok {
		return filemaker.Credential{}, fmt.Errorf("credential %q not found", name)
	}
	return filemaker.Credential{Username: entry.Username, Password: entry.Password, APIKey: entry.APIKey}, nil
}

func (v *Vault) Names() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	names := make([]string, 0, len(v.entries))
	for name := range v.entries {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (v *Vault) load() error {
	data, err := os.ReadFile(v.file)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read credential vault: %w", err)
	}
	if len(data) < v.box.NonceSize() {
		return fmt.Errorf("credential vault is truncated")
	}
	plain, err := v.box.Open(nil, data[:v.box.NonceSize()], data[v.box.NonceSize():], nil)
	if err != nil {
		return fmt.Errorf("decrypt credential vault: %w", err)
	}
	if err := json.Unmarshal(plain, &v.entries); err != nil {
		return fmt.Errorf("parse credential vault: %w", err)
	}
	return nil
}

func (v *Vault) persistLocked() error {
	plain, err := json.Marshal(v.entries)
	if err != nil {
		return err
	}
	nonce := make([]byte, v.box.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	data := v.box.Seal(nonce, nonce, plain, nil)
	if err := os.MkdirAll(filepath.Dir(v.file), 0700); err != nil {
		return fmt.Errorf("create credential directory: %w", err)
	}
	temporary := v.file + ".tmp"
	if err := os.WriteFile(temporary, data, 0600); err != nil {
		return fmt.Errorf("write credential vault: %w", err)
	}
	if err := os.Rename(temporary, v.file); err != nil {
		return fmt.Errorf("replace credential vault: %w", err)
	}
	return nil
}
