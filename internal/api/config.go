// Package api implements the small, database-free action dispatcher.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Address         string `json:"address"`
	Token           string `json:"token"`
	MaxRequestBytes int64  `json:"max_request_bytes"`
}

func LoadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()
	var c Config
	d := json.NewDecoder(io.LimitReader(f, 65537))
	d.DisallowUnknownFields()
	if err := d.Decode(&c); err != nil {
		return Config{}, errors.New("config must be a JSON object with address, token, and max_request_bytes")
	}
	if d.Decode(new(any)) != io.EOF {
		return Config{}, errors.New("config must contain exactly one JSON object")
	}
	return c, c.Validate()
}

func (c Config) Validate() error {
	_, port, err := net.SplitHostPort(c.Address)
	if err != nil {
		return errors.New("address must be host:port, for example 127.0.0.1:8080")
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return errors.New("address port must be between 1 and 65535")
	}
	if len(c.Token) < 32 || len(c.Token) > 256 || strings.ContainsAny(c.Token, " \t\r\n") {
		return errors.New("token must be 32-256 characters without whitespace; generate one with -init-config")
	}
	if c.MaxRequestBytes < 1024 || c.MaxRequestBytes > 1<<20 {
		return errors.New("max_request_bytes must be between 1024 and 1048576")
	}
	return nil
}
