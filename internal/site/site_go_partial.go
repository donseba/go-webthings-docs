package site

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	partial "github.com/donseba/go-partial"
	"github.com/donseba/go-partial/connector"
	"github.com/donseba/go-partial/exp/interactions"
	"github.com/donseba/go-partial/exp/templatehelpers"
	exterrors "github.com/donseba/go-partial/ext/errors"
	router "github.com/donseba/go-router"
)

type partialDocsApp struct {
	root *partial.Partial
}

type DocsPage struct{}

type DocsHeaderPage struct{}

type DocsNavPage struct {
	Nav    []NavItem
	Groups []string
}

type DocsShellPage struct {
	AppName string
	Header  DocsHeaderPage
	Sidebar DocsNavPage
}

func registerGoPartialDocsRoutes(r *router.Router, domain string) {
	r.Get("/go-partial/installation", goPartialDocs.page("templates/docs_installation.gohtml")).As(fmt.Sprintf("%s.go-partial.installation", domain))
	r.Get("/go-partial/rendering", goPartialDocs.page("templates/docs_rendering.gohtml")).As(fmt.Sprintf("%s.go-partial.rendering", domain))
	r.Get("/go-partial/data-context", goPartialDocs.page("templates/docs_data_context.gohtml")).As(fmt.Sprintf("%s.go-partial.data-context", domain))
	r.Get("/go-partial/selection-action", goPartialDocs.page("templates/docs_selection_action.gohtml")).As(fmt.Sprintf("%s.go-partial.selection-action", domain))
	r.Get("/go-partial/interactions", goPartialDocs.interactions).As(fmt.Sprintf("%s.go-partial.interactions", domain))
	r.Get("/go-partial/deferred", goPartialDocs.page("templates/docs_deferred.gohtml")).As(fmt.Sprintf("%s.go-partial.deferred", domain))
	r.Get("/go-partial/flash", goPartialDocs.page("templates/docs_flash.gohtml")).As(fmt.Sprintf("%s.go-partial.flash", domain))
	r.Get("/go-partial/flow", goPartialDocs.page("templates/docs_flow.gohtml")).As(fmt.Sprintf("%s.go-partial.flow", domain))
	r.Get("/go-partial/localization", goPartialDocs.page("templates/docs_localization.gohtml")).As(fmt.Sprintf("%s.go-partial.localization", domain))
	r.Get("/go-partial/integrations", goPartialDocs.page("templates/docs_integrations.gohtml")).As(fmt.Sprintf("%s.go-partial.integrations", domain))
	r.Get("/go-partial/integrations/htmx", goPartialDocs.page("templates/docs_htmx.gohtml")).As(fmt.Sprintf("%s.go-partial.htmx", domain))
	r.Get("/go-partial/integrations/sse", goPartialDocs.page("templates/docs_sse.gohtml")).As(fmt.Sprintf("%s.go-partial.sse", domain))
	r.Get("/go-partial/integrations/custom-clients", goPartialDocs.page("templates/docs_custom_clients.gohtml")).As(fmt.Sprintf("%s.go-partial.custom-clients", domain))
	r.Get("/go-partial/api", goPartialDocs.page("templates/docs_api.gohtml")).As(fmt.Sprintf("%s.go-partial.api", domain))
	r.Get("/go-partial/api/pageflow", goPartialDocs.page("templates/docs_pageflow_api.gohtml")).As(fmt.Sprintf("%s.go-partial.pageflow-api", domain))
	r.Get("/go-partial/target-resolver", goPartialDocs.page("templates/docs_target_resolver.gohtml")).As(fmt.Sprintf("%s.go-partial.target-resolver", domain))
	r.Get("/go-partial/connectors", goPartialDocs.page("templates/docs_connectors.gohtml")).As(fmt.Sprintf("%s.go-partial.connectors", domain))
	r.Get("/go-partial/template-functions", goPartialDocs.page("templates/docs_template_functions.gohtml")).As(fmt.Sprintf("%s.go-partial.template-functions", domain))
	r.Get("/go-partial/htmx", goPartialDocs.page("templates/docs_htmx.gohtml")).As(fmt.Sprintf("%s.go-partial.htmx.alias", domain))
	r.Get("/go-partial/error-boundaries", goPartialDocs.page("templates/docs_error_boundaries.gohtml")).As(fmt.Sprintf("%s.go-partial.error-boundaries", domain))
	r.Get("/go-partial/observability", goPartialDocs.page("templates/docs_observability.gohtml")).As(fmt.Sprintf("%s.go-partial.observability", domain))
}

