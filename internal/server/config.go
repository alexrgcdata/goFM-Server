package server

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// Config describes the server's listener, accepted bearer tokens, and routes.
type Config struct {
	Address     string        `json:"address"`
	BasePath    string        `json:"base_path,omitempty"`
	Tokens      []string      `json:"tokens"`
	AdminTokens []string      `json:"admin_tokens"`
	CORSOrigins []string      `json:"cors_origins"`
	Routes      []Route       `json:"routes"`
	Logs        LogConfig     `json:"logs"`
	Storage     StorageConfig `json:"storage"`
}

type StorageConfig struct {
	DBFile string `json:"db_file"`
}

// LogConfig controls bounded request history. Encryption is enabled by setting
// encryption_key_env to the name of an environment variable containing a
// base64-encoded 32-byte AES key.
type LogConfig struct {
	MaxEntries         int    `json:"max_entries"`
	File               string `json:"file"`
	EncryptionKeyEnv   string `json:"encryption_key_env"`
	CaptureBodyPreview bool   `json:"capture_body_preview"`
}

type Route struct {
	ID      string     `json:"id,omitempty"`
	Path    string     `json:"path"`
	Methods []string   `json:"methods"`
	Target  Target     `json:"target"`
	Hooks   HookConfig `json:"hooks,omitempty"`
}

// HookConfig names future lifecycle hooks. v1.0.2 records these names and any
// fireGoHookAfter request flag but deliberately performs no action.
type HookConfig struct {
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

type Target struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

func (c *Config) Validate() error {
	if c.Address == "" {
		c.Address = ":8080"
	}
	if c.BasePath != "" {
		c.BasePath = "/" + strings.Trim(c.BasePath, "/")
		if c.BasePath == "/" {
			return fmt.Errorf("base_path must be empty or a non-root path beginning with /")
		}
	}
	for i := range c.Tokens {
		c.Tokens[i] = strings.TrimSpace(c.Tokens[i])
		if c.Tokens[i] == "" {
			return fmt.Errorf("tokens[%d] is empty", i)
		}
	}
	for i := range c.AdminTokens {
		c.AdminTokens[i] = strings.TrimSpace(c.AdminTokens[i])
		if c.AdminTokens[i] == "" {
			return fmt.Errorf("admin_tokens[%d] is empty", i)
		}
	}
	if len(c.AdminTokens) == 0 {
		c.AdminTokens = append([]string(nil), c.Tokens...)
	}
	if c.Logs.MaxEntries == 0 {
		c.Logs.MaxEntries = 50
	}
	if c.Logs.MaxEntries < 1 || c.Logs.MaxEntries > 10000 {
		return fmt.Errorf("logs.max_entries must be between 1 and 10000")
	}
	for i := range c.CORSOrigins {
		origin := strings.TrimSpace(c.CORSOrigins[i])
		parsed, err := url.Parse(origin)
		if err != nil || origin == "" || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return fmt.Errorf("cors_origins[%d] must be an http(s) origin without a path or credentials", i)
		}
		c.CORSOrigins[i] = origin
	}
	seen := make(map[string]struct{})
	for i := range c.Routes {
		route := &c.Routes[i]
		if route.ID == "" {
			route.ID = fmt.Sprintf("route-%d", i+1)
		}
		if route.Hooks.Before == "" {
			route.Hooks.Before = "fireGoHookBefore"
		}
		if route.Hooks.After == "" {
			route.Hooks.After = "fireGoHookAfter"
		}
		if !strings.HasPrefix(route.Path, "/") || route.Path == "/" {
			return fmt.Errorf("routes[%d].path must be a non-root path beginning with /", i)
		}
		if len(route.Methods) == 0 {
			return fmt.Errorf("routes[%d].methods is required", i)
		}
		if route.Target.Type != "http" {
			return fmt.Errorf("routes[%d].target.type must be http", i)
		}
		target, err := url.Parse(route.Target.URL)
		if err != nil || target.Scheme == "" || target.Host == "" || (target.Scheme != "http" && target.Scheme != "https") || target.User != nil {
			return fmt.Errorf("routes[%d].target.url must be an absolute http(s) URL without credentials", i)
		}
		for _, method := range route.Methods {
			method = strings.ToUpper(strings.TrimSpace(method))
			if method == "" {
				return fmt.Errorf("routes[%d] includes an empty method", i)
			}
			key := method + " " + route.Path
			if _, exists := seen[key]; exists {
				return fmt.Errorf("duplicate route %s", key)
			}
			seen[key] = struct{}{}
		}
	}
	return nil
}
