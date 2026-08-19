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
	for i := range c.Tokens {
		c.Tokens[i] = strings.TrimSpace(c.Tokens[i])
		if c.Tokens[i] == "" {
			return fmt.Errorf("tokens[%d] is empty", i)
		}
	}
	for i := range c.Credentials {
		c.Credentials[i].Username = strings.TrimSpace(c.Credentials[i].Username)
		if c.Credentials[i].Username == "" || c.Credentials[i].Password == "" {
			return fmt.Errorf("credentials[%d] requires username and password", i)
		}
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
	for i, route := range c.Routes {
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
