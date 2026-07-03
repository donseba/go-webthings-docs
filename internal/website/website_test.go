package website

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
				"Server-rendered partials that stay useful with HTMX",
				`<meta name="description" content="go-partial is a small rendering layer for Go applications that want reusable template regions, targeted updates, out-of-band swaps, and predictable server-side behavior.">`,
				`<meta property="og:image" content="https://docs.gowebthings.com/assets/img/go-partial-400.png">`,
				"href=\"/go-partial/installation\"",
				"src=\"/assets/img/go-partial-40.png\"",
				"aria-current=\"page\"",
				"href=\"/go-docs\"",
				"href=\"/go-router\"",
				"href=\"http://rocketweb.nl:8080\"",
			},
		},
		{
			name:       "docs root index",
			host:       "docs.gowebthings.com",
			path:       "/",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Docs for go-webthings",
				"src=\"/assets/img/go-webthings-400.png\"",
				"src=\"/assets/img/go-partial-300.png\"",
				"src=\"/assets/img/go-doc-300.png\"",
				"src=\"/assets/img/go-router-300.png\"",
				"class=\"root-card\"",
			},
		},
		{
			name:       "production main apex",
			host:       "gowebthings.com",
			path:       "/",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"go-webthings",
				`<body class="main-body">`,
				"Composable Go packages",
				`<link rel="canonical" href="https://gowebthings.com">`,
				`<meta property="og:image" content="https://gowebthings.com/assets/img/go-webthings-400.png">`,
				`<link rel="icon" href="/assets/img/favicon.ico" sizes="any">`,
				"href=\"https://docs.gowebthings.com/go-partial\"",
				"href=\"https://docs.gowebthings.com/go-docs\"",
				"href=\"https://docs.gowebthings.com/go-router\"",
			},
		},
		{
			name:       "production main www",
			host:       "www.gowebthings.com",
			path:       "/",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"go-webthings",
				"href=\"https://docs.gowebthings.com/go-partial\"",
				"src=\"/assets/img/go-webthings-400.png\"",
			},
		},
		{
			name:       "local main www",
			host:       "www.rocketweb.nl:8080",
			path:       "/",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"go-webthings",
				"href=\"http://docs.rocketweb.nl:8080/go-partial\"",
				"href=\"/assets/css/styles.css\"",
			},
		},
		{
			name:       "production main element",
			host:       "gowebthings.com",
			path:       "/go-router",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Go Router",
				"Host-aware HTTP routing built on top of net/http.",
				"https://gowebthings.com/go-router",
				"Open go-router",
			},
		},
		{
			name:       "local showcase",
			host:       "showcase.rocketweb.nl:8080",
			path:       "/go-router",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Go Router showcase coming soon",
				"Coming soon",
				"https://docs.gowebthings.com/go-partial",
			},
		},
		{
			name:       "production docs",
			host:       "docs.gowebthings.com",
			path:       "/go-docs",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Typed contracts for Go templates",
				`<title>Typed contracts for Go templates - go-doc docs</title>`,
				`<link rel="icon" href="/assets/img/favicon.ico" sizes="any">`,
				"href=\"/go-docs/install\"",
				"href=\"/assets/css/styles.css\"",
				"src=\"/assets/js/code-highlight.js\"",
				"src=\"/assets/img/go-doc-40.png\"",
				"href=\"/go-partial\"",
				"href=\"/go-docs\" hx-get=\"/go-docs\"",
				"class=\"active\" aria-current=\"page\"",
				"hx-get=\"/go-docs/install\"",
				"href=\"/go-router\"",
				"href=\"https://gowebthings.com\"",
			},
		},
		{
			name:       "go docs nested docs page",
			host:       "docs.gowebthings.com",
			path:       "/go-docs/contracts",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Template contracts",
				"Contract first, runtime second.",
				"href=\"/go-docs/install\"",
				"hx-get=\"/go-docs/install\"",
			},
		},
		{
			name:       "go router docs",
			host:       "docs.gowebthings.com",
			path:       "/go-router",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"HTTP routing for Go websites",
				`<meta property="og:image" content="https://docs.gowebthings.com/assets/img/go-router-400.png">`,
				"src=\"/assets/img/go-router-40.png\"",
				"href=\"/go-router/routing\"",
				"href=\"/go-partial\"",
				"href=\"/go-docs\"",
				"href=\"/go-router\" hx-get=\"/go-router\"",
				"class=\"active\" aria-current=\"page\"",
				"hx-get=\"/go-router/routing\"",
			},
		},
		{
			name:       "go router host docs",
			host:       "docs.gowebthings.com",
			path:       "/go-router/hosts",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Hosts and subdomains",
				`r.Subdomain("docs", "gowebthings.com"`,
				"href=\"/go-router/routing\"",
			},
		},
		{
			name:       "go router middleware overview",
			host:       "docs.gowebthings.com",
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
			host:       "docs.gowebthings.com",
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
			host:       "docs.gowebthings.com",
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
			host:       "docs.gowebthings.com",
			path:       "/go-partial/rendering",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Rendering model",
				"href=\"/go-partial/data-context\"",
				"hx-get=\"/go-partial/data-context\"",
			},
		},
		{
			name:       "go partial interactions docs page",
			host:       "docs.gowebthings.com",
			path:       "/go-partial/interactions",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Interaction helpers",
				"poll-notifications",
				"hx-trigger=\"every 10s\"",
			},
		},
		{
			name:       "old nested docs path is not canonical",
			host:       "docs.gowebthings.com",
			path:       "/go-partial/docs/rendering",
			wantStatus: http.StatusNotFound,
			wantBody: []string{
				"Element not found",
			},
		},
		{
			name:       "production showcase",
			host:       "showcase.gowebthings.com",
			path:       "/go-docs",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Go Docs showcase coming soon",
				"Coming soon",
				"https://showcase.gowebthings.com/go-docs",
				"https://docs.gowebthings.com/go-partial",
			},
		},
		{
			name:       "unknown element",
			host:       "docs.gowebthings.com",
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
	req := httptest.NewRequest(http.MethodGet, "/assets/css/styles.css", nil)
	req.Host = "docs.gowebthings.com"
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "tailwindcss") || !strings.Contains(body, "slate-950") {
		t.Fatalf("expected shared docs stylesheet, got:\n%s", body)
	}
	if body := rec.Body.String(); !strings.Contains(body, "syntax-keyword") || !strings.Contains(body, "display:table") {
		t.Fatalf("expected docs code and table styles, got:\n%s", body)
	}
	if body := rec.Body.String(); !strings.Contains(body, ".root-hub") || !strings.Contains(body, ".root-card") {
		t.Fatalf("expected root hub styles, got:\n%s", body)
	}
}

