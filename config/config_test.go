package config

import (
	"reflect"
	"testing"
)

func TestConfigEnvironmentOverridesPreserveOIDCMetadata(t *testing.T) {
	t.Setenv("DONETICK_JWT_SECRET", "runtime-jwt-secret-with-at-least-32-characters")
	t.Setenv("DONETICK_OAUTH2_CLIENT_ID", "runtime-client-id")
	t.Setenv("DONETICK_OAUTH2_CLIENT_SECRET", "runtime-client-secret")

	cfg := &Config{
		Jwt: JwtConfig{Secret: "yaml-jwt-secret"},
		OAuth2Config: OAuth2Config{
			ClientID:     "yaml-client-id",
			ClientSecret: "yaml-client-secret",
			AuthURL:      "https://id.example.test/authorize",
			TokenURL:     "https://id.example.test/token",
			UserInfoURL:  "https://id.example.test/userinfo",
			RedirectURL:  "https://tasks.example.test/auth/oauth2",
			Scopes:       []string{"openid", "profile", "email"},
			Name:         "Pocket ID",
		},
	}

	configEnvironmentOverrides(cfg)

	if cfg.Jwt.Secret != "runtime-jwt-secret-with-at-least-32-characters" {
		t.Fatalf("JWT secret was not overridden")
	}
	if cfg.OAuth2Config.ClientID != "runtime-client-id" || cfg.OAuth2Config.ClientSecret != "runtime-client-secret" {
		t.Fatalf("OIDC credentials were not overridden")
	}
	if cfg.OAuth2Config.AuthURL != "https://id.example.test/authorize" || cfg.OAuth2Config.Name != "Pocket ID" {
		t.Fatalf("OIDC metadata from YAML was not preserved")
	}
	if !reflect.DeepEqual(cfg.OAuth2Config.Scopes, []string{"openid", "profile", "email"}) {
		t.Fatalf("OIDC scopes from YAML were not preserved: %v", cfg.OAuth2Config.Scopes)
	}
}
