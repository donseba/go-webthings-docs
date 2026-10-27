package website

import (
	"html"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"donseba/go-webthings-docs/internal/moduledoc"
)

func get(t *testing.T, host, path string, fragment bool) (int, string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Host = host

	if fragment {
		setHTMXPartialRequest(req, "main#content")
	}

	rec := httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, req)

	return rec.Code, strings.Join(strings.Fields(rec.Body.String()), " ")
}

func TestModulesData(t *testing.T) {
	if len(moduleSite.modules) == 0 {
		t.Fatal("no modules loaded from deploy/website/docs/modules")
	}

	for _, module := range moduleSite.modules {
		if module.Name == "" || module.Summary == "" {
			t.Errorf("%s: missing name or summary", module.Slug)
		}

		if !slices.Contains(moduledoc.Groups, module.Group) {
			t.Errorf("%s: unknown group %q", module.Slug, module.Group)
		}

		if module.Guide == "" {
			t.Errorf("%s: no guide; add modules/%s/README.md to go-webthings and run task modules", module.Slug, module.Slug)
		}

		if !strings.HasPrefix(module.Constructor.Signature, "func Module(") {
			t.Errorf("%s: constructor %q", module.Slug, module.Constructor.Signature)
		}

		if module.Example.Code == "" || module.Example.Key == "" {
			t.Errorf("%s: missing example instance", module.Slug)
		}
	}
}

func TestModulesIndex(t *testing.T) {
	for _, host := range []string{"modules.gowebthings.com", "modules.rocketweb.nl:8080", "modules.localhost:8080"} {
		code, body := get(t, host, "/", false)
		if code != http.StatusOK {
			t.Fatalf("%s: status %d", host, code)
		}

		for _, module := range moduleSite.modules {
			link := `href="/` + module.Slug + `" hx-get="/` + module.Slug + `" hx-target="#content" hx-push-url="true"`
			if !strings.Contains(body, link) {
				t.Fatalf("%s: index is missing %s", host, module.Slug)
			}
		}

		if strings.Contains(body, `href="/go-partial"`) {
			t.Fatalf("%s: the modules header should not list the components", host)
		}
	}
}

func TestModulePages(t *testing.T) {
	for _, module := range moduleSite.modules {
		for _, page := range moduledoc.Pages {
			path := modulePath(module.Slug, page.Path)

			t.Run(path, func(t *testing.T) {
				code, body := get(t, "modules.gowebthings.com", path, false)
				if code != http.StatusOK {
					t.Fatalf("status %d: %s", code, body)
				}

				for _, want := range []string{
					html.EscapeString(module.Name),
					`<link rel="canonical" href="https://modules.gowebthings.com` + path + `">`,
					`href="` + path + `" hx-get="` + path + `" hx-target="#content" hx-push-url="true" aria-current="page"`,
					`href="https://docs.gowebthings.com"`,
				} {
					if !strings.Contains(body, want) {
						t.Fatalf("page is missing %q", want)
					}
				}

				if strings.Count(body, `aria-current="page"`) != 1 {
					t.Fatal("exactly one sidebar link should be current")
				}

				code, body = get(t, "modules.gowebthings.com", path, true)
				if code != http.StatusOK {
					t.Fatalf("fragment status %d", code)
				}

				if strings.Contains(body, "<!doctype html>") || !strings.Contains(body, `hx-swap-oob="true"`) {
					t.Fatal("fragment should be content with out-of-band navigation")
				}

				rail := `href="/` + module.Slug + `" hx-get="/` + module.Slug + `" hx-target="#content" hx-push-url="true" aria-current="true"`
				if !strings.Contains(body, `id="docs-rail"`) || !strings.Contains(body, rail) {
					t.Fatal("fragment should refresh the rail with the module marked")
				}
			})
		}
	}
}

func TestModulesNotFound(t *testing.T) {
	for _, path := range []string{"/nope", "/publishing/nope", "/publishing/templates/x", "/publishing/"} {
		code, body := get(t, "modules.gowebthings.com", path, false)
		if code != http.StatusNotFound {
			t.Fatalf("%s: status %d", path, code)
		}

		if !strings.Contains(body, "Module not found") {
			t.Fatalf("%s: missing the not found page", path)
		}
	}
}

func TestModulesAssets(t *testing.T) {
	for _, path := range []string{"/assets/css/styles.css", "/assets/img/icons.svg", "/favicon.ico"} {
		if code, _ := get(t, "modules.gowebthings.com", path, false); code != http.StatusOK {
			t.Fatalf("%s: status %d", path, code)
		}
	}
}

func TestMainModulesPage(t *testing.T) {
	code, body := get(t, "rocketweb.nl:8080", "/modules", false)
	if code != http.StatusOK {
		t.Fatalf("status %d", code)
	}

	if !strings.Contains(body, `href="/modules" hx-get="/modules" hx-target="#content" hx-push-url="true" aria-current="page"`) {
		t.Fatal("navbar should mark Modules as current")
	}

	for _, module := range moduleSite.modules {
		if !strings.Contains(body, `href="http://modules.rocketweb.nl:8080/`+module.Slug+`"`) {
			t.Fatalf("missing a link to the %s docs", module.Slug)
		}
	}
}

func TestDocsHeaderLinksToModules(t *testing.T) {
	code, body := get(t, "docs.gowebthings.com", "/go-form", false)
	if code != http.StatusOK {
		t.Fatalf("status %d", code)
	}

	if !strings.Contains(body, `href="https://modules.gowebthings.com"`) {
		t.Fatal("docs header should link to the modules")
	}
}
