package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// stubTokenSource hands out sequential access tokens with a fixed lifetime.
type stubTokenSource struct {
	calls    int
	lifetime time.Duration
	err      error
}

func (s *stubTokenSource) Token() (*oauth2.Token, error) {
	if s.err != nil {
		return nil, s.err
	}
	s.calls++
	return &oauth2.Token{
		AccessToken: fmt.Sprintf("ya29.access-%d", s.calls),
		Expiry:      time.Now().Add(s.lifetime),
	}, nil
}

func countingPrincipal(email string, calls *int) func(context.Context, *oauth2.Token) (string, error) {
	return func(context.Context, *oauth2.Token) (string, error) {
		*calls++
		return email, nil
	}
}

func decodeSegment(t *testing.T, seg string) []byte {
	t.Helper()
	b, err := base64.RawURLEncoding.DecodeString(seg)
	require.NoError(t, err, "segments must be unpadded base64url")
	return b
}

// writeFakeADC points ADC at a service-account key file. Loading credentials
// does not touch the network, so the provider can be built offline.
func writeFakeADC(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "adc.json")
	body := `{"type":"service_account","project_id":"p","private_key_id":"k",` +
		`"private_key":"-----BEGIN PRIVATE KEY-----\nMIIB\n-----END PRIVATE KEY-----\n",` +
		`"client_email":"kafkactl@p.iam.gserviceaccount.com","client_id":"1",` +
		`"token_uri":"https://oauth2.googleapis.com/token"}`
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", path)
}

func TestBuildGCPToken_Format(t *testing.T) {
	iat := time.Unix(1_700_000_000, 0)
	exp := iat.Add(time.Hour)
	tok := buildGCPToken("ya29.secret", "sa@proj.iam.gserviceaccount.com", exp, iat)

	parts := strings.Split(tok, ".")
	require.Len(t, parts, 3)
	assert.NotContains(t, tok, "=")

	var header map[string]string
	require.NoError(t, json.Unmarshal(decodeSegment(t, parts[0]), &header))
	assert.Equal(t, map[string]string{"typ": "JWT", "alg": "GOOG_OAUTH2_TOKEN"}, header)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(decodeSegment(t, parts[1]), &payload))
	assert.Equal(t, float64(exp.Unix()), payload["exp"])
	assert.Equal(t, float64(iat.Unix()), payload["iat"])
	assert.Equal(t, "Google", payload["iss"])
	assert.Equal(t, "sa@proj.iam.gserviceaccount.com", payload["sub"])

	assert.Equal(t, "ya29.secret", string(decodeSegment(t, parts[2])),
		"the signature segment carries the Google access token")
}

func TestGCPTokenProvider_CachesUntilNearExpiry(t *testing.T) {
	src := &stubTokenSource{lifetime: time.Hour}
	principalCalls := 0
	p := &gcpTokenProvider{source: src, principal: countingPrincipal("me@example.com", &principalCalls), now: time.Now}

	first, err := p.Token()
	require.NoError(t, err)
	second, err := p.Token()
	require.NoError(t, err)

	assert.Equal(t, first.Token, second.Token)
	assert.Equal(t, 1, src.calls, "token must be cached until near expiry")
}

func TestGCPTokenProvider_RefreshesAfterBufferAndKeepsPrincipal(t *testing.T) {
	// A lifetime inside the refresh buffer forces a refresh on the next call.
	src := &stubTokenSource{lifetime: time.Second}
	principalCalls := 0
	p := &gcpTokenProvider{source: src, principal: countingPrincipal("me@example.com", &principalCalls), now: time.Now}

	_, err := p.Token()
	require.NoError(t, err)
	tok, err := p.Token()
	require.NoError(t, err)

	assert.Equal(t, 2, src.calls)
	assert.Equal(t, "ya29.access-2", string(decodeSegment(t, strings.Split(tok.Token, ".")[2])))
	assert.Equal(t, 1, principalCalls, "the principal is resolved once and cached")
}

