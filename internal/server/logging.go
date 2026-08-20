package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// LogEntry is intentionally metadata-first. Bodies are omitted unless the
// operator explicitly enables short response previews.
type LogEntry struct {
	ID                 string `json:"id"`
	StartedAt          string `json:"started_at"`
	DurationMS         int64  `json:"duration_ms"`
	Method             string `json:"method"`
	Path               string `json:"path"`
	Status             int    `json:"status"`
	Route              string `json:"route,omitempty"`
	Upstream           string `json:"upstream,omitempty"`
	Outcome            string `json:"outcome"`
	Error              string `json:"error,omitempty"`
	ResponsePreview    string `json:"response_preview,omitempty"`
	HookAfterRequested bool   `json:"hook_after_requested,omitempty"`
}

type logStore struct {
	mu      sync.RWMutex
	max     int
	file    string
	box     cipher.AEAD
	entries []LogEntry
}

func newLogStore(config LogConfig) (*logStore, error) {
	store := &logStore{max: config.MaxEntries, file: config.File}
	if config.EncryptionKeyEnv != "" {
		encoded := strings.TrimSpace(os.Getenv(config.EncryptionKeyEnv))
		key, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil || len(key) != 32 {
			return nil, fmt.Errorf("logs encryption key must be base64-encoded 32 bytes from %s", config.EncryptionKeyEnv)
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, fmt.Errorf("create log cipher: %w", err)
		}
		store.box, err = cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("create log AEAD: %w", err)
		}
	}
	if store.file != "" && store.box == nil {
		return nil, fmt.Errorf("logs.file requires logs.encryption_key_env")
	}
	if store.file != "" {
		_ = store.load()
	}
	return store, nil
}

func (s *logStore) add(entry LogEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, entry)
	if len(s.entries) > s.max {
		s.entries = s.entries[len(s.entries)-s.max:]
	}
	if s.file != "" {
		_ = s.persistLocked()
	}
}

func (s *logStore) list() []LogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]LogEntry, len(s.entries))
	copy(result, s.entries)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

func (s *logStore) load() error {
	data, err := os.ReadFile(s.file)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(data) < s.box.NonceSize() {
		return fmt.Errorf("encrypted log file is truncated")
	}
	plain, err := s.box.Open(nil, data[:s.box.NonceSize()], data[s.box.NonceSize():], nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(plain, &s.entries)
}

func (s *logStore) persistLocked() error {
	plain, err := json.Marshal(s.entries)
	if err != nil {
		return err
	}
	nonce := make([]byte, s.box.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	data := s.box.Seal(nonce, nonce, plain, nil)
	return os.WriteFile(s.file, data, 0600)
}

func requestID(r *http.Request) string {
	// This is intentionally a short stable identifier for logs and support.
	hash := sha256.Sum256([]byte(r.Method + " " + r.URL.RequestURI() + " " + time.Now().UTC().String()))
	return base64.RawURLEncoding.EncodeToString(hash[:8])
}
