package site

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	partial "github.com/donseba/go-partial"
	router "github.com/donseba/go-router"
)

var siteFS fs.FS = siteFileSystem()

var (
	elementNames = []string{
		"go-partial",
		"go-docs",
		"go-router",
	}

	goPartialDocs = mustNewGoPartialDocs()
	goDocsDocs    = mustNewGoDocsDocs()
	goRouterDocs  = mustNewGoRouterDocs()
)

type Section string

const (
	SectionDocs     Section = "docs"
	SectionShowcase Section = "showcase"
)

type Element struct {
	Slug        string
	Name        string
	Description string
	Image       string
}

type PageData struct {
	Title       string
	Section     Section
	Host        string
	Element     *Element
	Elements    []Element
	Production  string
	Local       string
	Description string
}

type NavItem struct {
	Path  string
	Label string
	Group string
}

func NewRouter() http.Handler {
	r := router.New(http.NewServeMux(), "go-webthings docs", "0.1.0")
	r.HandleStatus(http.StatusNotFound, renderNotFound)

	registerDomain(r, "rocketweb.nl")
	registerDomain(r, "go-webthings.com")

	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "https://docs.go-webthings.com/go-router", http.StatusTemporaryRedirect)
	}).As("home")

	return r
}

func registerDomain(r *router.Router, domain string) {
	r.Subdomain("docs", domain, func(docs *router.Router) {
		registerStaticRoutes(docs)
		registerSectionRoutes(docs, SectionDocs, domain)
	})

	r.Subdomain("showcase", domain, func(showcase *router.Router) {
		registerStaticRoutes(showcase)
		registerSectionRoutes(showcase, SectionShowcase, domain)
	})
}

func registerStaticRoutes(r *router.Router) {
	r.ServeFiles("/assets/", http.FS(mustSubFS(siteFS, "assets")))
}

func registerSectionRoutes(r *router.Router, section Section, domain string) {
	r.Get("/", sectionIndex(section, domain)).As(fmt.Sprintf("%s.%s.index", domain, section))
	r.Get("/{element}", sectionElement(section, domain)).As(fmt.Sprintf("%s.%s.element", domain, section))

	if section == SectionDocs {
		registerGoPartialDocsRoutes(r, domain)
		registerGoDocsDocsRoutes(r, domain)
		registerGoRouterDocsRoutes(r, domain)
	}
}

func sectionIndex(section Section, domain string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			renderNotFound(w, req)
			return
		}

		renderPage(w, http.StatusOK, PageData{
			Title:       fmt.Sprintf("%s for go-webthings", title(section)),
			Section:     section,
			Host:        req.Host,
			Elements:    elements(),
			Production:  fmt.Sprintf("https://%s.go-webthings.com/:element", section),
			Local:       fmt.Sprintf("http://%s.rocketweb.nl:8080/:element", section),
			Description: fmt.Sprintf("Choose a go-webthings element to view its %s.", section),
		})
	}
}

func sectionElement(section Section, domain string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		slug := req.PathValue("element")
		if section == SectionDocs && slug == "go-partial" {
			goPartialDocs.overview(w, req)
			return
		}
		if section == SectionDocs && slug == "go-docs" {
			goDocsDocs.overview(w, req)
			return
		}
		if section == SectionDocs && slug == "go-router" {
			goRouterDocs.overview(w, req)
			return
		}

		element, ok := findElement(slug)
		if !ok {
			renderNotFound(w, req)
			return
		}

		renderPage(w, http.StatusOK, PageData{
			Title:       fmt.Sprintf("%s %s", element.Name, title(section)),
			Section:     section,
			Host:        req.Host,
			Element:     &element,
			Elements:    elements(),
			Production:  fmt.Sprintf("https://%s.go-webthings.com/%s", section, element.Slug),
			Local:       fmt.Sprintf("http://%s.rocketweb.nl:8080/%s", section, element.Slug),
			Description: sectionDescription(section, element),
		})
	}
}

func renderNotFound(w http.ResponseWriter, req *http.Request) {
	renderPage(w, http.StatusNotFound, PageData{
		Title:       "Element not found",
		Host:        req.Host,
		Elements:    elements(),
		Description: "This element is not registered for go-webthings docs yet.",
	})
}

func renderPage(w http.ResponseWriter, status int, data PageData) {
	page := partial.NewID("content", "templates/general/page.gohtml").
		SetFileSystem(siteFS).
		SetDot(data)
	out, err := partial.Render(context.Background(), page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(out))
}

func siteFileSystem() fs.FS {
	if dir := os.Getenv("ASSET_DIR"); dir != "" {
		return os.DirFS(dir)
	}
	if _, err := os.Stat("deploy/docs/templates"); err == nil {
		return os.DirFS("deploy/docs")
	}
	if _, err := os.Stat("templates"); err == nil {
		return os.DirFS(".")
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
		deployDir := filepath.Join(root, "deploy", "docs")
		if _, err := os.Stat(filepath.Join(deployDir, "templates")); err == nil {
			return os.DirFS(deployDir)
		}
	}
	return os.DirFS(".")
}

func mustSubFS(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}

func findElement(slug string) (Element, bool) {
	for _, element := range elements() {
		if element.Slug == slug {
			return element, true
		}
	}

	return Element{}, false
}

func elements() []Element {
	items := make([]Element, 0, len(elementNames))
	for _, slug := range elementNames {
		items = append(items, Element{
			Slug:        slug,
			Name:        nameFromSlug(slug),
			Description: elementDescription(slug),
			Image:       elementImage(slug),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Slug < items[j].Slug
	})

	return items
}

func nameFromSlug(slug string) string {
	words := strings.Fields(strings.ReplaceAll(slug, "-", " "))
	for i, word := range words {
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}

	return strings.Join(words, " ")
}

func elementImage(slug string) string {
	switch slug {
	case "go-partial":
		return "/assets/img/go-partial-300.png"
	case "go-docs":
		return "/assets/img/go-doc-300.png"
	case "go-router":
		return "/assets/img/go-router-300.png"
	default:
		return "/assets/img/go-webthings-300.png"
	}
}

func elementDescription(slug string) string {
	switch slug {
	case "go-partial":
		return "Partial and full-page rendering for Go templates."
	case "go-docs":
		return "Typed editor tooling and diagnostics for Go templates."
	case "go-router":
		return "Host-aware HTTP routing built on top of net/http."
	default:
		return "A go-webthings element."
	}
}

func sectionDescription(section Section, element Element) string {
	switch section {
	case SectionDocs:
		return fmt.Sprintf("Documentation landing page for %s.", element.Slug)
	case SectionShowcase:
		return fmt.Sprintf("Showcase landing page for %s.", element.Slug)
	default:
		return element.Description
	}
}

func title(section Section) string {
	switch section {
	case SectionDocs:
		return "Docs"
	case SectionShowcase:
		return "Showcase"
	default:
		return "Page"
	}
}
