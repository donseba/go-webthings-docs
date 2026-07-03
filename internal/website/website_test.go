package website

import (
	"bufio"
	"context"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
				`<meta property="og:image" content="https://docs.gowebthings.com/assets/img/logo-go-partial.png">`,
				"href=\"/go-partial/installation\"",
				"src=\"/assets/img/logo-go-partial.png\"",
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
				"src=\"/assets/img/logo-go-webthings.png\"",
				"src=\"/assets/img/logo-go-partial.png\"",
				"src=\"/assets/img/logo-go-doc.png\"",
				"src=\"/assets/img/logo-go-router.png\"",
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
				`<meta property="og:image" content="https://gowebthings.com/assets/img/logo-go-webthings.png">`,
				`<link rel="icon" href="/assets/img/favicon.ico" sizes="any">`,
				`href="/" hx-get="/" hx-target="#content" hx-push-url="true"`,
				`href="/components" hx-get="/components" hx-target="#content" hx-push-url="true"`,
				`href="/generate" hx-get="/generate" hx-target="#content" hx-push-url="true"`,
				"href=\"https://docs.gowebthings.com/go-partial\"",
				"href=\"https://showcase.gowebthings.com\"",
				"Small Go packages for HTML-first websites.",
				"HTMX-friendly updates without turning the browser into the main application runtime",
				"Three packages, one HTML response.",
				"Requests are routed by go-router",
				"current core",
				"go-translator",
				"go-form",
				"go-importmap",
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
				"href=\"https://showcase.gowebthings.com\"",
				"src=\"/assets/img/logo-go-webthings.png\"",
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
				"href=\"http://showcase.rocketweb.nl:8080\"",
				"href=\"/assets/css/styles.css\"",
			},
		},
		{
			name:       "production main components",
			host:       "gowebthings.com",
			path:       "/components",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"go-webthings components",
				`href="/components" hx-get="/components" hx-target="#content" hx-push-url="true" aria-current="page"`,
				"Typed contracts and editor metadata",
				"Server-side partial rendering",
				"Host-aware HTTP routing",
				"href=\"https://docs.gowebthings.com/go-docs\"",
				"href=\"https://showcase.gowebthings.com\"",
				"href=\"https://github.com/donseba/go-router\"",
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
			name:       "production generator",
			host:       "gowebthings.com",
			path:       "/generate",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"Go logo generator",
				`hx-get="/generate/preview"`,
				`hx-target="#generator-preview"`,
				`src="/generate/image?text=WebThings"`,
			},
		},
		{
			name:       "local showcase",
			host:       "showcase.rocketweb.nl:8080",
			path:       "/",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"go-webthings showcase",
				"Server-rendered partials",
				"This application renders normal pages and htmx requests through the same partial tree.",
				"href=\"/rows\"",
				"href=\"/shop\"",
				"href=\"/metrics/live\"",
				"href=\"http://rocketweb.nl:8080\"",
				"href=\"http://docs.rocketweb.nl:8080/go-partial\"",
				"id=\"nav-joke\"",
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
				"src=\"/assets/img/logo-go-doc.png\"",
				"href=\"/go-partial\"",
				"href=\"/go-docs\" hx-get=\"/go-docs\"",
				"class=\"active\" aria-current=\"page\"",
				"hx-get=\"/go-docs/install\"",
				"href=\"/go-router\"",
				"href=\"https://gowebthings.com\"",
				"href=\"https://showcase.gowebthings.com\"",
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
				`<meta property="og:image" content="https://docs.gowebthings.com/assets/img/logo-go-router.png">`,
				"src=\"/assets/img/logo-go-router.png\"",
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
			path:       "/",
			wantStatus: http.StatusOK,
			wantBody: []string{
				"go-webthings showcase",
				"Server-rendered partials",
				"Webshop",
				"Live metrics",
				"Interaction helpers",
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
	if !strings.Contains(body, ".main-body") || !strings.Contains(body, ".main-intro") {
		t.Fatalf("expected main website stylesheet, got:\n%s", body)
	}
	if strings.Contains(body, ".docs-body") {
		t.Fatalf("main stylesheet should not be the docs stylesheet, got:\n%s", body)
	}
}

func TestMainBulletins(t *testing.T) {
	if len(mainBulletins) < 30 {
		t.Fatalf("expected at least 30 main bulletins, got %d", len(mainBulletins))
	}
	bulletin := randomBulletin()
	if bulletin == "" {
		t.Fatal("expected random bulletin")
	}
	if strings.Contains(bulletin, "rocketweb.nl") || strings.Contains(bulletin, "gowebthings.com") {
		t.Fatalf("bulletin should be fake news, not environment copy: %q", bulletin)
	}
}

func TestMainBulletinEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/bulletin", nil)
	req.Host = "gowebthings.com"
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `class="main-bulletin-line"`) {
		t.Fatalf("expected bulletin fragment, got:\n%s", body)
	}
	if strings.Contains(body, "<script") || strings.Contains(body, "<!doctype") {
		t.Fatalf("expected a small HTML fragment, got:\n%s", body)
	}
}

