package website

import (
	"context"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	partial "github.com/donseba/go-partial"
	router "github.com/donseba/go-router"
)

var (
	websiteFS fs.FS = docsFileSystem()
	mainFS    fs.FS = mainFileSystem()
)

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
	SectionMain     Section = "main"
	SectionDocs     Section = "docs"
	SectionShowcase Section = "showcase"
)

type Element struct {
	Slug        string
	Name        string
	Description string
	Image       string
	Link        string
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
	SEO         SEOData
	ComingSoon  bool
}

type NavItem struct {
	Path  string
	Label string
	Group string
}

type SEOData struct {
	Title       string
	Description string
	Canonical   string
	Image       string
	SiteName    string
	Type        string
}

func NewRouter() http.Handler {
	r := router.New(http.NewServeMux(), "go-webthings docs", "0.1.0")
	r.HandleStatus(http.StatusNotFound, renderNotFound)

	registerDomain(r, "rocketweb.nl")
	registerDomain(r, "gowebthings.com")

	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "https://docs.gowebthings.com/go-router", http.StatusTemporaryRedirect)
	}).As("home")

	return r
}

func registerDomain(r *router.Router, domain string) {
	r.Host(domain, func(main *router.Router) {
		registerMainRoutes(main, domain, "main")
	})

	r.Subdomain("www", domain, func(main *router.Router) {
		registerMainRoutes(main, domain, "www.main")
	})

	r.Subdomain("docs", domain, func(docs *router.Router) {
		registerStaticRoutes(docs, websiteFS)
		registerSectionRoutes(docs, SectionDocs, domain)
	})

	r.Subdomain("showcase", domain, func(showcase *router.Router) {
		registerStaticRoutes(showcase, websiteFS)
		registerSectionRoutes(showcase, SectionShowcase, domain)
	})
}

func registerMainRoutes(r *router.Router, domain, routeScope string) {
	registerStaticRoutes(r, mainFS)
	r.Get("/", mainIndex(domain)).As(fmt.Sprintf("%s.%s.index", domain, routeScope))
	r.Get("/{element}", mainElement(domain)).As(fmt.Sprintf("%s.%s.element", domain, routeScope))
}

func registerStaticRoutes(r *router.Router, fsys fs.FS) {
	r.ServeFiles("/assets/", http.FS(mustSubFS(fsys, "assets")))
	r.Get("/favicon.ico", serveAssetFile(fsys, "assets/img/favicon.ico"))
}

func serveAssetFile(fsys fs.FS, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		body, err := fs.ReadFile(fsys, path)
		if err != nil {
			renderNotFound(w, req)
			return
		}
		if contentType := mime.TypeByExtension(filepath.Ext(path)); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		_, _ = w.Write(body)
	}
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

func mainIndex(domain string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			renderNotFound(w, req)
			return
		}

		renderPage(w, http.StatusOK, PageData{
			Title:       "go-webthings",
			Section:     SectionMain,
			Host:        req.Host,
			Elements:    mainElements(req),
			Production:  "https://gowebthings.com",
			Local:       fmt.Sprintf("http://%s:8080", domain),
			Description: "Composable Go packages for server-rendered websites, docs, routing, and interactive partials.",
		})
	}
}

func mainElement(domain string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		element, ok := findElement(req.PathValue("element"))
		if !ok {
			renderNotFound(w, req)
			return
		}

		element.Link = docsElementURL(req, element.Slug)
		renderPage(w, http.StatusOK, PageData{
			Title:       element.Name,
			Section:     SectionMain,
			Host:        req.Host,
			Element:     &element,
			Elements:    mainElements(req),
			Production:  fmt.Sprintf("https://gowebthings.com/%s", element.Slug),
			Local:       fmt.Sprintf("http://%s:8080/%s", domain, element.Slug),
			Description: element.Description,
		})
	}
}

func sectionIndex(section Section, domain string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			renderNotFound(w, req)
			return
		}
		if section == SectionShowcase {
			renderShowcaseComingSoon(w, req, domain, "")
			return
		}

		renderPage(w, http.StatusOK, PageData{
			Title:       fmt.Sprintf("%s for go-webthings", title(section)),
			Section:     section,
			Host:        req.Host,
			Elements:    sectionElements(),
			Production:  fmt.Sprintf("https://%s.gowebthings.com/:element", section),
			Local:       fmt.Sprintf("http://%s.rocketweb.nl:8080/:element", section),
			Description: fmt.Sprintf("Choose a go-webthings element to view its %s.", section),
		})
	}
}

