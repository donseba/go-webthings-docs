package website

import (
	"fmt"
	"net/http"

	partial "github.com/donseba/go-partial"
	router "github.com/donseba/go-router"
)

func registerDocsRoutes(r *router.Router, domain string) {
	r.Get("/", docsIndex).As(fmt.Sprintf("%s.docs.index", domain))
	r.Get("/{element}", docsElement).As(fmt.Sprintf("%s.docs.element", domain))
	registerGoPartialDocsRoutes(r, domain)
	registerGoDocsDocsRoutes(r, domain)
	registerGoRouterDocsRoutes(r, domain)
}

func docsIndex(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		renderNotFound(w, req)
		return
	}

	data := PageData{
		Title:       "Docs for go-webthings",
		Section:     SectionDocs,
		Host:        req.Host,
		Elements:    sectionElements(),
		Description: "Choose a go-webthings element to view its docs.",
	}
	page := partial.NewID("content", "templates/general/page.gohtml").SetFileSystem(websiteFS).SetDot(data)
	renderStandalonePage(w, req, http.StatusOK, data, page)
}

func docsElement(w http.ResponseWriter, req *http.Request) {
	slug := req.PathValue("element")
	if slug == "go-partial" {
		goPartialDocs.docs.render(w, req, goPartialDocs.pages["/"], nil)
		return
	}
	if slug == "go-docs" {
		goDocsDocs.docs.render(w, req, goDocsDocs.pages["/"], nil)
		return
	}
	if slug == "go-router" {
		goRouterDocs.docs.render(w, req, goRouterDocs.pages["/"], nil)
		return
	}

	element, ok := findElement(slug)
	if !ok {
		renderNotFound(w, req)
		return
	}

	data := PageData{
		Title:       fmt.Sprintf("%s docs", element.Name),
		Section:     SectionDocs,
		Host:        req.Host,
		Element:     &element,
		Elements:    sectionElements(),
		Description: sectionDescription(SectionDocs, element),
	}
	page := partial.NewID("content", "templates/general/page.gohtml").SetFileSystem(websiteFS).SetDot(data)
	renderStandalonePage(w, req, http.StatusOK, data, page)
}

func renderNotFound(w http.ResponseWriter, req *http.Request) {
	data := PageData{
		Title:       "Element not found",
		Section:     sectionFromHost(req.Host),
		Host:        req.Host,
		Elements:    sectionElements(),
		Description: "This element is not registered for go-webthings docs yet.",
	}
	switch data.Section {
	case SectionShowcase:
		page := partial.NewID("content", "templates/page.gohtml").SetFileSystem(showcaseFS).SetDot(data)
		renderStandalonePage(w, req, http.StatusNotFound, data, page)
	case SectionDocs:
		page := partial.NewID("content", "templates/general/page.gohtml").SetFileSystem(websiteFS).SetDot(data)
		renderStandalonePage(w, req, http.StatusNotFound, data, page)
	default:
		page := partial.NewID("content", "templates/page.gohtml").SetFileSystem(mainFS).SetDot(data)
		renderMainShell(w, req, http.StatusNotFound, data, page)
	}
}
