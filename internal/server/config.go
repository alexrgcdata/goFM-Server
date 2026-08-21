package server

import (
<<<<<<< HEAD
	"encoding/base64"
=======
>>>>>>> 0d8871589256cb66840deaca1805331e2759ccc6
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
<<<<<<< HEAD

	"gofm-server/internal/filemaker"
=======
>>>>>>> 0d8871589256cb66840deaca1805331e2759ccc6
)

// Config describes the server's listener, accepted bearer tokens, and routes.
type Config struct {
<<<<<<< HEAD
	Address              string             `json:"address"`
	BasePath             string             `json:"base_path,omitempty"`
	Tokens               []string           `json:"tokens,omitempty"`
	AdminTokens          []string           `json:"admin_tokens,omitempty"`
	TokenEnv             string             `json:"token_env,omitempty"`
	AdminTokenEnv        string             `json:"admin_token_env,omitempty"`
	CORSOrigins          []string           `json:"cors_origins"`
	Routes               []Route            `json:"routes"`
	Logs                 LogConfig          `json:"logs"`
	Storage              StorageConfig      `json:"storage"`
	Credentials          VaultConfig        `json:"credentials"`
	Security             SecurityConfig     `json:"security"`
	FileMakerConnections []filemaker.Target `json:"filemaker_connections,omitempty"`
}

type VaultConfig struct {
	File             string `json:"file"`
	EncryptionKeyEnv string `json:"encryption_key_env"`
}

type SecurityConfig struct {
	MaxRequestBodyBytes    int64    `json:"max_request_body_bytes"`
	RateLimitPerMinute     int      `json:"rate_limit_per_minute"`
	UpstreamTimeoutSeconds int      `json:"upstream_timeout_seconds"`
	AllowedUpstreamHosts   []string `json:"allowed_upstream_hosts,omitempty"`
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
	Auth    string     `json:"auth,omitempty"`
}

// HookConfig names future lifecycle hooks. v1.0.2 records these names and any
// fireGoHookAfter request flag but deliberately performs no action.
type HookConfig struct {
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
=======
	Address     string       `json:"address"`
	Tokens      []string     `json:"tokens"`
	Credentials []Credential `json:"credentials"`
	CORSOrigins []string     `json:"cors_origins"`
	Routes      []Route      `json:"routes"`
}

