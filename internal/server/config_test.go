package server

import "testing"

func TestConfigLoadsSeparateTokensFromEnvironment(t *testing.T) {
	t.Setenv("GOFM_TEST_APP_TOKEN", "application-secret")
	t.Setenv("GOFM_TEST_ADMIN_TOKEN", "administrator-secret")
	config := Config{TokenEnv: "GOFM_TEST_APP_TOKEN", AdminTokenEnv: "GOFM_TEST_ADMIN_TOKEN"}
	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}
	if config.Tokens[0] != "application-secret" || config.AdminTokens[0] != "administrator-secret" {
		t.Fatal("environment tokens were not loaded")
	}
}

func TestConfigRejectsMissingTokenEnvironmentVariable(t *testing.T) {
	config := Config{TokenEnv: "GOFM_TEST_MISSING_TOKEN"}
	if err := config.Validate(); err == nil {
		t.Fatal("expected missing environment token error")
	}
}

func TestConfigRequiresDifferentApplicationAndAdminTokens(t *testing.T) {
	config := Config{Tokens: []string{"same"}, AdminTokens: []string{"same"}}
	if err := config.Validate(); err == nil {
		t.Fatal("expected shared token rejection")
	}
}

func TestConfigRejectsInvalidRouteAuthorization(t *testing.T) {
	config := Config{Tokens: []string{"app"}, AdminTokens: []string{"admin"}, Routes: []Route{{Path: "/bad", Methods: []string{"GET"}, Auth: "public", Target: Target{Type: "http", URL: "https://example.com"}}}}
	if err := config.Validate(); err == nil {
		t.Fatal("expected invalid route authorization error")
	}
}