func TestGCPTokenProvider_ZeroExpiryDefaultsToOneHour(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	p := &gcpTokenProvider{
		source:    oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "static"}),
		principal: func(context.Context, *oauth2.Token) (string, error) { return "x@y", nil },
		now:       func() time.Time { return now },
	}
	tok, err := p.Token()
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(decodeSegment(t, strings.Split(tok.Token, ".")[1]), &payload))
	assert.Equal(t, float64(now.Add(time.Hour).Unix()), payload["exp"])
}

func TestGCPTokenProvider_Errors(t *testing.T) {
	t.Run("token source failure", func(t *testing.T) {
		p := &gcpTokenProvider{source: &stubTokenSource{err: errors.New("adc expired")}, now: time.Now}
		_, err := p.Token()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to obtain Google access token")
		assert.Contains(t, err.Error(), "adc expired")
	})
	t.Run("principal failure", func(t *testing.T) {
		p := &gcpTokenProvider{
			source:    &stubTokenSource{lifetime: time.Hour},
			principal: func(context.Context, *oauth2.Token) (string, error) { return "", errors.New("no email") },
			now:       time.Now,
		}
		_, err := p.Token()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve Google principal")
	})
}

func TestEmailFromCredentialsJSON(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"empty", "", ""},
		{"invalid json", "{", ""},
		{"service account key", `{"type":"service_account","client_email":"sa@p.iam.gserviceaccount.com"}`, "sa@p.iam.gserviceaccount.com"},
		{"impersonated", `{"type":"impersonated_service_account","service_account_impersonation_url":"https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/kafka@p.iam.gserviceaccount.com:generateAccessToken"}`, "kafka@p.iam.gserviceaccount.com"},
		{"authorized user", `{"type":"authorized_user","client_id":"x"}`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, emailFromCredentialsJSON([]byte(tt.raw)))
		})
	}
}

func TestResolveGCPPrincipal_KnownEmailWins(t *testing.T) {
	email, err := resolveGCPPrincipal(context.Background(), "sa@p", &oauth2.Token{AccessToken: "t"})
	require.NoError(t, err)
	assert.Equal(t, "sa@p", email)
}

func TestTokenInfoEmail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Empty(t, r.URL.RawQuery, "the access token must not travel in the URL")
		_ = r.ParseForm()
		switch r.PostForm.Get("access_token") {
		case "good":
			fmt.Fprint(w, `{"email":"user@example.com"}`)
		case "noemail":
			fmt.Fprint(w, `{"scope":"cloud-platform"}`)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	email, err := tokenInfoEmail(context.Background(), srv.Client(), srv.URL, "good")
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", email)

	_, err = tokenInfoEmail(context.Background(), srv.Client(), srv.URL, "noemail")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "userinfo.email")

	_, err = tokenInfoEmail(context.Background(), srv.Client(), srv.URL, "bad")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 400")
}

func TestLoadTokenProviderPlugin_GCP(t *testing.T) {
	writeFakeADC(t)

	provider, err := LoadTokenProviderPlugin("gcp", nil, nil)
	require.NoError(t, err)
	p, ok := provider.(*gcpTokenProvider)
	require.True(t, ok, "'gcp' must resolve to the built-in provider, not an external plugin")
	email, err := p.principal(context.Background(), &oauth2.Token{})
	require.NoError(t, err)
	assert.Equal(t, "kafkactl@p.iam.gserviceaccount.com", email, "principal comes from the key file")
}

func TestNewGCPTokenProvider_Options(t *testing.T) {
	writeFakeADC(t)

	provider, err := newGCPTokenProvider(map[string]any{"principal": "override@p.iam.gserviceaccount.com"})
	require.NoError(t, err)
	email, err := provider.(*gcpTokenProvider).principal(context.Background(), &oauth2.Token{})
	require.NoError(t, err)
	assert.Equal(t, "override@p.iam.gserviceaccount.com", email, "configured principal wins")

	_, err = newGCPTokenProvider(map[string]any{"principal": 42})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "option 'principal' must be a string")
}

func TestNewGCPTokenProvider_MissingADC(t *testing.T) {
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", filepath.Join(t.TempDir(), "missing.json"))

	_, err := newGCPTokenProvider(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Application Default Credentials")
}
