package website

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	router "github.com/donseba/go-router"
)

type goRouterApp struct {
	docs  *docsRenderer
	pages map[string]docsPage
}

func registerGoRouterDocsRoutes(r *router.Router, domain string) {
	for path, page := range goRouterDocs.pages {
		if path == "/" {
			continue
		}
		page := page
		r.Get(GoRouterPath(path), func(w http.ResponseWriter, req *http.Request) {
			goRouterDocs.docs.render(w, req, page, nil)
		}).As(fmt.Sprintf("%s.go-router.%s", domain, strings.TrimPrefix(path, "/")))
	}
}

func mustNewGoRouterDocs() *goRouterApp {
	nav := []NavItem{
		{Path: "/", Label: "Introduction", Group: "Guide"},
		{Path: "/routing", Label: "Routing", Group: "Guide"},
		{Path: "/groups", Label: "Groups and with", Group: "Guide"},
		{Path: "/hosts", Label: "Hosts and subdomains", Group: "Guide"},
		{Path: "/named-routes", Label: "Named routes", Group: "Guide"},
		{Path: "/mounting", Label: "Mounting", Group: "Guide"},
		{Path: "/middleware", Label: "Overview", Group: "Middleware"},
		{Path: "/middleware/request-id", Label: "Request ID", Group: "Middleware"},
		{Path: "/middleware/logger", Label: "Logger", Group: "Middleware"},
		{Path: "/middleware/recover", Label: "Recover", Group: "Middleware"},
		{Path: "/middleware/timeout", Label: "Timeout", Group: "Middleware"},
		{Path: "/middleware/real-ip", Label: "Real IP", Group: "Middleware"},
		{Path: "/middleware/security-headers", Label: "Security headers", Group: "Middleware"},
		{Path: "/middleware/cors", Label: "CORS", Group: "Middleware"},
		{Path: "/middleware/content-length", Label: "Content length", Group: "Middleware"},
		{Path: "/middleware/timer", Label: "Timer", Group: "Middleware"},
		{Path: "/params", Label: "Path params", Group: "Reference"},
		{Path: "/json-helpers", Label: "JSON helpers", Group: "Reference"},
		{Path: "/static-files", Label: "Static files", Group: "Reference"},
		{Path: "/status-handlers", Label: "Status handlers", Group: "Reference"},
		{Path: "/trailing-slashes", Label: "Trailing slashes", Group: "Reference"},
		{Path: "/diagnostics", Label: "Diagnostics", Group: "Reference"},
		{Path: "/openapi", Label: "OpenAPI", Group: "Reference"},
		{Path: "/production", Label: "Production", Group: "Reference"},
	}

	return &goRouterApp{
		docs: newDocsRenderer(docsRendererConfig{
			BasePath:  "/go-router",
			AppName:   "go-router",
			LogName:   "go-router",
			Logo:      "gr",
			Title:     "go-router",
			Subtitle:  "host-aware routing for Go websites",
			GitHubURL: "https://github.com/donseba/go-router",
			Nav:       nav,
			Funcs: []template.FuncMap{{
				"goRouterPath": GoRouterPath,
			}},
		}),
		pages: docsPages("templates/go_router", map[string]docsPage{
			"/": {
				Template:    "overview.gohtml",
				Title:       "HTTP routing for Go websites",
				Description: "go-router wraps net/http with host routing, groups, middleware, named routes, and practical docs helpers.",
				Section:     "Documentation",
			},
			"/routing": {
				Template:    "routing.gohtml",
				Title:       "Routing",
				Description: "Register method-aware routes, groups, and path parameters using standard ServeMux patterns.",
				Section:     "Guide",
			},
			"/groups": {
				Template:    "groups.gohtml",
				Title:       "Groups and with",
				Description: "Share path prefixes, middleware, and docs metadata without leaving net/http.",
				Section:     "Guide",
			},
			"/hosts": {
				Template:    "hosts.gohtml",
				Title:       "Hosts and subdomains",
				Description: "Scope handlers to exact hosts, named subdomains, and wildcard subdomains.",
				Section:     "Guide",
			},
			"/named-routes": {
				Template:    "named_routes.gohtml",
				Title:       "Named routes",
				Description: "Generate paths and full URLs from stable route names.",
				Section:     "Guide",
			},
			"/mounting": {
				Template:    "mounting.gohtml",
				Title:       "Mounting routers",
				Description: "Mount normal handlers or child routers while preserving route names, route walks, and OpenAPI metadata.",
				Section:     "Guide",
			},
			"/middleware": {
				Template:    "middleware.gohtml",
				Title:       "Middleware",
				Description: "Use standard net/http middleware globally, inside groups, or inside host routers.",
				Section:     "Middleware",
			},
			"/middleware/request-id": {
				Template:    "middleware_request_id.gohtml",
				Title:       "Request ID middleware",
				Description: "Propagate or generate an X-Request-ID value and store it on the request context.",
				Section:     "Middleware",
			},
			"/middleware/logger": {
				Template:    "middleware_logger.gohtml",
				Title:       "Logger middleware",
				Description: "Log method, URI, status, bytes written, duration, and request ID.",
				Section:     "Middleware",
			},
			"/middleware/recover": {
				Template:    "middleware_recover.gohtml",
				Title:       "Recover middleware",
				Description: "Convert panics into 500 responses while logging the recovered value.",
				Section:     "Middleware",
			},
			"/middleware/timeout": {
				Template:    "middleware_timeout.gohtml",
				Title:       "Timeout middleware",
				Description: "Wrap handlers with http.TimeoutHandler for request deadlines.",
				Section:     "Middleware",
			},
			"/middleware/real-ip": {
				Template:    "middleware_real_ip.gohtml",
				Title:       "Real IP middleware",
				Description: "Set RemoteAddr from forwarding headers, preferably only for trusted proxies.",
				Section:     "Middleware",
			},
			"/middleware/security-headers": {
				Template:    "middleware_security_headers.gohtml",
				Title:       "Security headers middleware",
				Description: "Apply practical browser security headers with optional overrides.",
				Section:     "Middleware",
			},
			"/middleware/cors": {
				Template:    "middleware_cors.gohtml",
				Title:       "CORS middleware",
				Description: "Allow configured origins, methods, headers, credentials, and preflight cache age.",
				Section:     "Middleware",
			},
			"/middleware/content-length": {
				Template:    "middleware_content_length.gohtml",
				Title:       "Content length middleware",
				Description: "Buffer normal responses so Content-Length can be set before the body is written.",
				Section:     "Middleware",
			},
			"/middleware/timer": {
				Template:    "middleware_timer.gohtml",
				Title:       "Timer middleware",
				Description: "A tiny development timer that logs duration, method, and path.",
				Section:     "Middleware",
			},
			"/params": {
				Template:    "params.gohtml",
				Title:       "Path parameters",
				Description: "Read ServeMux path values directly or through typed helper functions.",
				Section:     "Reference",
			},
			"/json-helpers": {
				Template:    "json_helpers.gohtml",
				Title:       "JSON helpers",
				Description: "Small request and response helpers for compact JSON APIs.",
				Section:     "Reference",
			},
			"/static-files": {
				Template:    "static_files.gohtml",
				Title:       "Static files",
				Description: "Serve asset directories and files, including host-scoped assets.",
				Section:     "Reference",
			},
			"/status-handlers": {
				Template:    "status_handlers.gohtml",
				Title:       "Status handlers",
				Description: "Customize router-generated 404, 405, and other status responses.",
				Section:     "Reference",
			},
			"/trailing-slashes": {
				Template:    "trailing_slashes.gohtml",
				Title:       "Trailing slashes",
				Description: "Choose whether the router redirects between slash and non-slash variants.",
				Section:     "Reference",
			},
			"/diagnostics": {
				Template:    "diagnostics.gohtml",
				Title:       "Route diagnostics",
				Description: "Walk the route tree or print a route table during startup and tests.",
				Section:     "Reference",
			},
			"/openapi": {
				Template:    "openapi.gohtml",
				Title:       "OpenAPI",
				Description: "Attach lightweight route metadata and inspect the generated document.",
				Section:     "Reference",
			},
			"/production": {
				Template:    "production.gohtml",
				Title:       "Production",
				Description: "Startup registration, middleware defaults, and deployment checks.",
				Section:     "Reference",
			},
		}),
	}
}

func GoRouterPath(path string) string {
	if path == "" || path == "/" {
		return "/go-router"
	}
	return "/go-router" + path
}