func sectionElement(section Section, domain string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		slug := req.PathValue("element")
		if section == SectionShowcase {
			renderShowcaseComingSoon(w, req, domain, slug)
			return
		}
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
			Elements:    sectionElements(),
			Production:  fmt.Sprintf("https://%s.gowebthings.com/%s", section, element.Slug),
			Local:       fmt.Sprintf("http://%s.rocketweb.nl:8080/%s", section, element.Slug),
			Description: sectionDescription(section, element),
		})
	}
}

func renderShowcaseComingSoon(w http.ResponseWriter, req *http.Request, domain, slug string) {
	title := "Showcase coming soon"
	description := "Interactive go-webthings examples are being prepared. For now, the docs are ready."
	production := "https://showcase.gowebthings.com"
	local := fmt.Sprintf("http://showcase.%s:8080", domain)
	if slug != "" {
		if _, ok := findElement(slug); !ok {
			renderNotFound(w, req)
			return
		}
		title = fmt.Sprintf("%s showcase coming soon", nameFromSlug(slug))
		description = fmt.Sprintf("The %s showcase is coming soon. Use the docs while the live examples are being prepared.", slug)
		production = fmt.Sprintf("%s/%s", production, slug)
		local = fmt.Sprintf("%s/%s", local, slug)
	}

	renderPage(w, http.StatusOK, PageData{
		Title:       title,
		Section:     SectionShowcase,
		Host:        req.Host,
		Elements:    sectionElements(),
		Production:  production,
		Local:       local,
		Description: description,
		ComingSoon:  true,
	})
}

func renderNotFound(w http.ResponseWriter, req *http.Request) {
	renderPage(w, http.StatusNotFound, PageData{
		Title:       "Element not found",
		Section:     sectionFromHost(req.Host),
		Host:        req.Host,
		Elements:    sectionElements(),
		Description: "This element is not registered for go-webthings docs yet.",
	})
}

