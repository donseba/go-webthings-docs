package website

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	partial "github.com/donseba/go-partial"
	"github.com/donseba/go-partial/exp/interactions"
	router "github.com/donseba/go-router"
)

type partialDocsApp struct {
	docs  *docsRenderer
	pages map[string]docsPage
}

func registerGoPartialDocsRoutes(r *router.Router, domain string) {
	for path, page := range goPartialDocs.pages {
		if path == "/" {
			continue
		}
		page := page
		r.Get(DocsPath(path), func(w http.ResponseWriter, req *http.Request) {
			goPartialDocs.docs.render(w, req, page, nil)
		}).As(fmt.Sprintf("%s.go-partial.%s", domain, strings.TrimPrefix(path, "/")))
	}
	r.Get("/go-partial/htmx", func(w http.ResponseWriter, req *http.Request) {
		goPartialDocs.docs.render(w, req, goPartialDocs.pages["/integrations/htmx"], nil)
	}).As(domain + ".go-partial.htmx.alias")
}

func mustNewGoPartialDocs() *partialDocsApp {
	docs := newDocsRenderer(docsRendererConfig{
		BasePath:  "/go-partial",
		AppName:   "go-partial",
		LogName:   "go-partial",
		Logo:      "gp",
		Title:     "go-partial",
		Subtitle:  "server-rendered partials for Go",
		GitHubURL: "https://github.com/donseba/go-partial",
		Nav:       goPartialNavItems(),
		Funcs: []template.FuncMap{
			interactions.FuncMap(),
			{
				"docsPath": DocsPath,
			},
		},
	})

	return &partialDocsApp{
		docs:  docs,
		pages: docsPages("templates/go_partial", goPartialPages()),
	}
}

func DocsPath(path string) string {
	if path == "" || path == "/" {
		return "/go-partial"
	}
	return "/go-partial" + path
}

func configureInteractions(content *partial.Partial) {
	content.SetContract("interaction",
		interactions.NewPoll("/notifications").As("Notifications").Every(10*time.Second),
		interactions.NewOn("cart:changed", "/cart/summary").As("CartChanged").Target("#cart"),
		interactions.NewRefresh("/cart/summary").As("CartRefresh").Target("#cart").Swap(interactions.SwapOuterHTML),
	)
}

func goPartialPages() map[string]docsPage {
	return map[string]docsPage{
		"/":                            {Template: "overview.gohtml", Title: "Server-rendered partials that stay useful with HTMX", Description: "go-partial is a small rendering layer for Go applications that want reusable template regions, targeted updates, out-of-band swaps, and predictable server-side behavior.", Section: "Documentation"},
		"/installation":                {Template: "installation.gohtml", Title: "Installation", Description: "Install the package, choose a connector, and point go-partial at your template filesystem.", Section: "Guide"},
		"/rendering":                   {Template: "rendering.gohtml", Title: "Rendering model", Description: "A page is a partial tree. A wrapper partial can render a content child, registered regions keep stable IDs, and HTMX can request one target without losing the surrounding model.", Section: "Guide"},
		"/data-context":                {Template: "data_context.gohtml", Title: "Data and context", Description: "Prefer typed app models. go-partial adds request-aware helpers around normal Go templates; it should not become a second application state container.", Section: "Guide"},
		"/selection-action":            {Template: "selection_action.gohtml", Title: "Selection and action", Description: "Selection and action both read request intent from the configured connector, but they solve different problems. Selection chooses one registered partial. Action lets Go code decide what to render.", Section: "Guide"},
		"/interactions":                {Template: "interactions.gohtml", Title: "Interaction helpers", Description: "Interaction helpers describe how a server-rendered partial should reach the browser. They are delivery hints; your page data can stay on dot.", Section: "Guide", Configure: configureInteractions},
		"/deferred":                    {Template: "deferred.gohtml", Title: "Deferred partials", Description: "async renders a connector-aware placeholder that loads another endpoint after the current page has rendered. It is useful for slow panels, optional sections, and row fragments that should not block the first response.", Section: "Guide"},
		"/flash":                       {Template: "flash.gohtml", Title: "Flash messages", Description: "exp/flash renders request-scoped messages for SSR and HTMX responses. It stays deliberately small: your app owns persistence, while the package owns message shape, helpers, and default markup.", Section: "exp"},
		"/flow":                        {Template: "flow.gohtml", Title: "Page flows", Description: "PageFlow is a small coordinator for multi-step server-rendered screens. It knows the ordered steps, current step, validation callbacks, and collected step data; your application still owns routing and storage.", Section: "Guide"},
		"/localization":                {Template: "localization.gohtml", Title: "Localization", Description: "Localization is intentionally split in two: localizer carries the active locale, while translation functions such as tl, tn, ctl, and ctn are user-provided template functions.", Section: "Guide"},
		"/integrations":                {Template: "integrations.gohtml", Title: "Integration overview", Description: "go-partial sits on the server side. Integrations decide how a browser request names a target, selection, or action, and how server response intent is written back to the client.", Section: "Integration"},
		"/integrations/htmx":           {Template: "htmx.gohtml", Title: "HTMX integration", Description: "The HTMX connector reads request headers and lets go-partial render the requested target instead of the full page.", Section: "Integration"},
		"/integrations/sse":            {Template: "sse.gohtml", Title: "Server-sent events", Description: "SSE is a streaming writer layer. It does not replace connectors; it sends events after your handler decides what changed.", Section: "Integration"},
		"/integrations/custom-clients": {Template: "custom_clients.gohtml", Title: "Custom clients", Description: "Use the neutral connector when your frontend is plain fetch, a small controller, tests, or another client that can send predictable headers.", Section: "Integration"},
		"/api":                         {Template: "api.gohtml", Title: "Core API", Description: "The public surface is small: build partials, register relationships, choose a connector, then render through request-aware writer APIs.", Section: "Reference"},
		"/api/pageflow":                {Template: "pageflow_api.gohtml", Title: "PageFlow API", Description: "PageFlow defines the flow. pageflow.SessionData carries per-user state. That split lets a single flow definition be reused safely across requests, tabs, and users.", Section: "Reference"},
		"/target-resolver":             {Template: "target_resolver.gohtml", Title: "Target resolver", Description: "target.WithResolver handles DOM targets that are not fixed partial IDs. It is the tool for tables, feeds, cards, and repeated rows.", Section: "Guide"},
		"/connectors":                  {Template: "connectors.gohtml", Title: "Connectors", Description: "Connectors let go-partial understand request headers from a frontend library or a custom fetch client.", Section: "Reference"},
		"/template-functions":          {Template: "template_functions.gohtml", Title: "Template functions", Description: "go-partial adds rendering helpers, connector helpers, URL helpers, OOB helpers, and small utility functions to normal html/template files.", Section: "Reference"},
		"/error-boundaries":            {Template: "error_boundaries.gohtml", Title: "Error boundaries", Description: "A failing registered partial renders a section-level fallback. The page shell survives, which is useful for development and production.", Section: "Guide"},
		"/observability":               {Template: "observability.gohtml", Title: "Observability", Description: "go-partial emits small diagnostic events, but it does not own logging, metrics storage, tracing providers, queues, or exporters. Applications attach sinks and decide where events go.", Section: "ext and exp"},
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
