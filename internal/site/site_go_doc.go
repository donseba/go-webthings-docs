package site

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"

	router "github.com/donseba/go-router"
)

type goDocsApp struct {
	fs    fs.FS
	nav   []NavItem
	pages map[string]goDocsPage
}

type goDocsPage struct {
	Template    string
	Title       string
	Description string
	Section     string
}

type GoDocsPageData struct {
	Title       string
	Description string
	Section     string
	AppName     string
	Path        string
	BasePath    string
	Nav         []NavItem
}

func registerGoDocsDocsRoutes(r *router.Router, domain string) {
	for path, page := range goDocsDocs.pages {
		if path == "/" {
			continue
		}
		routePath := GoDocsPath(path)
		r.Get(routePath, goDocsDocs.page(page)).As(fmt.Sprintf("%s.go-docs.%s", domain, strings.TrimPrefix(path, "/")))
	}
}

func mustNewGoDocsDocs() *goDocsApp {
	return &goDocsApp{
		fs: mustSubFS(siteFS, "elements/go_docs"),
		nav: []NavItem{
			{Path: "/", Label: "Introduction", Group: "Guide"},
			{Path: "/install", Label: "Install", Group: "Guide"},
			{Path: "/contracts", Label: "Contracts", Group: "Guide"},
			{Path: "/annotations", Label: "Annotations", Group: "Guide"},
			{Path: "/generated-helpers", Label: "Generated helpers", Group: "Guide"},
			{Path: "/editor", Label: "Editor support", Group: "Guide"},
			{Path: "/renderer", Label: "Renderer", Group: "Guide"},
			{Path: "/cli", Label: "CLI and index", Group: "Reference"},
			{Path: "/lsp", Label: "LSP behavior", Group: "Reference"},
		},
		pages: map[string]goDocsPage{
			"/": {
				Template:    "templates/overview.gohtml",
				Title:       "Typed contracts for Go templates",
				Description: "go-doc adds editor intelligence to normal html/template files.",
				Section:     "Documentation",
			},
			"/install": {
				Template:    "templates/install.gohtml",
				Title:       "Install",
				Description: "Set up the CLI and editor integrations.",
				Section:     "Getting Started",
			},
			"/contracts": {
				Template:    "templates/contracts.gohtml",
				Title:       "Template contracts",
				Description: "Declare the data shape once, then let the editor follow it.",
				Section:     "Core Concepts",
			},
			"/annotations": {
				Template:    "templates/annotations.gohtml",
				Title:       "Annotations",
				Description: "Model, dot, function, and symbol annotations that describe template data.",
				Section:     "Core Concepts",
			},
			"/generated-helpers": {
				Template:    "templates/generated_helpers.gohtml",
				Title:       "Generated helpers",
				Description: "Experimental package-like helper namespaces for normal Go templates.",
				Section:     "Core Concepts",
			},
			"/editor": {
				Template:    "templates/editor.gohtml",
				Title:       "Editor support",
				Description: "Completion, diagnostics, hover, and navigation across supported editors.",
				Section:     "Tooling",
			},
			"/renderer": {
				Template:    "templates/renderer.gohtml",
				Title:       "Renderer",
				Description: "A small helper for registering model values without changing template execution.",
				Section:     "Runtime",
			},
			"/cli": {
				Template:    "templates/cli.gohtml",
				Title:       "CLI and index",
				Description: "How go-doc scans packages and produces editor metadata.",
				Section:     "Reference",
			},
			"/lsp": {
				Template:    "templates/lsp.gohtml",
				Title:       "LSP behavior",
				Description: "What the language server understands today.",
				Section:     "Reference",
			},
		},
	}
}

func GoDocsPath(path string) string {
	if path == "" || path == "/" {
		return "/go-docs"
	}
	return "/go-docs" + path
}

func (app *goDocsApp) overview(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSuffix(r.URL.Path, "/") != "/go-docs" {
		renderNotFound(w, r)
		return
	}
	app.render(w, r, app.pages["/"])
}

func (app *goDocsApp) page(page goDocsPage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.render(w, r, page)
	}
}

func (app *goDocsApp) render(w http.ResponseWriter, r *http.Request, page goDocsPage) {
	tmpl, err := template.New("layout.gohtml").Funcs(template.FuncMap{
		"isActive": func(path string) bool {
			return r.URL.Path == path
		},
		"goDocsPath": GoDocsPath,
	}).ParseFS(
		app.fs,
		"templates/layout.gohtml",
		"templates/sidebar.gohtml",
		"templates/header.gohtml",
		page.Template,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout.gohtml", GoDocsPageData{
		Title:       page.Title,
		Description: page.Description,
		Section:     page.Section,
		AppName:     "go-docs",
		Path:        r.URL.Path,
		BasePath:    "/go-docs",
		Nav:         app.nav,
	}); err != nil {
		log.Printf("render go-docs docs error: %v", err)
	}
}