func TestMainGeneratorEndpoints(t *testing.T) {
	handler := NewRouter()

	t.Run("preview", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/generate/preview?text=Router", nil)
		req.Host = "gowebthings.com"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, `class="main-generator-image"`) || !strings.Contains(body, `/generate/image?text=Router`) {
			t.Fatalf("expected generator preview fragment, got:\n%s", body)
		}
	})

	t.Run("image", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/generate/image?text=Router", nil)
		req.Host = "gowebthings.com"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); got != "image/png" {
			t.Fatalf("expected image/png content type, got %q", got)
		}
		if rec.Body.Len() == 0 {
			t.Fatal("expected PNG body")
		}
	})
}

func TestHTMXMainNavigationReturnsContentFragment(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/generate", nil)
	req.Host = "gowebthings.com"
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "content")
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "<!doctype html>") || strings.Contains(body, "<body") {
		t.Fatalf("expected HTMX fragment without full shell, got:\n%s", body)
	}
	for _, want := range []string{
		`hx-get="/generate/preview"`,
		`id="main-navbar"`,
		`hx-swap-oob="true"`,
		`href="/generate" hx-get="/generate" hx-target="#content" hx-push-url="true" aria-current="page"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected body to contain %q\nbody:\n%s", want, body)
		}
	}
}

func TestShowcaseStylesheet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/assets/css/styles.css", nil)
	req.Host = "showcase.gowebthings.com"
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "tailwindcss") || !strings.Contains(body, "bg-\\[\\#040816\\]\\/90") {
		t.Fatalf("expected showcase stylesheet, got:\n%s", body)
	}
	if !strings.Contains(body, ".showcase-retro") || !strings.Contains(body, "max-\\[820px\\]\\:block") {
		t.Fatalf("expected retro showcase theme and generated responsive classes, got:\n%s", body)
	}
	if strings.Contains(body, ".docs-body") || strings.Contains(body, ".main-body") {
		t.Fatalf("showcase stylesheet should not be the docs or main stylesheet, got:\n%s", body)
	}
}

func TestShowcaseHTMXFragments(t *testing.T) {
	handler := NewRouter()

	t.Run("selection returns content and oob sidebar", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/selection", nil)
		req.Host = "showcase.gowebthings.com"
		req.Header.Set("HX-Request", "true")
		req.Header.Set("HX-Target", "content")
		req.Header.Set("X-Select", "details")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		body := rec.Body.String()
		for _, want := range []string{
			"Selection partials",
			"Details",
			"alternate partial was selected",
			`id="app-header"`,
			`id="showcase-sidebar"`,
			`id="nav-joke"`,
			`hx-swap-oob="true"`,
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("expected body to contain %q\nbody:\n%s", want, body)
			}
		}
		if strings.Contains(body, "<!doctype html>") || strings.Contains(body, "<body") {
			t.Fatalf("expected HTMX fragment without full shell, got:\n%s", body)
		}
	})

	t.Run("row target resolves one row", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/rows/refresh-row?id=1", nil)
		req.Host = "showcase.gowebthings.com"
		req.Header.Set("HX-Request", "true")
		req.Header.Set("HX-Target", "row-1")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, `id="row-1"`) || !strings.Contains(body, "Updated ") {
			t.Fatalf("expected refreshed row, got:\n%s", body)
		}
		if strings.Contains(body, `id="row-2"`) || strings.Contains(body, "<!doctype html>") {
			t.Fatalf("expected one row target response, got:\n%s", body)
		}
	})

	t.Run("cart add returns popup and oob button", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/shop/cart/add?id=1", nil)
		req.Host = "showcase.gowebthings.com"
		req.Header.Set("HX-Request", "true")
		req.Header.Set("HX-Target", "cart-popup")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		body := rec.Body.String()
		for _, want := range []string{
			`id="cart-popup"`,
			"Canvas Tote",
			`id="shop-cart-button"`,
			`hx-swap-oob="true"`,
			"1 items",
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("expected body to contain %q\nbody:\n%s", want, body)
			}
		}
	})
}

func TestShowcaseSSEFlushesThroughRouter(t *testing.T) {
	server := httptest.NewServer(NewRouter())
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/sse/stream", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "showcase.gowebthings.com"

	res, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	if got := res.Header.Get("Content-Type"); !strings.Contains(got, "text/event-stream") {
		t.Fatalf("expected event-stream content type, got %q", got)
	}

	reader := bufio.NewReader(res.Body)
	start := time.Now()
	var step1, step2 time.Duration
	for step2 == 0 {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read SSE line: %v", err)
		}
		switch {
		case strings.Contains(line, `"step":1`):
			step1 = time.Since(start)
		case strings.Contains(line, `"step":2`):
			step2 = time.Since(start)
		}
	}
	if step1 == 0 {
		t.Fatal("expected first progress signal")
	}
	if step2-step1 < 500*time.Millisecond {
		t.Fatalf("expected SSE progress to flush over time, got step1=%s step2=%s", step1, step2)
	}
}

func TestShowcaseInteractionStreamFlushesThroughRouter(t *testing.T) {
	server := httptest.NewServer(NewRouter())
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/interactions/stream", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "showcase.gowebthings.com"

	res, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	reader := bufio.NewReader(res.Body)
	start := time.Now()
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read initial SSE line: %v", err)
	}
	if !strings.Contains(line, ": connected") {
		t.Fatalf("expected initial stream comment, got %q", line)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("expected initial stream comment to flush immediately, got %s", elapsed)
	}

	var dataAt time.Duration
	for dataAt == 0 {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read interaction stream line: %v", err)
		}
		if strings.HasPrefix(line, "data: ") {
			dataAt = time.Since(start)
		}
	}
	if dataAt < time.Second {
		t.Fatalf("expected streamed data after the handler delay, got %s", dataAt)
	}
}

func TestShowcaseLiveMetricsStreamFlushesThroughRouter(t *testing.T) {
	server := httptest.NewServer(NewRouter())
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/metrics/live/stream", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "showcase.gowebthings.com"

	res, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	start := time.Now()
	line, err := bufio.NewReader(res.Body).ReadString('\n')
	if err != nil {
		t.Fatalf("read live metrics stream line: %v", err)
	}
	if !strings.Contains(line, ": go-partial live metrics") {
		t.Fatalf("expected live metrics stream comment, got %q", line)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("expected live metrics stream comment to flush immediately, got %s", elapsed)
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
	for _, tmpl := range []string{
		"templates/layout.gohtml",
		"templates/navbar.gohtml",
		"templates/bullitin.gohtml",
		"templates/webring.gohtml",
	} {
		if _, err := fs.Stat(mainFS, tmpl); err != nil {
			t.Fatalf("expected main website shared template %s to exist: %v", tmpl, err)
		}
	}
	if _, err := fs.Stat(mainFS, "templates/generate.gohtml"); err != nil {
		t.Fatalf("expected main generator template to exist: %v", err)
	}
	if _, err := fs.Stat(mainFS, "assets/css/styles.css"); err != nil {
		t.Fatalf("expected main website stylesheet to exist: %v", err)
	}
	for _, tmpl := range []string{
		"templates/shell.gohtml",
		"templates/header.gohtml",
		"templates/home.gohtml",
		"templates/rows.gohtml",
		"templates/selection.gohtml",
		"templates/tabs.gohtml",
		"templates/action.gohtml",
		"templates/async.gohtml",
		"templates/interactions.gohtml",
		"templates/shop.gohtml",
		"templates/shop_cart_button.gohtml",
		"templates/shop_cart_popup.gohtml",
		"templates/metrics.gohtml",
		"templates/live_metrics.gohtml",
		"templates/logger.gohtml",
		"templates/sse.gohtml",
	} {
		if _, err := fs.Stat(showcaseFS, tmpl); err != nil {
			t.Fatalf("expected showcase website template %s to exist: %v", tmpl, err)
		}
	}
	if _, err := fs.Stat(showcaseFS, "assets/css/styles.css"); err != nil {
		t.Fatalf("expected showcase website stylesheet to exist: %v", err)
	}
}

func TestShowcaseAssetsAreScoped(t *testing.T) {
	entries, err := fs.ReadDir(showcaseFS, "assets/img")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"favicon.ico":           true,
		"logo-go-doc.png":       true,
		"logo-go-partial.png":   true,
		"logo-go-router.png":    true,
		"logo-go-webthings.png": true,
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !want[entry.Name()] {
			t.Fatalf("showcase asset %s is not used by the showcase template", entry.Name())
		}
		delete(want, entry.Name())
	}
	for name := range want {
		t.Fatalf("expected showcase asset %s to exist", name)
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
