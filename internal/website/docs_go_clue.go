package website

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	router "github.com/donseba/go-router"
)

type goClueApp struct {
	docs  *docsRenderer
	pages map[string]docsPage
}

func registerGoClueDocsRoutes(r *router.Router, domain string) {
	r.Get("/go-docs", redirectGoClueDocs).As(fmt.Sprintf("%s.go-clue.legacy", domain))
	r.Get("/go-docs/{page...}", redirectGoClueDocs).As(fmt.Sprintf("%s.go-clue.legacy.page", domain))
	for path, page := range goClueDocs.pages {
		if path == "/" {
			continue
		}
		page := page
		r.Get(GoCluePath(path), func(w http.ResponseWriter, req *http.Request) {
			goClueDocs.docs.render(w, req, page, nil)
		}).As(fmt.Sprintf("%s.go-clue.%s", domain, strings.TrimPrefix(path, "/")))
	}
}

func redirectGoClueDocs(w http.ResponseWriter, req *http.Request) {
	target := GoCluePath(strings.TrimPrefix(req.URL.EscapedPath(), "/go-docs"))
	if req.URL.RawQuery != "" {
		target += "?" + req.URL.RawQuery
	}
	http.Redirect(w, req, target, http.StatusMovedPermanently)
}

func mustNewGoClueDocs() *goClueApp {
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

	return &goClueApp{
		docs: newDocsRenderer(docsRendererConfig{
			BasePath:  "/go-clue",
			AppName:   "go-clue",
			LogName:   "go-clue",
			Logo:      "gc",
			Title:     "go-clue",
			Subtitle:  "typed contracts for Go templates",
			GitHubURL: "https://github.com/donseba/go-clue",
			Nav:       nav,
			Funcs: []template.FuncMap{{
				"goCluePath": GoCluePath,
			}},
		}),
		pages: docsPages("templates/go_clue", map[string]docsPage{
			"/": {
				Template:    "overview.gohtml",
				Title:       "Typed contracts for Go templates",
				Description: "go-clue adds editor intelligence to normal html/template files.",
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
				Description: "How go-clue scans packages and produces editor metadata.",
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

func GoCluePath(path string) string {
	if path == "" || path == "/" {
		return "/go-clue"
	}
	return "/go-clue" + path
}