// Credential is a development login accepted by POST /auth/token.
// Production deployments should replace this with an identity provider.
type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Route struct {
	Path    string   `json:"path"`
	Methods []string `json:"methods"`
	Target  Target   `json:"target"`
>>>>>>> 0d8871589256cb66840deaca1805331e2759ccc6
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
<<<<<<< HEAD
	if c.BasePath != "" {
		c.BasePath = "/" + strings.Trim(c.BasePath, "/")
		if c.BasePath == "/" {
			return fmt.Errorf("base_path must be empty or a non-root path beginning with /")
		}
	}
	if c.TokenEnv != "" {
		token := strings.TrimSpace(os.Getenv(c.TokenEnv))
		if token == "" {
			return fmt.Errorf("application token environment variable %s is empty", c.TokenEnv)
		}
		c.Tokens = append(c.Tokens, token)
	}
	if c.AdminTokenEnv != "" {
		token := strings.TrimSpace(os.Getenv(c.AdminTokenEnv))
		if token == "" {
			return fmt.Errorf("admin token environment variable %s is empty", c.AdminTokenEnv)
		}
		c.AdminTokens = append(c.AdminTokens, token)
	}
=======
>>>>>>> 0d8871589256cb66840deaca1805331e2759ccc6
	for i := range c.Tokens {
		c.Tokens[i] = strings.TrimSpace(c.Tokens[i])
		if c.Tokens[i] == "" {
			return fmt.Errorf("tokens[%d] is empty", i)
		}
	}
<<<<<<< HEAD
	for i := range c.AdminTokens {
		c.AdminTokens[i] = strings.TrimSpace(c.AdminTokens[i])
		if c.AdminTokens[i] == "" {
			return fmt.Errorf("admin_tokens[%d] is empty", i)
		}
	}
	if len(c.Tokens) == 0 {
		return fmt.Errorf("an application token must be configured through token_env")
	}
	if len(c.AdminTokens) == 0 {
		return fmt.Errorf("a separate administrator token must be configured through admin_token_env")
	}
	for _, application := range c.Tokens {
		for _, administrator := range c.AdminTokens {
			if application == administrator {
				return fmt.Errorf("application and administrator tokens must be different")
			}
		}
	}
	if c.Security.MaxRequestBodyBytes == 0 {
		c.Security.MaxRequestBodyBytes = 1 << 20
	}
	if c.Security.MaxRequestBodyBytes < 1024 || c.Security.MaxRequestBodyBytes > 64<<20 {
		return fmt.Errorf("security.max_request_body_bytes must be between 1024 and 67108864")
	}
	if c.Security.RateLimitPerMinute == 0 {
		c.Security.RateLimitPerMinute = 120
	}
	if c.Security.RateLimitPerMinute < 1 || c.Security.RateLimitPerMinute > 100000 {
		return fmt.Errorf("security.rate_limit_per_minute must be between 1 and 100000")
	}
	if c.Security.UpstreamTimeoutSeconds == 0 {
		c.Security.UpstreamTimeoutSeconds = 30
	}
	if c.Security.UpstreamTimeoutSeconds < 1 || c.Security.UpstreamTimeoutSeconds > 300 {
		return fmt.Errorf("security.upstream_timeout_seconds must be between 1 and 300")
	}
	for i, host := range c.Security.AllowedUpstreamHosts {
		host = strings.TrimSpace(host)
		if host == "" || strings.ContainsAny(host, "/?#@") {
			return fmt.Errorf("security.allowed_upstream_hosts[%d] must be a host or host:port", i)
		}
		c.Security.AllowedUpstreamHosts[i] = host
	}
	if (c.Credentials.File == "") != (c.Credentials.EncryptionKeyEnv == "") {
		return fmt.Errorf("credentials.file and credentials.encryption_key_env must be configured together")
	}
	if c.Credentials.File != "" {
		if err := validateEncryptionKey(c.Credentials.EncryptionKeyEnv); err != nil {
			return fmt.Errorf("credentials: %w", err)
		}
	}
	if c.Logs.File != "" {
		if c.Logs.EncryptionKeyEnv == "" {
			return fmt.Errorf("logs.file requires logs.encryption_key_env")
		}
		if err := validateEncryptionKey(c.Logs.EncryptionKeyEnv); err != nil {
			return fmt.Errorf("logs: %w", err)
		}
	}
	if c.Logs.MaxEntries == 0 {
		c.Logs.MaxEntries = 50
	}
	if c.Logs.MaxEntries < 1 || c.Logs.MaxEntries > 10000 {
		return fmt.Errorf("logs.max_entries must be between 1 and 10000")
	}
=======
	for i := range c.Credentials {
		c.Credentials[i].Username = strings.TrimSpace(c.Credentials[i].Username)
		if c.Credentials[i].Username == "" || c.Credentials[i].Password == "" {
			return fmt.Errorf("credentials[%d] requires username and password", i)
		}
	}
>>>>>>> 0d8871589256cb66840deaca1805331e2759ccc6
	for i := range c.CORSOrigins {
		origin := strings.TrimSpace(c.CORSOrigins[i])
		parsed, err := url.Parse(origin)
		if err != nil || origin == "" || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return fmt.Errorf("cors_origins[%d] must be an http(s) origin without a path or credentials", i)
		}
		c.CORSOrigins[i] = origin
	}
	seen := make(map[string]struct{})
<<<<<<< HEAD
	connections := make(map[string]struct{})
	for i := range c.FileMakerConnections {
		connection := &c.FileMakerConnections[i]
		if err := connection.Validate(); err != nil {
			return err
		}
		if _, exists := connections[connection.Name]; exists {
			return fmt.Errorf("duplicate FileMaker connection %s", connection.Name)
		}
		connections[connection.Name] = struct{}{}
	}
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
		if route.Auth == "" {
			route.Auth = "application"
		}
		if route.Auth != "application" && route.Auth != "admin" {
			return fmt.Errorf("routes[%d].auth must be application or admin", i)
		}
=======
	for i, route := range c.Routes {
>>>>>>> 0d8871589256cb66840deaca1805331e2759ccc6
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
<<<<<<< HEAD

func validateEncryptionKey(environment string) error {
	encoded := strings.TrimSpace(os.Getenv(environment))
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(key) != 32 {
		return fmt.Errorf("%s must contain a base64-encoded 32-byte key", environment)
	}
	return nil
}
=======
>>>>>>> 0d8871589256cb66840deaca1805331e2759ccc6
