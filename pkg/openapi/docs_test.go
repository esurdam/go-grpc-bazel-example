package openapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/esurdam/go-grpc-bazel-example/pkg/openapi"
)

func TestDocsHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		specURL string
		title   string
		wantIn  []string
		wantNot []string
	}{
		{
			name:    "defaults",
			specURL: "",
			title:   "",
			wantIn: []string{
				"API documentation",
				`url: "/swagger.json"`,
				"@scalar/api-reference@",
				"Scalar.createApiReference",
			},
		},
		{
			name:    "custom title and spec",
			specURL: "/v1/openapi.json",
			title:   "Greeter API",
			wantIn: []string{
				"Greeter API",
				`url: "/v1/openapi.json"`,
				"Scalar.createApiReference",
			},
		},
		{
			name:    "escapes hostile title and spec URL",
			specURL: `"</script><script>alert(1)</script>`,
			title:   `<script>alert("xss")</script>`,
			wantIn: []string{
				`&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;`,
				`url: "\"\u003c/script\u003e\u003cscript\u003ealert(1)\u003c/script\u003e"`,
			},
			wantNot: []string{
				`<script>alert("xss")</script>`,
				`url: "</script><script>alert(1)</script>"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/docs/", nil)
			rr := httptest.NewRecorder()
			openapi.DocsHandler(tt.specURL, tt.title).ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
			}
			ct := rr.Header().Get("Content-Type")
			if !strings.HasPrefix(ct, "text/html") {
				t.Fatalf("Content-Type = %q, want text/html", ct)
			}
			body := rr.Body.String()
			for _, want := range tt.wantIn {
				if !strings.Contains(body, want) {
					t.Errorf("body missing %q\nbody:\n%s", want, body)
				}
			}
			for _, notWant := range tt.wantNot {
				if strings.Contains(body, notWant) {
					t.Errorf("body unexpectedly contains %q", notWant)
				}
			}
		})
	}
}

func TestSpecHandler(t *testing.T) {
	t.Parallel()

	spec := []byte(`{"openapi":"3.0.0","info":{"title":"t","version":"1"}}`)
	req := httptest.NewRequest(http.MethodGet, "/swagger.json", nil)
	rr := httptest.NewRecorder()
	openapi.SpecHandler(spec).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	if got := rr.Body.String(); got != string(spec) {
		t.Fatalf("body = %q, want %q", got, spec)
	}
}

func TestMountRoutes(t *testing.T) {
	t.Parallel()

	spec := []byte(`{"swagger":"2.0","info":{"title":"Helloworld API","version":"1.0"}}`)
	mux := http.NewServeMux()
	openapi.Mount(mux, spec, "Helloworld API")

	t.Run("docs redirects to trailing slash", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/docs", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusFound {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusFound)
		}
		if got := rr.Header().Get("Location"); got != "/docs/" {
			t.Fatalf("Location = %q, want /docs/", got)
		}
	})

	t.Run("docs slash serves UI", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/docs/", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "Helloworld API") {
			t.Fatalf("docs body missing title")
		}
		if !strings.Contains(body, `url: "/swagger.json"`) {
			t.Fatalf("docs body missing spec url")
		}
	})

	t.Run("swagger json", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/swagger.json", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if got := rr.Body.String(); got != string(spec) {
			t.Fatalf("body = %q, want %q", got, spec)
		}
	})
}
