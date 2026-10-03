package website

import (
	"html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestComponentDocumentationSections(t *testing.T) {
	handler := NewRouter()
	for slug, app := range componentDocs {
		for _, nav := range app.docs.nav {
			if nav.Path == "/" {
				continue
			}
			path := "/" + slug + nav.Path
			page := app.pages[nav.Path]
			for _, host := range []string{"docs.gowebthings.com", "docs.rocketweb.nl:8080"} {
				for _, fragment := range []bool{false, true} {
					name := host + path
					if fragment {
						name += "/fragment"
					}
					t.Run(name, func(t *testing.T) {
						req := httptest.NewRequest(http.MethodGet, path, nil)
						req.Host = host
						if fragment {
							req.Header.Set("HX-Request", "true")
							req.Header.Set("HX-Target", "content")
						}
						rec := httptest.NewRecorder()
						handler.ServeHTTP(rec, req)
						if rec.Code != http.StatusOK {
							t.Fatalf("section returned %d: %s", rec.Code, rec.Body.String())
						}
						body := strings.Join(strings.Fields(rec.Body.String()), " ")
						for _, want := range []string{
							html.EscapeString(page.Title),
							app.docs.subtitle,
							`href="` + path + `" hx-get="` + path + `" hx-target="#content" hx-push-url="true" aria-current="page"`,
						} {
							if !strings.Contains(body, want) {
								t.Fatalf("section missing %q", want)
							}
						}
						for _, item := range app.docs.nav {
							href := "/" + slug
							if item.Path != "/" {
								href += item.Path
							}
							if !strings.Contains(body, `href="`+href+`" hx-get="`+href+`"`) {
								t.Fatalf("section is missing navigation to %s", href)
							}
						}
						if fragment {
							if strings.Contains(body, "<!doctype html>") || strings.Contains(body, "<body") {
								t.Fatal("fragment response contains a full document")
							}
							if !strings.Contains(body, `hx-swap-oob="true"`) {
								t.Fatal("fragment response must refresh the navigation")
							}
						} else if !strings.Contains(body, `href="https://docs.gowebthings.com`+path+`"`) {
							t.Fatal("section is missing its canonical URL")
						}
					})
				}
			}
		}
	}
}

func TestComponentDocumentation(t *testing.T) {
	handler := NewRouter()
	for _, component := range []struct {
		slug  string
		title string
	}{
		{"go-form", "HTML forms from Go structs"},
		{"go-importmap", "Prepare JavaScript and CSS assets"},
		{"go-translator", "Translate Go templates with gettext"},
	} {
		for _, host := range []string{"docs.gowebthings.com", "docs.rocketweb.nl:8080"} {
			for _, fragment := range []bool{false, true} {
				name := component.slug + "/" + host
				if fragment {
					name += "/fragment"
				}
				t.Run(name, func(t *testing.T) {
					req := httptest.NewRequest(http.MethodGet, "/"+component.slug, nil)
					req.Host = host
					if fragment {
						req.Header.Set("HX-Request", "true")
						req.Header.Set("HX-Target", "content")
					}
					rec := httptest.NewRecorder()
					handler.ServeHTTP(rec, req)
					if rec.Code != http.StatusOK {
						t.Fatalf("expected status 200, got %d", rec.Code)
					}
					body := rec.Body.String()
					for _, want := range []string{
						component.title,
						`src="/assets/img/logo-` + component.slug + `.png"`,
						`href="/` + component.slug + `" hx-get="/` + component.slug + `" hx-target="#content" hx-push-url="true" aria-current="page"`,
						`href="/go-form"`, `href="/go-importmap"`, `href="/go-translator"`,
					} {
						if !strings.Contains(body, want) {
							t.Fatalf("response missing %q", want)
						}
					}
					if fragment {
						if strings.Contains(body, "<!doctype html>") || strings.Contains(body, "<body") {
							t.Fatal("fragment response contains a full document")
						}
						if !strings.Contains(body, `hx-swap-oob="true"`) {
							t.Fatal("fragment response must refresh the component navigation")
						}
					} else if !strings.Contains(body, `content="https://docs.gowebthings.com/assets/img/logo-`+component.slug+`.png"`) {
						t.Fatal("full page must use its component logo for the social preview")
					}
				})
			}
		}
	}
}

func TestComponentDiscovery(t *testing.T) {
	handler := NewRouter()
	for _, route := range []struct {
		host   string
		path   string
		prefix string
	}{
		{"docs.gowebthings.com", "/", "/"},
		{"gowebthings.com", "/components", "https://docs.gowebthings.com/"},
		{"gowebthings.com", "/", "https://docs.gowebthings.com/"},
	} {
		t.Run(route.host+route.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, route.path, nil)
			req.Host = route.host
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", rec.Code)
			}
			slugs := []string{"go-form", "go-importmap", "go-translator", "go-partial", "go-router"}
			if route.host == "gowebthings.com" && route.path == "/" {
				slugs = slugs[:3]
			}
			for _, slug := range slugs {
				wants := []string{`href="` + route.prefix + slug + `"`}
				if route.host != "gowebthings.com" || route.path != "/" {
					wants = append(wants, `src="/assets/img/logo-`+slug+`.png"`)
				}
				for _, want := range wants {
					if !strings.Contains(rec.Body.String(), want) {
						t.Fatalf("component index missing %q", want)
					}
				}
				assetReq := httptest.NewRequest(http.MethodGet, "/assets/img/logo-"+slug+".png", nil)
				assetReq.Host = route.host
				assetRec := httptest.NewRecorder()
				handler.ServeHTTP(assetRec, assetReq)
				if assetRec.Code != http.StatusOK || assetRec.Header().Get("Content-Type") != "image/png" {
					t.Fatalf("component %s logo is not served as a PNG", slug)
				}
			}
		})
	}
}
