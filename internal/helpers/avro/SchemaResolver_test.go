package avro

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSchemaResolverResolveSourceReadsLocalFile(t *testing.T) {
	schemaPath := filepath.Join(t.TempDir(), "message.avsc")
	if err := os.WriteFile(schemaPath, []byte(testRecordSchema), 0o600); err != nil {
		t.Fatalf("failed to write test schema %q: %v", schemaPath, err)
	}

	resolver := NewSchemaResolver(Standard, true)
	codec, err := resolver.ResolveSource(schemaPath)
	if err != nil {
		t.Fatalf("ResolveSource(): unexpected error: %v", err)
	}
	if codec.Schema() != testRecordSchema {
		t.Fatalf("resolved schema = %q, want %q", codec.Schema(), testRecordSchema)
	}
}

func TestSchemaResolverResolveSourceTreatsWindowsDrivePathAsFile(t *testing.T) {
	if filepath.Separator != '\\' {
		t.Skip("Windows drive path behavior can only be exercised on Windows")
	}

	schemaPath := filepath.Join(t.TempDir(), "windows-path.avsc")
	volumeName := filepath.VolumeName(schemaPath)
	if len(volumeName) != 2 || volumeName[1] != ':' {
		t.Skipf("temporary path %q does not have a Windows drive letter", schemaPath)
	}
	if err := os.WriteFile(schemaPath, []byte(testRecordSchema), 0o600); err != nil {
		t.Fatalf("failed to write test schema %q: %v", schemaPath, err)
	}

	resolver := NewSchemaResolver(Standard, true)
	codec, err := resolver.ResolveSource(schemaPath)
	if err != nil {
		t.Fatalf("ResolveSource(%q): unexpected error: %v", schemaPath, err)
	}
	if codec.Schema() != testRecordSchema {
		t.Fatalf("resolved schema = %q, want %q", codec.Schema(), testRecordSchema)
	}
}

func TestSchemaResolverResolveSourceUsesHTTPForURL(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		if request.Method != http.MethodGet {
			t.Errorf("request method = %s, want GET", request.Method)
		}
		_, _ = response.Write([]byte(testRecordSchema))
	}))
	defer server.Close()

	resolver := NewSchemaResolver(Standard, true)
	codec, err := resolver.ResolveSource(" \t" + server.URL + "/schema?version=1#fragment\r\n")
	if err != nil {
		t.Fatalf("ResolveSource(): unexpected error: %v", err)
	}
	if codec.Schema() != testRecordSchema {
		t.Fatalf("resolved schema = %q, want %q", codec.Schema(), testRecordSchema)
	}
	if requests.Load() != 1 {
		t.Fatalf("HTTP request count = %d, want 1", requests.Load())
	}
}

func TestSchemaResolverResolveSourceRejectsMalformedHTTPURLWithoutDisclosingIt(t *testing.T) {
	resolver := NewSchemaResolver(Standard, true)
	_, err := resolver.ResolveSource("https://user:password@example.invalid/%zz?token=secret#fragment")
	if err == nil {
		t.Fatal("ResolveSource() returned nil error for a malformed HTTP URL")
	}
	const wantErrPart = "invalid avro schema URI"
	if !strings.Contains(err.Error(), wantErrPart) {
		t.Fatalf("ResolveSource() error = %q, want it to contain %q", err, wantErrPart)
	}
	for _, secret := range []string{"user", "password", "token=secret", "fragment"} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("ResolveSource() error %q disclosed %q", err, secret)
		}
	}
}

