package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/oauth2/clientcredentials"
)

func TestSchemaRegistryOAuthOverTLS(t *testing.T) {
	var tokenRequests int

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			tokenRequests++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"abc","token_type":"Bearer","expires_in":3600}`))
		case "/subjects":
			if r.Header.Get("Authorization") != "Bearer abc" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`["a-value"]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	clientContext := &ClientContext{}
	clientContext.SchemaRegistry.URL = server.URL
	clientContext.SchemaRegistry.TLS = TLSConfig{Enabled: true, Insecure: true}
	clientContext.SchemaRegistry.OAuth = clientcredentials.Config{
		ClientID:     "id",
		ClientSecret: "secret",
		TokenURL:     server.URL + "/token",
	}

	registry, err := CreateCachingSchemaRegistry(clientContext)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	subjects, err := registry.GetSubjects()
	if err != nil {
		t.Fatalf("expected token fetch and registry call to succeed over TLS: %v", err)
	}

	if len(subjects) != 1 || tokenRequests != 1 {
		t.Fatalf("expected 1 subject and 1 token request, got %d and %d", len(subjects), tokenRequests)
	}
}

func TestValidateSchemaRegistryOAuth(t *testing.T) {
	full := clientcredentials.Config{ClientID: "id", ClientSecret: "secret", TokenURL: "https://token"}

	testCases := []struct {
		name    string
		config  SchemaRegistryConfig
		wantErr bool
	}{
		{"unset", SchemaRegistryConfig{}, false},
		{"complete", SchemaRegistryConfig{OAuth: full}, false},
		{"missing token url", SchemaRegistryConfig{OAuth: clientcredentials.Config{ClientID: "id", ClientSecret: "secret"}}, true},
		{"secret without client id", SchemaRegistryConfig{OAuth: clientcredentials.Config{ClientSecret: "secret"}}, true},
		{"with basic auth", SchemaRegistryConfig{OAuth: full, Username: "user"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSchemaRegistryOAuth(tc.config)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v, got %v", tc.wantErr, err)
			}
		})
	}
}