func renderPage(w http.ResponseWriter, status int, data PageData) {
	data.SEO = mainSEO(data)
	page := partial.NewID("content", pageTemplate(data.Section)).
		SetFileSystem(pageFileSystem(data.Section)).
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

func mainSEO(data PageData) SEOData {
	pageTitle := data.Title
	if data.Section != "" && data.Section != SectionMain {
		pageTitle = fmt.Sprintf("%s - %s", data.Title, title(data.Section))
	}
	description := data.Description
	if description == "" {
		description = "Composable Go packages for server-rendered websites, docs, routing, and interactive partials."
	}
	image := "/assets/img/go-webthings-400.png"
	if data.Element != nil && data.Element.Image != "" {
		image = data.Element.Image
	}
	return SEOData{
		Title:       pageTitle,
		Description: description,
		Canonical:   productionCanonical(data),
		Image:       absoluteSEOImage(data, image),
		SiteName:    "go-webthings",
		Type:        "website",
	}
}

func productionCanonical(data PageData) string {
	if data.Production == "" {
		return "https://gowebthings.com"
	}
	return strings.Replace(data.Production, ":element", "", 1)
}

func absoluteSEOImage(data PageData, image string) string {
	if strings.HasPrefix(image, "http://") || strings.HasPrefix(image, "https://") {
		return image
	}
	return strings.TrimRight(originFromURL(productionCanonical(data)), "/") + "/" + strings.TrimPrefix(image, "/")
}

func originFromURL(rawURL string) string {
	if scheme, rest, ok := strings.Cut(rawURL, "://"); ok {
		host, _, _ := strings.Cut(rest, "/")
		return scheme + "://" + host
	}
	return "https://gowebthings.com"
}

func docsFileSystem() fs.FS {
	if dir := os.Getenv("ASSET_DIR"); dir != "" {
		return os.DirFS(dir)
	}
	if _, err := os.Stat("deploy/website/docs/templates"); err == nil {
		return os.DirFS("deploy/website/docs")
	}
	if _, err := os.Stat("docs/templates"); err == nil {
		return os.DirFS("docs")
	}
	if _, err := os.Stat("templates"); err == nil {
		return os.DirFS(".")
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
		deployDir := filepath.Join(root, "deploy", "website", "docs")
		if _, err := os.Stat(filepath.Join(deployDir, "templates")); err == nil {
			return os.DirFS(deployDir)
		}
	}
	return os.DirFS(".")
}

func mainFileSystem() fs.FS {
	if dir := os.Getenv("MAIN_ASSET_DIR"); dir != "" {
		return os.DirFS(dir)
	}
	if _, err := os.Stat("deploy/website/main/templates"); err == nil {
		return os.DirFS("deploy/website/main")
	}
	if _, err := os.Stat("main/templates"); err == nil {
		return os.DirFS("main")
	}
	if _, err := os.Stat("templates"); err == nil {
		return os.DirFS(".")
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
		deployDir := filepath.Join(root, "deploy", "website", "main")
		if _, err := os.Stat(filepath.Join(deployDir, "templates")); err == nil {
			return os.DirFS(deployDir)
		}
	}
	return os.DirFS(".")
}

func pageFileSystem(section Section) fs.FS {
	if section == SectionMain || section == "" {
		return mainFS
	}
	return websiteFS
}

func pageTemplate(section Section) string {
	if section == SectionMain || section == "" {
		return "templates/page.gohtml"
	}
	return "templates/general/page.gohtml"
}

func sectionFromHost(host string) Section {
	host = hostWithoutPort(host)
	switch {
	case strings.HasPrefix(host, "docs."):
		return SectionDocs
	case strings.HasPrefix(host, "showcase."):
		return SectionShowcase
	default:
		return SectionMain
	}
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
	return elementsWithLinks(nil)
}

func sectionElements() []Element {
	return elementsWithLinks(func(slug string) string {
		return "/" + slug
	})
}

func mainElements(req *http.Request) []Element {
	return elementsWithLinks(func(slug string) string {
		return docsElementURL(req, slug)
	})
}

func elementsWithLinks(linkFor func(string) string) []Element {
	items := make([]Element, 0, len(elementNames))
	for _, slug := range elementNames {
		items = append(items, Element{
			Slug:        slug,
			Name:        nameFromSlug(slug),
			Description: elementDescription(slug),
			Image:       elementImage(slug),
			Link:        linkForSlug(linkFor, slug),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Slug < items[j].Slug
	})

	return items
}

func linkForSlug(linkFor func(string) string, slug string) string {
	if linkFor == nil {
		return ""
	}
	return linkFor(slug)
}

func docsElementURL(req *http.Request, slug string) string {
	return mainFamilyURL(req, "docs", slug)
}

func docsHomeURL(req *http.Request) string {
	return mainFamilyURL(req, "docs", "")
}

func mainWebsiteURL(req *http.Request) string {
	return mainFamilyURL(req, "", "")
}

func mainFamilyURL(req *http.Request, subdomain, path string) string {
	host := hostWithoutPort(req.Host)
	port := portFromHost(req.Host)
	scheme := "https"
	if strings.HasSuffix(host, "rocketweb.nl") || strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
		scheme = "http"
	}

	baseDomain := strings.TrimPrefix(host, "www.")
	baseDomain = strings.TrimPrefix(baseDomain, "docs.")
	baseDomain = strings.TrimPrefix(baseDomain, "showcase.")
	if baseDomain == "" {
		baseDomain = "gowebthings.com"
	}

	targetHost := baseDomain
	if subdomain != "" {
		targetHost = subdomain + "." + baseDomain
	}
	if port != "" {
		targetHost += ":" + port
	}

	if path != "" {
		path = "/" + strings.TrimPrefix(path, "/")
	}

	return fmt.Sprintf("%s://%s%s", scheme, targetHost, path)
}

func hostWithoutPort(host string) string {
	if h, _, ok := strings.Cut(host, ":"); ok {
		return h
	}
	return host
}

func portFromHost(host string) string {
	_, port, ok := strings.Cut(host, ":")
	if !ok {
		return ""
	}
	return port
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
	case SectionMain:
		return "Main"
	case SectionDocs:
		return "Docs"
	case SectionShowcase:
		return "Showcase"
	default:
		return "Page"
	}
}
