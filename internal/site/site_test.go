package site

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSubdomainElementRoutes(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		path       string
		wantStatus int
		wantBody   []string
	}{
		{
			name:       "local docs",
			host:       "docs.rocketweb.nl:8080",
			path:       "/go-partial",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"go-partial documentation",
				"Server-rendered partials that stay useful with HTMX",
				"href=\"/go-partial/installation\"",
				"aria-current=\"page\"",
				"href=\"/go-docs\"",
				"href=\"/go-router\"",
			},
		},
		{
			name:       "local showcase",
			host:       "showcase.rocketweb.nl:8080",
			path:       "/go-router",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Go Router Showcase",
				"https://showcase.go-webthings.com/go-router",
				"http://showcase.rocketweb.nl:8080/go-router",
			},
		},
		{
			name:       "production docs",
			host:       "docs.go-webthings.com",
			path:       "/go-docs",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Typed contracts for Go templates",
				"href=\"/go-docs/install\"",
				"href=\"/assets/site.css\"",
				"href=\"/go-partial\"",
				"href=\"/go-docs\" class=\"active\" aria-current=\"page\"",
				"href=\"/go-router\"",
			},
		},
		{
			name:       "go docs nested docs page",
			host:       "docs.go-webthings.com",
			path:       "/go-docs/contracts",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Template contracts",
				"Contract first, runtime second.",
				"href=\"/go-docs/install\"",
			},
		},
		{
			name:       "go router docs",
			host:       "docs.go-webthings.com",
			path:       "/go-router",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"HTTP routing for Go websites",
				"href=\"/go-router/routing\"",
				"href=\"/go-partial\"",
				"href=\"/go-docs\"",
				"href=\"/go-router\" class=\"active\" aria-current=\"page\"",
			},
		},
		{
			name:       "go router host docs",
			host:       "docs.go-webthings.com",
			path:       "/go-router/hosts",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Hosts and subdomains",
				`r.Subdomain("docs", "go-webthings.com"`,
				"href=\"/go-router/routing\"",
			},
		},
		{
			name:       "go router middleware overview",
			host:       "docs.go-webthings.com",
			path:       "/go-router/middleware",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Middleware",
				"href=\"/go-router/middleware/request-id\"",
				"href=\"/go-router/middleware/cors\"",
			},
		},
		{
			name:       "go router cors middleware",
			host:       "docs.go-webthings.com",
			path:       "/go-router/middleware/cors",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"CORS middleware",
				"AllowedOrigins",
				"204 No Content",
			},
		},
		{
			name:       "go router diagnostics",
			host:       "docs.go-webthings.com",
			path:       "/go-router/diagnostics",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Route diagnostics",
				"r.Walk",
				"r.RouteTable",
			},
		},
		{
			name:       "go partial nested docs page",
			host:       "docs.go-webthings.com",
			path:       "/go-partial/rendering",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Rendering model",
				"href=\"/go-partial/data-context\"",
				"hx-get=\"/go-partial/data-context\"",
			},
		},
		{
			name:       "old nested docs path is not canonical",
			host:       "docs.go-webthings.com",
			path:       "/go-partial/docs/rendering",
			wantStatus: http.StatusNotFound,
			wantBody: []string{
				"Element not found",
			},
		},
		{
			name:       "production showcase",
			host:       "showcase.go-webthings.com",
			path:       "/go-docs",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Go Docs Showcase",
				"https://showcase.go-webthings.com/go-docs",
			},
		},
		{
			name:       "unknown element",
			host:       "docs.go-webthings.com",
			path:       "/go-webthings",
			wantStatus: http.StatusNotFound,
			wantBody: []string{
				"Element not found",
			},
		},
	}

	handler := NewRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Host = tt.host
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			body := rec.Body.String()
			for _, want := range tt.wantBody {
				if !strings.Contains(body, want) {
					t.Fatalf("expected body to contain %q\nbody:\n%s", want, body)
				}
			}
		})
	}
}

func TestSharedStylesheet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/assets/site.css", nil)
	req.Host = "docs.go-webthings.com"
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "tailwindcss") || !strings.Contains(body, "slate-950") {
		t.Fatalf("expected shared docs stylesheet, got:\n%s", body)
	}
}

func TestUnknownHostFallsBackToApexRedirect(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "rocketweb.nl:8080"
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "https://docs.go-webthings.com/go-router" {
		t.Fatalf("expected redirect to production docs, got %q", got)
	}
}
