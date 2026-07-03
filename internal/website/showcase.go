package website

import (
	"fmt"
	"net/http"

	partial "github.com/donseba/go-partial"
	router "github.com/donseba/go-router"
)

func registerShowcaseRoutes(r *router.Router, domain string) {
	r.Get("/", showcaseIndex).As(fmt.Sprintf("%s.showcase.index", domain))
	r.Get("/{element}", showcaseElement).As(fmt.Sprintf("%s.showcase.element", domain))
}

func showcaseIndex(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		renderNotFound(w, req)
		return
	}
	renderShowcaseComingSoon(w, req, "")
}

func showcaseElement(w http.ResponseWriter, req *http.Request) {
	renderShowcaseComingSoon(w, req, req.PathValue("element"))
}

func renderShowcaseComingSoon(w http.ResponseWriter, req *http.Request, slug string) {
	title := "Showcase coming soon"
	description := "Interactive go-webthings examples are being prepared. For now, the docs are ready."
	if slug != "" {
		if _, ok := findElement(slug); !ok {
			renderNotFound(w, req)
			return
		}
		title = fmt.Sprintf("%s showcase coming soon", nameFromSlug(slug))
		description = fmt.Sprintf("The %s showcase is coming soon. Use the docs while the live examples are being prepared.", slug)
	}

	data := PageData{
		Title:       title,
		Section:     SectionShowcase,
		Host:        req.Host,
		Elements:    sectionElements(),
		Description: description,
		ComingSoon:  true,
	}
	page := partial.NewID("content", "templates/page.gohtml").SetFileSystem(showcaseFS).SetDot(data)
	renderStandalonePage(w, req, http.StatusOK, data, page)
}
