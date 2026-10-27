package showcase

import (
	"net/http/httptest"
	"testing"
)

func TestShowcaseLinksKeepTheLocalScheme(t *testing.T) {
	tests := map[string]struct {
		docs string
		main string
	}{
		"showcase.gowebthings.com":   {"https://docs.gowebthings.com/go-partial", "https://gowebthings.com"},
		"showcase.rocketweb.nl:8080": {"http://docs.rocketweb.nl:8080/go-partial", "http://rocketweb.nl:8080"},
		"showcase.localhost:8080":    {"http://docs.localhost:8080/go-partial", "http://localhost:8080"},
		"localhost:8080":             {"http://docs.localhost:8080/go-partial", "http://localhost:8080"},
	}

	for host, want := range tests {
		req := httptest.NewRequest("GET", "/", nil)
		req.Host = host

		if got := showcaseDocsURL(req); got != want.docs {
			t.Errorf("%s: docs = %q, want %q", host, got, want.docs)
		}
		if got := showcaseMainURL(req); got != want.main {
			t.Errorf("%s: main = %q, want %q", host, got, want.main)
		}
	}
}