func TestSchemaResolverResolveURIRejectsUnsuccessfulHTTPResponses(t *testing.T) {
	for _, statusCode := range []int{
		http.StatusMovedPermanently,
		http.StatusNotModified,
		http.StatusBadRequest,
		http.StatusInternalServerError,
	} {
		t.Run(http.StatusText(statusCode), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
				response.WriteHeader(statusCode)
				_, _ = response.Write([]byte("response-body-must-not-appear"))
			}))
			defer server.Close()

			resolver := NewSchemaResolver(Standard, false)
			_, err := resolver.ResolveURI(server.URL + "/schema?token=query-secret#fragment-secret")
			if err == nil {
				t.Fatalf("ResolveURI() returned nil error for status %d", statusCode)
			}
			wantErrPart := fmt.Sprintf("status %d", statusCode)
			if !strings.Contains(err.Error(), wantErrPart) {
				t.Fatalf("ResolveURI() error = %q, want it to contain %q", err, wantErrPart)
			}
			for _, secret := range []string{"query-secret", "fragment-secret", "response-body-must-not-appear"} {
				if strings.Contains(err.Error(), secret) {
					t.Fatalf("ResolveURI() error %q disclosed %q", err, secret)
				}
			}
		})
	}
}

func TestSchemaResolverResolveURITimesOut(t *testing.T) {
	requestCanceled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		<-request.Context().Done()
		close(requestCanceled)
	}))
	defer server.Close()

	resolver := NewSchemaResolver(Standard, false)
	resolver.httpClient.Timeout = 200 * time.Millisecond

	_, err := resolver.ResolveURI(server.URL + "/slow?token=timeout-secret#timeout-fragment")
	if err == nil {
		t.Fatal("ResolveURI() returned nil error for a timed-out response")
	}
	const wantErrPart = "timed out"
	if !strings.Contains(err.Error(), wantErrPart) {
		t.Fatalf("ResolveURI() error = %q, want it to contain %q", err, wantErrPart)
	}
	if strings.Contains(err.Error(), "timeout-secret") || strings.Contains(err.Error(), "timeout-fragment") {
		t.Fatalf("ResolveURI() timeout error disclosed URL secrets: %q", err)
	}

	select {
	case <-requestCanceled:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not observe request cancellation")
	}
}

func TestSchemaResolverResolveURIRejectsOversizedResponses(t *testing.T) {
	for _, test := range []struct {
		name   string
		stream bool
	}{
		{name: "content length"},
		{name: "streamed body", stream: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
				if test.stream {
					if flusher, ok := response.(http.Flusher); ok {
						response.WriteHeader(http.StatusOK)
						flusher.Flush()
					}
					_, _ = response.Write([]byte(strings.Repeat("x", maxSchemaResponseBytes+1)))
					return
				}

				response.Header().Set("Content-Length", strconv.Itoa(maxSchemaResponseBytes+1))
				response.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			resolver := NewSchemaResolver(Standard, false)
			_, err := resolver.ResolveURI(server.URL)
			if err == nil {
				t.Fatal("ResolveURI() returned nil error for an oversized response")
			}
			wantErrPart := fmt.Sprintf("exceeds %d bytes", maxSchemaResponseBytes)
			if !strings.Contains(err.Error(), wantErrPart) {
				t.Fatalf("ResolveURI() error = %q, want it to contain %q", err, wantErrPart)
			}
		})
	}
}

func TestSchemaResolverResolveURIAcceptsResponsesAtSizeLimit(t *testing.T) {
	schema := testRecordSchema + strings.Repeat(" ", maxSchemaResponseBytes-len(testRecordSchema))

	for _, test := range []struct {
		name   string
		stream bool
	}{
		{name: "content length"},
		{name: "streamed body", stream: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
				if test.stream {
					if flusher, ok := response.(http.Flusher); ok {
						response.WriteHeader(http.StatusOK)
						flusher.Flush()
					}
				} else {
					response.Header().Set("Content-Length", strconv.Itoa(maxSchemaResponseBytes))
				}
				_, _ = response.Write([]byte(schema))
			}))
			defer server.Close()

			resolver := NewSchemaResolver(Standard, false)
			codec, err := resolver.ResolveURI(server.URL)
			if err != nil {
				t.Fatalf("ResolveURI(): unexpected error: %v", err)
			}
			if len(codec.Schema()) != maxSchemaResponseBytes {
				t.Fatalf("resolved schema length = %d, want %d", len(codec.Schema()), maxSchemaResponseBytes)
			}
		})
	}
}

