package website

import (
	"html"
	"net/http"
	"strings"
	"testing"
)

func TestGuidesData(t *testing.T) {
	if len(guideDocs.order) == 0 {
		t.Fatal("no guides loaded from deploy/website/docs/guides")
	}

	for _, guide := range guideDocs.order {
		if guide.Title == "" || guide.Summary == "" || guide.Body == "" {
			t.Errorf("%s: missing title, summary or body", guide.Slug)
		}
	}
}

func TestGuidesOverview(t *testing.T) {
	code, body := get(t, "docs.gowebthings.com", guidesBase, false)
	if code != http.StatusOK {
		t.Fatalf("status %d", code)
	}

	rail := `href="/go-webthings" hx-get="/go-webthings" hx-target="#content" hx-push-url="true" aria-current="true"`
	if !strings.Contains(body, rail) {
		t.Fatal("rail should mark go-webthings")
	}

	for _, guide := range guideDocs.order {
		if !strings.Contains(body, `href="`+guidesBase+"/"+guide.Slug+`"`) {
			t.Fatalf("overview is missing %s", guide.Slug)
		}
	}
}

func TestGuidePages(t *testing.T) {
	for _, guide := range guideDocs.order {
		path := guidesBase + "/" + guide.Slug

		t.Run(path, func(t *testing.T) {
			code, body := get(t, "docs.gowebthings.com", path, false)
			if code != http.StatusOK {
				t.Fatalf("status %d", code)
			}

			for _, want := range []string{
				html.EscapeString(guide.Title),
				`<link rel="canonical" href="https://docs.gowebthings.com` + path + `">`,
				`href="` + path + `" hx-get="` + path + `" hx-target="#content" hx-push-url="true" aria-current="page"`,
			} {
				if !strings.Contains(body, want) {
					t.Fatalf("page is missing %q", want)
				}
			}

			if strings.Contains(body, `href="wt-`) {
				t.Fatal("generated cross-site links must be resolved")
			}
		})
	}
}

func TestResolveLinks(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Host = "docs.rocketweb.nl:8080"

	got := resolveLinks(req, `<a href="wt-modules:/publishing">x</a> <a href="wt-docs:/go-webthings/templates#layout">y</a>`)
	want := `<a href="http://modules.rocketweb.nl:8080/publishing">x</a> <a href="http://docs.rocketweb.nl:8080/go-webthings/templates#layout">y</a>`

	if string(got) != want {
		t.Fatalf("resolveLinks:\n got %s\nwant %s", got, want)
	}
}