func TestMainStylesheet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/assets/css/styles.css", nil)
	req.Host = "gowebthings.com"
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, ".main-body") || !strings.Contains(body, ".main-card") {
		t.Fatalf("expected main website stylesheet, got:\n%s", body)
	}
	if strings.Contains(body, ".docs-body") {
		t.Fatalf("main stylesheet should not be the docs stylesheet, got:\n%s", body)
	}
}

func TestDocsCodeHighlightAsset(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/assets/js/code-highlight.js", nil)
	req.Host = "docs.gowebthings.com"
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "highlightAll") || !strings.Contains(body, "htmx:afterSwap") || !strings.Contains(body, "highlightGoDocComment") {
		t.Fatalf("expected docs highlighter asset, got:\n%s", body)
	}
}

func TestFaviconRoutes(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	req.Host = "gowebthings.com"
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "image/") && !strings.Contains(contentType, "application/octet-stream") {
		t.Fatalf("expected favicon content type, got %q", contentType)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("expected favicon body")
	}
}

func TestDocsElementTemplatesUseSharedArticleShape(t *testing.T) {
	err := fs.WalkDir(websiteFS, "templates", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".gohtml") {
			return nil
		}
		if !strings.HasPrefix(path, "templates/go_partial/") && !strings.HasPrefix(path, "templates/go_doc/") && !strings.HasPrefix(path, "templates/go_router/") {
			return nil
		}

		bodyBytes, err := fs.ReadFile(websiteFS, path)
		if err != nil {
			return err
		}
		body := strings.TrimSpace(string(bodyBytes))
		if !strings.HasPrefix(body, "{{/* @dot donseba/go-webthings-docs/internal/website.DocsArticleData */}}") {
			t.Fatalf("%s should start with the DocsArticleData @dot contract", path)
		}
		if count := strings.Count(body, "<article"); count != 1 {
			t.Fatalf("%s should contain one article wrapper, got %d", path, count)
		}
		if count := strings.Count(body, `{{ template "hero.gohtml" . }}`); count != 1 {
			t.Fatalf("%s should render the shared hero once, got %d", path, count)
		}
		if strings.Contains(body, "text-5xl") || strings.Contains(body, "text-emerald-300") {
			t.Fatalf("%s should not duplicate shared hero styling", path)
		}
		if !strings.HasSuffix(body, "</article>") {
			t.Fatalf("%s should close with </article>", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDocsTemplatesUseDeployLayout(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	for _, dir := range []string{
		"deploy/website/docs",
		"deploy/website/main",
		"deploy/website/showcase",
	} {
		if info, err := os.Stat(filepath.Join(repoRoot, dir)); err != nil || !info.IsDir() {
			t.Fatalf("expected deploy section %s to exist: %v", dir, err)
		}
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "deploy/docs")); err == nil {
		t.Fatal("old deploy/docs tree should not exist")
	}

	for _, dir := range []string{
		"templates/general",
		"templates/go_partial",
		"templates/go_doc",
		"templates/go_router",
	} {
		if _, err := fs.Stat(websiteFS, dir); err != nil {
			t.Fatalf("expected %s to exist: %v", dir, err)
		}
	}

	if _, err := fs.Stat(websiteFS, "elements"); err == nil {
		t.Fatal("old element template tree should not exist")
	}
	if _, err := fs.Stat(mainFS, "templates/page.gohtml"); err != nil {
		t.Fatalf("expected main website template to exist: %v", err)
	}
	if _, err := fs.Stat(mainFS, "assets/css/styles.css"); err != nil {
		t.Fatalf("expected main website stylesheet to exist: %v", err)
	}
}

func TestWebsiteFileSystemFromDeployRoot(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	deployRoot := filepath.Join(repoRoot, "deploy", "website")
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
	if err := os.Chdir(deployRoot); err != nil {
		t.Fatal(err)
	}

	fsys := docsFileSystem()
	if _, err := fs.Stat(fsys, "templates/general/layout.gohtml"); err != nil {
		t.Fatalf("expected deploy root to resolve docs templates: %v", err)
	}
}

func TestDocsPageTemplatesUseElementDirectories(t *testing.T) {
	for name, test := range map[string]struct {
		pages map[string]docsPage
		dir   string
	}{
		"go-partial": {pages: goPartialDocs.pages, dir: "templates/go_partial/"},
		"go-docs":    {pages: goDocsDocs.pages, dir: "templates/go_doc/"},
		"go-router":  {pages: goRouterDocs.pages, dir: "templates/go_router/"},
	} {
		for route, page := range test.pages {
			if !strings.HasPrefix(page.Template, test.dir) {
				t.Fatalf("%s route %s should use %s, got %q", name, route, test.dir, page.Template)
			}
			if strings.Contains(strings.TrimPrefix(page.Template, test.dir), "/") || strings.Contains(page.Template, `\`) {
				t.Fatalf("%s route %s should keep templates flat inside the element directory, got %q", name, route, page.Template)
			}
			if !strings.HasSuffix(page.Template, ".gohtml") {
				t.Fatalf("%s route %s should point to a gohtml template, got %q", name, route, page.Template)
			}
		}
	}
}

func TestDocsInternalLinksUseHTMX(t *testing.T) {
	err := fs.WalkDir(websiteFS, "templates", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".gohtml") {
			return nil
		}
		if !strings.HasPrefix(path, "templates/go_partial/") && !strings.HasPrefix(path, "templates/go_doc/") && !strings.HasPrefix(path, "templates/go_router/") {
			return nil
		}

		bodyBytes, err := fs.ReadFile(websiteFS, path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(bodyBytes), "\n") {
			if !strings.Contains(line, "<a ") || !strings.Contains(line, "href=") {
				continue
			}
			internal := strings.Contains(line, `href="{{ basePath }}`) ||
				strings.Contains(line, `href="{{ goDocsPath`) ||
				strings.Contains(line, `href="{{ goRouterPath`)
			if internal && (!strings.Contains(line, "hx-get=") || !strings.Contains(line, `hx-target="#content"`) || !strings.Contains(line, `hx-push-url="true"`)) {
				t.Fatalf("%s:%d internal docs link should use HTMX navigation: %s", path, i+1, strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDocsPromptFaces(t *testing.T) {
	if len(docsPromptFaces) < 20 {
		t.Fatalf("expected at least 20 prompt faces, got %d", len(docsPromptFaces))
	}
	for _, want := range []string{"#_>", "-_-", "^-^", "#_#"} {
		found := false
		for _, face := range docsPromptFaces {
			if face == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected prompt face list to contain %q", want)
		}
	}
}

func TestHTMXDocsRequestsReturnFragments(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "go partial",
			path: "/go-partial/rendering",
			want: []string{
				`hx-swap-oob="true"`,
				`id="app-header"`,
				`id="docs-sidebar"`,
				"Rendering model",
				"Wrapper plus content",
			},
		},
		{
			name: "go docs",
			path: "/go-docs/annotations",
			want: []string{
				`hx-swap-oob="true"`,
				`id="app-header"`,
				`id="docs-sidebar"`,
				"Annotations",
				"Model, dot, function, and symbol annotations",
			},
		},
		{
			name: "go router",
			path: "/go-router/middleware/timeout",
			want: []string{
				`hx-swap-oob="true"`,
				`id="app-header"`,
				`id="docs-sidebar"`,
				"Timeout middleware",
				"Wrap handlers with http.TimeoutHandler",
			},
		},
	}

	handler := NewRouter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Host = "docs.gowebthings.com"
			req.Header.Set("HX-Request", "true")
			req.Header.Set("HX-Target", "content")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
			}
			body := rec.Body.String()
			if strings.Contains(body, "<!doctype html>") || strings.Contains(body, "<body") {
				t.Fatalf("expected HTMX fragment without full shell, got:\n%s", body)
			}
			for _, want := range tt.want {
				if !strings.Contains(body, want) {
					t.Fatalf("expected body to contain %q\nbody:\n%s", want, body)
				}
			}
		})
	}
}

func TestUnknownHostFallsBackToApexRedirect(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "unknown.rocketweb.nl:8080"
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "https://docs.gowebthings.com/go-router" {
		t.Fatalf("expected redirect to production docs, got %q", got)
	}
}