func TestSchemaResolverResolveURIFollowsSafeRedirect(t *testing.T) {
	var targetRequests atomic.Int32
	targetReferer := make(chan string, 1)
	target := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		targetRequests.Add(1)
		targetReferer <- request.Header.Get("Referer")
		_, _ = response.Write([]byte(testRecordSchema))
	}))
	defer target.Close()

	source := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, target.URL+"/schema", http.StatusFound)
	}))
	defer source.Close()

	resolver := NewSchemaResolver(Standard, false)
	codec, err := resolver.ResolveURI(source.URL + "?token=secret#source-fragment")
	if err != nil {
		t.Fatalf("ResolveURI(): unexpected error: %v", err)
	}
	if codec.Schema() != testRecordSchema {
		t.Fatalf("resolved schema = %q, want %q", codec.Schema(), testRecordSchema)
	}
	if targetRequests.Load() != 1 {
		t.Fatalf("redirect target request count = %d, want 1", targetRequests.Load())
	}
	if referer := <-targetReferer; referer != "" {
		t.Fatalf("redirect target Referer = %q, want empty", referer)
	}
}

func TestCheckSchemaRedirectRemovesSensitiveHeaders(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "https://example.invalid/schema.avsc", nil)
	for _, header := range []string{
		"Authorization",
		"Proxy-Authorization",
		"Cookie",
		"Referer",
	} {
		request.Header.Set(header, "secret")
	}
	request.Header.Set("X-Test-Header", "keep")

	if err := checkSchemaRedirect(request, nil); err != nil {
		t.Fatalf("checkSchemaRedirect(): unexpected error: %v", err)
	}

	for _, header := range []string{
		"Authorization",
		"Proxy-Authorization",
		"Cookie",
		"Referer",
	} {
		if value := request.Header.Get(header); value != "" {
			t.Errorf("redirect request header %q = %q, want empty", header, value)
		}
	}
	if value := request.Header.Get("X-Test-Header"); value != "keep" {
		t.Errorf("redirect request header %q = %q, want %q", "X-Test-Header", value, "keep")
	}
}

func TestSchemaResolverResolveURIRejectsUnsafeRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Location", "ftp://example.invalid/schema.avsc")
		response.WriteHeader(http.StatusFound)
	}))
	defer server.Close()

	resolver := NewSchemaResolver(Standard, false)
	_, err := resolver.ResolveURI(server.URL + "?token=redirect-secret")
	if err == nil {
		t.Fatal("ResolveURI() returned nil error for a non-HTTP redirect")
	}
	const wantErrPart = "redirect target must use http or https"
	if !strings.Contains(err.Error(), wantErrPart) {
		t.Fatalf("ResolveURI() error = %q, want it to contain %q", err, wantErrPart)
	}
	if strings.Contains(err.Error(), "redirect-secret") {
		t.Fatalf("ResolveURI() error disclosed the source query: %q", err)
	}
}

func TestSchemaResolverResolveURIRejectsRedirectWithoutHostname(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Location", "http://:8080/schema.avsc")
		response.WriteHeader(http.StatusFound)
	}))
	defer server.Close()

	resolver := NewSchemaResolver(Standard, false)
	_, err := resolver.ResolveURI(server.URL)
	if err == nil {
		t.Fatal("ResolveURI() returned nil error for a redirect without a hostname")
	}
	const wantErrPart = "redirect target must include a host"
	if !strings.Contains(err.Error(), wantErrPart) {
		t.Fatalf("ResolveURI() error = %q, want it to contain %q", err, wantErrPart)
	}
}

