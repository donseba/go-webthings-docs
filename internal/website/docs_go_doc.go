package website

import (
	"html/template"
	"net/http"
	"strings"

	router "github.com/donseba/go-router"
)

type goDocsApp struct {
	docs  *docsRenderer
	pages map[string]docsPage
}

func registerGoDocsDocsRoutes(r *router.Router, domain string) {
	registerDocsPageRoutes(r, domain, "go-docs", goDocsDocs.pages, GoDocsPath, goDocsDocs.page)
}

func mustNewGoDocsDocs() *goDocsApp {
	nav := []NavItem{
		{Path: "/", Label: "Introduction", Group: "Guide"},
		{Path: "/install", Label: "Install", Group: "Guide"},
		{Path: "/contracts", Label: "Contracts", Group: "Guide"},
		{Path: "/annotations", Label: "Annotations", Group: "Guide"},
		{Path: "/generated-helpers", Label: "Generated helpers", Group: "Guide"},
		{Path: "/editor", Label: "Editor support", Group: "Guide"},
		{Path: "/renderer", Label: "Renderer", Group: "Guide"},
		{Path: "/cli", Label: "CLI and index", Group: "Reference"},
		{Path: "/lsp", Label: "LSP behavior", Group: "Reference"},
	}

	return &goDocsApp{
		docs: newDocsRenderer(docsRendererConfig{
			BasePath:  "/go-docs",
			AppName:   "go-docs",
			LogName:   "go-docs",
			Logo:      "gd",
			Title:     "go-doc",
			Subtitle:  "typed contracts for Go templates",
			GitHubURL: "https://github.com/donseba/go-doc",
			Nav:       nav,
			Funcs: []template.FuncMap{{
				"goDocsPath": GoDocsPath,
			}},
		}),
		pages: docsPages("templates/go_doc", map[string]docsPage{
			"/": {
				Template:    "overview.gohtml",
				Title:       "Typed contracts for Go templates",
				Description: "go-doc adds editor intelligence to normal html/template files.",
				Section:     "Documentation",
			},
			"/install": {
				Template:    "install.gohtml",
				Title:       "Install",
				Description: "Set up the CLI and editor integrations.",
				Section:     "Getting Started",
			},
			"/contracts": {
				Template:    "contracts.gohtml",
				Title:       "Template contracts",
				Description: "Declare the data shape once, then let the editor follow it.",
				Section:     "Core Concepts",
			},
			"/annotations": {
				Template:    "annotations.gohtml",
				Title:       "Annotations",
				Description: "Model, dot, function, and symbol annotations that describe template data.",
				Section:     "Core Concepts",
			},
			"/generated-helpers": {
				Template:    "generated_helpers.gohtml",
				Title:       "Generated helpers",
				Description: "Experimental package-like helper namespaces for normal Go templates.",
				Section:     "Core Concepts",
			},
			"/editor": {
				Template:    "editor.gohtml",
				Title:       "Editor support",
				Description: "Completion, diagnostics, hover, and navigation across supported editors.",
				Section:     "Tooling",
			},
			"/renderer": {
				Template:    "renderer.gohtml",
				Title:       "Renderer",
				Description: "A small helper for registering model values without changing template execution.",
				Section:     "Runtime",
			},
			"/cli": {
				Template:    "cli.gohtml",
				Title:       "CLI and index",
				Description: "How go-doc scans packages and produces editor metadata.",
				Section:     "Reference",
			},
			"/lsp": {
				Template:    "lsp.gohtml",
				Title:       "LSP behavior",
				Description: "What the language server understands today.",
				Section:     "Reference",
			},
		}),
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

func (app *goDocsApp) page(page docsPage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.render(w, r, page)
	}
}

func (app *goDocsApp) render(w http.ResponseWriter, r *http.Request, page docsPage) {
	app.docs.render(w, r, page, nil)
}
