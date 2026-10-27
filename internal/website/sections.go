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
	registerGoClueDocsRoutes(r, domain)
	registerGoRouterDocsRoutes(r, domain)
	registerComponentDocsRoutes(r, domain)
	registerGuideRoutes(r, domain)
}

// DocsIndexData is the dot of the docs home page.
type DocsIndexData struct {
	Article  DocsArticleData
	Elements []Element
}

// docsHome renders the docs home page: the rail and a card per component.
var docsHome = newDocsRenderer(docsRendererConfig{
	BasePath:  "/",
	AppName:   "docs",
	LogoImage: "/assets/img/logo-go-webthings.png",
	Title:     "go-webthings",
	Subtitle:  "html-first go components",
})

func docsIndex(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		renderNotFound(w, req)
		return
	}

	page := docsPage{
		Template:    "templates/general/index.gohtml",
		Title:       "Docs for go-webthings",
		Description: "Choose a go-webthings element to view its docs.",
		Section:     "docs",
	}

	elements := append([]Element{{
		Slug:        "go-webthings",
		Name:        "go-webthings",
		Description: "The backend of your website: guides for building a site with pages, modules, themes and languages.",
		Link:        guidesBase,
	}}, sectionElements()...)
	for i := range elements {
		elements[i].Image = elementImage(elements[i].Slug)
	}

	data := DocsIndexData{
		Article:  article(page),
		Elements: elements,
	}

	docsHome.write(w, req, http.StatusOK, page, data, DocsNavPage{})
}

func docsElement(w http.ResponseWriter, req *http.Request) {
	slug := req.PathValue("element")
	if slug == "go-partial" {
		goPartialDocs.docs.render(w, req, goPartialDocs.pages["/"], nil)
		return
	}
	if slug == "go-clue" {
		goClueDocs.docs.render(w, req, goClueDocs.pages["/"], nil)
		return
	}
	if slug == "go-router" {
		goRouterDocs.docs.render(w, req, goRouterDocs.pages["/"], nil)
		return
	}

	if app, ok := componentDocs[slug]; ok {
		app.docs.render(w, req, app.pages["/"], nil)
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
		http.NotFound(w, req)
	case SectionModules:
		moduleSite.notFound(w, req)
	case SectionDocs:
		page := partial.NewID("content", "templates/general/page.gohtml").SetFileSystem(websiteFS).SetDot(data)
		renderStandalonePage(w, req, http.StatusNotFound, data, page)
	default:
		page := partial.NewID("content", "templates/page.gohtml").SetFileSystem(mainFS).SetDot(data)
		renderMainShell(w, req, http.StatusNotFound, data, page)
	}
}