func TestSchemaResolverResolveURIDoesNotDiscloseMalformedRedirectLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set(
			"Location",
			"https://user:password@example.invalid/%zz?token=redirect-secret#redirect-fragment",
		)
		response.WriteHeader(http.StatusFound)
	}))
	defer server.Close()

	resolver := NewSchemaResolver(Standard, false)
	_, err := resolver.ResolveURI(server.URL)
	if err == nil {
		t.Fatal("ResolveURI() returned nil error for a malformed redirect location")
	}
	const wantErrPart = "request failed"
	if !strings.Contains(err.Error(), wantErrPart) {
		t.Fatalf("ResolveURI() error = %q, want it to contain %q", err, wantErrPart)
	}
	for _, secret := range []string{"user", "password", "token=redirect-secret", "redirect-fragment", "%zz"} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("ResolveURI() error %q disclosed %q from the redirect location", err, secret)
		}
	}
}

func TestSchemaResolverResolveURIStopsAfterTenRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var redirect int
		_, _ = fmt.Sscanf(strings.TrimPrefix(request.URL.Path, "/"), "%d", &redirect)
		http.Redirect(response, request, fmt.Sprintf("/%d", redirect+1), http.StatusFound)
	}))
	defer server.Close()

	resolver := NewSchemaResolver(Standard, false)
	_, err := resolver.ResolveURI(server.URL + "/0")
	if err == nil {
		t.Fatal("ResolveURI() returned nil error for an endless redirect chain")
	}
	const wantErrPart = "stopped after 10 redirects"
	if !strings.Contains(err.Error(), wantErrPart) {
		t.Fatalf("ResolveURI() error = %q, want it to contain %q", err, wantErrPart)
	}
}

func TestSchemaResolverResolveURIRejectsUnsupportedAndCredentialedURIs(t *testing.T) {
	for _, test := range []struct {
		name        string
		uri         string
		wantErrPart string
		secrets     []string
	}{
		{name: "empty", uri: " \r\n", wantErrPart: "must not be empty"},
		{name: "relative", uri: "schemas/message.avsc", wantErrPart: "only http and https"},
		{name: "file", uri: "file:///tmp/message.avsc", wantErrPart: "only http and https"},
		{name: "ftp", uri: "ftp://example.invalid/message.avsc", wantErrPart: "only http and https"},
		{name: "missing host", uri: "https:///message.avsc?token=secret#fragment", wantErrPart: "must include a host", secrets: []string{"token=secret", "fragment"}},
		{name: "missing hostname", uri: "https://:443/message.avsc?token=secret#fragment", wantErrPart: "must include a host", secrets: []string{"token=secret", "fragment"}},
		{name: "userinfo", uri: "https://user:password@example.invalid/message.avsc?token=secret#fragment", wantErrPart: "must not include user information", secrets: []string{"user:password", "password", "token=secret", "fragment"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			resolver := NewSchemaResolver(Standard, false)
			_, err := resolver.ResolveURI(test.uri)
			if err == nil {
				t.Fatal("ResolveURI() returned nil error")
			}
			if !strings.Contains(err.Error(), test.wantErrPart) {
				t.Fatalf("ResolveURI() error = %q, want it to contain %q", err, test.wantErrPart)
			}
			for _, secret := range test.secrets {
				if strings.Contains(err.Error(), secret) {
					t.Fatalf("ResolveURI() error %q disclosed %q", err, secret)
				}
			}
		})
	}
}

func TestSchemaResolverResolveURICachesCompiledCodec(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = response.Write([]byte(testRecordSchema))
	}))
	defer server.Close()

	resolver := NewSchemaResolver(Standard, true)
	const callers = 8
	codecs := make([]*MessageCodec, callers)
	errorsByCaller := make([]error, callers)
	var waitGroup sync.WaitGroup

	for caller := range callers {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			codecs[caller], errorsByCaller[caller] = resolver.ResolveURI(server.URL)
		}()
	}
	waitGroup.Wait()

	for caller, err := range errorsByCaller {
		if err != nil {
			t.Fatalf("ResolveURI() caller %d: unexpected error: %v", caller, err)
		}
		if codecs[caller] != codecs[0] {
			t.Fatalf("ResolveURI() caller %d returned codec %p, want cached codec %p", caller, codecs[caller], codecs[0])
		}
	}
	if requests.Load() != 1 {
		t.Fatalf("HTTP request count = %d, want 1", requests.Load())
	}
}