func mustNewGoPartialDocs() *partialDocsApp {
	docsFS := mustSubFS(siteFS, "elements/go_partial")
	root := partial.NewID("shell", "templates/shell.gohtml").
		SetConnector(connector.NewHTMX(nil)).
		SetFileSystem(docsFS).
		SetBasePath("/go-partial").
		UseTemplateCache(true).
		Use(exterrors.Stage(exterrors.WithMode(exterrors.ModeDetailed))).
		SetFunc(interactions.FuncMap(), templatehelpers.FuncMap(), template.FuncMap{
			"docsPath": DocsPath,
		})

	return &partialDocsApp{root: root}
}

func DocsPath(path string) string {
	if path == "" || path == "/" {
		return "/go-partial"
	}
	return "/go-partial" + path
}

func (app *partialDocsApp) overview(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSuffix(r.URL.Path, "/") != "/go-partial" {
		renderNotFound(w, r)
		return
	}
	app.render(w, r, "templates/docs_overview.gohtml")
}

func (app *partialDocsApp) page(tmpl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.render(w, r, tmpl)
	}
}

func (app *partialDocsApp) interactions(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "templates/docs_interactions.gohtml", func(content *partial.Partial) {
		content.SetContract("interaction",
			interactions.NewPoll("/notifications").As("Notifications").Every(10*time.Second),
			interactions.NewOn("cart:changed", "/cart/summary").As("CartChanged").Target("#cart"),
			interactions.NewRefresh("/cart/summary").As("CartRefresh").Target("#cart").Swap(interactions.SwapOuterHTML),
		)
	})
}

func (app *partialDocsApp) render(w http.ResponseWriter, r *http.Request, tmpl string, configure ...func(*partial.Partial)) {
	content := partial.NewID("content", tmpl).SetDot(DocsPage{})
	for _, fn := range configure {
		if fn != nil {
			fn(content)
		}
	}

	header := DocsHeaderPage{}
	nav := goPartialNavItems()
	sidebar := DocsNavPage{Nav: nav, Groups: navGroups(nav)}
	root := app.root.Clone().SetDot(DocsShellPage{
		AppName: "go-partial",
		Header:  header,
		Sidebar: sidebar,
	})
	root.WithOOB(partial.NewID("header", "templates/header.gohtml").SetDot(header).SetAlwaysSwapOOB(true))
	root.WithOOB(partial.NewID("sidebar", "templates/sidebar.gohtml").SetDot(sidebar).SetAlwaysSwapOOB(true))
	root.SetContent(content)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := partial.Write(r.Context(), w, r, root); err != nil {
		log.Printf("render go-partial docs error: %v", err)
	}
}

func goPartialNavItems() []NavItem {
	return []NavItem{
		{Path: "/", Label: "Overview", Group: "Guide"},
		{Path: "/installation", Label: "Installation", Group: "Guide"},
		{Path: "/rendering", Label: "Rendering model", Group: "Guide"},
		{Path: "/data-context", Label: "Data and context", Group: "Guide"},
		{Path: "/deferred", Label: "Deferred partials", Group: "Guide"},
		{Path: "/error-boundaries", Label: "Error boundaries", Group: "ext"},
		{Path: "/observability", Label: "Observability", Group: "ext"},
		{Path: "/flash", Label: "Flash messages", Group: "exp"},
		{Path: "/selection-action", Label: "Selection and action", Group: "exp"},
		{Path: "/interactions", Label: "Interaction helpers", Group: "exp"},
		{Path: "/flow", Label: "Page flows", Group: "exp"},
		{Path: "/target-resolver", Label: "Target resolver", Group: "exp"},
		{Path: "/localization", Label: "Localization", Group: "exp"},
		{Path: "/integrations/sse", Label: "Server-sent events", Group: "exp"},
		{Path: "/api/pageflow", Label: "PageFlow API", Group: "exp"},
		{Path: "/integrations", Label: "Overview", Group: "Integration"},
		{Path: "/integrations/htmx", Label: "HTMX", Group: "Integration"},
		{Path: "/integrations/custom-clients", Label: "Custom clients", Group: "Integration"},
		{Path: "/api", Label: "Core API", Group: "Reference"},
		{Path: "/template-functions", Label: "Template functions", Group: "Reference"},
		{Path: "/connectors", Label: "Connectors", Group: "Reference"},
	}
}

func navGroups(items []NavItem) []string {
	seen := make(map[string]struct{}, len(items))
	groups := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item.Group]; ok {
			continue
		}
		seen[item.Group] = struct{}{}
		groups = append(groups, item.Group)
	}
	return groups
}