func TestSchemaResolverResolveURICacheIgnoresFragmentAndPreservesQuery(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		switch request.URL.Query().Get("version") {
		case "1":
			_, _ = response.Write([]byte(testRecordSchema))
		case "2":
			_, _ = response.Write([]byte(`"string"`))
		default:
			t.Errorf("unexpected schema version query %q", request.URL.RawQuery)
		}
	}))
	defer server.Close()

	resolver := NewSchemaResolver(Standard, true)
	first, err := resolver.ResolveURI(server.URL + "/schema?version=1#revision%2Fone")
	if err != nil {
		t.Fatalf("first ResolveURI(): unexpected error: %v", err)
	}
	second, err := resolver.ResolveURI(server.URL + "/schema?version=1#revision-two")
	if err != nil {
		t.Fatalf("second ResolveURI(): unexpected error: %v", err)
	}
	third, err := resolver.ResolveURI(server.URL + "/schema?version=2#revision-one")
	if err != nil {
		t.Fatalf("third ResolveURI(): unexpected error: %v", err)
	}

	if first != second {
		t.Fatalf("fragment-only URI change returned codec %p, want cached codec %p", second, first)
	}
	if third == first {
		t.Fatalf("different query returned cached codec %p, want a distinct codec", third)
	}
	if third.Schema() != `"string"` {
		t.Fatalf("resolved schema = %q, want %q", third.Schema(), `"string"`)
	}
	if requests.Load() != 2 {
		t.Fatalf("HTTP request count = %d, want 2", requests.Load())
	}
}

func TestSchemaResolverResolveURIFetchesEveryTimeWhenCacheIsDisabled(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = response.Write([]byte(testRecordSchema))
	}))
	defer server.Close()

	resolver := NewSchemaResolver(Standard, false)
	first, err := resolver.ResolveURI(server.URL)
	if err != nil {
		t.Fatalf("first ResolveURI(): unexpected error: %v", err)
	}
	second, err := resolver.ResolveURI(server.URL)
	if err != nil {
		t.Fatalf("second ResolveURI(): unexpected error: %v", err)
	}

	if first == second {
		t.Fatal("ResolveURI() returned the same compiled codec while caching was disabled")
	}
	if requests.Load() != 2 {
		t.Fatalf("HTTP request count = %d, want 2", requests.Load())
	}
}

func TestSchemaResolverResolveURIDoesNotCacheFailures(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) == 1 {
			response.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = response.Write([]byte(testRecordSchema))
	}))
	defer server.Close()

	resolver := NewSchemaResolver(Standard, true)
	if _, err := resolver.ResolveURI(server.URL); err == nil {
		t.Fatal("first ResolveURI() returned nil error for a failed response")
	}
	if _, err := resolver.ResolveURI(server.URL); err != nil {
		t.Fatalf("second ResolveURI(): unexpected error: %v", err)
	}
	if requests.Load() != 2 {
		t.Fatalf("HTTP request count = %d, want 2", requests.Load())
	}
}

func TestSchemaResolverResolveURIRejectsEmptyAndInvalidSchemas(t *testing.T) {
	for _, test := range []struct {
		name    string
		content string
		secret  string
	}{
		{name: "empty"},
		{name: "whitespace", content: " \r\n\t"},
		{name: "invalid", content: `"remote-schema-secret"`, secret: "remote-schema-secret"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
				_, _ = response.Write([]byte(test.content))
			}))
			defer server.Close()

			resolver := NewSchemaResolver(Standard, false)
			_, err := resolver.ResolveURI(server.URL)
			if err == nil {
				t.Fatal("ResolveURI() returned nil error for an unusable schema")
			}
			const wantErrPart = "failed to load avro schema"
			if !strings.Contains(err.Error(), wantErrPart) {
				t.Fatalf("ResolveURI() error = %q, want it to contain %q", err, wantErrPart)
			}
			if test.secret != "" && strings.Contains(err.Error(), test.secret) {
				t.Fatalf("ResolveURI() error %q disclosed the schema response", err)
			}
		})
	}
}
