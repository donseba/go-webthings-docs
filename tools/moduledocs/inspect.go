package main

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"

	"donseba/go-webthings-docs/internal/moduledoc"

	webthings "github.com/donseba/go-webthings"
	"github.com/donseba/go-webthings/admin"
	"github.com/donseba/go-webthings/api"
	"github.com/donseba/go-webthings/frontend"

	"github.com/donseba/go-router"
)

// captured is what a module's Setup returned.
type captured struct {
	parts  moduledoc.Parts
	admin  []moduledoc.AdminSection
	routes []moduledoc.Route
}

// inspect starts an app with every module on a throwaway SQLite database
// and keeps what each Setup returns.
func inspect(ctx context.Context, list []entry) (map[string]captured, error) {
	dir, err := os.MkdirTemp("", "moduledocs-")
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = os.RemoveAll(dir)
	}()

	found := map[string]captured{}
	modules := make([]webthings.Module, 0, len(list))

	for _, e := range list {
		module := e.module
		setup := module.Setup

		module.Setup = func(env webthings.Env) (webthings.Parts, error) {
			parts, err := setup(env)
			if err != nil {
				return parts, err
			}

			found[module.Key] = capture(env, parts)

			return parts, nil
		}

		modules = append(modules, module)
	}

	app := webthings.New(webthings.Config{
		Name:             "moduledocs",
		StateDir:         dir,
		Insecure:         true,
		CacheTTL:         -1,
		ScheduleInterval: -1,
	})

	if err := app.Use(modules...); err != nil {
		return nil, err
	}

	if err := app.Start(ctx); err != nil {
		return nil, err
	}

	if err := app.Close(); err != nil {
		return nil, err
	}

	for _, e := range list {
		if _, ok := found[e.module.Key]; !ok {
			return nil, fmt.Errorf("%s: Setup did not run", e.slug)
		}
	}

	return found, nil
}

func capture(env webthings.Env, parts webthings.Parts) captured {
	pages := routerRoutes(parts.Public, "page")
	feeds := feedRoutes(parts.Feeds)
	endpoints := apiRoutes(env.Key, parts.API)
	actions := actionRoutes(env)

	c := captured{
		parts: moduledoc.Parts{
			Public:     parts.Public != nil,
			Actions:    len(actions) > 0,
			Feeds:      len(feeds) > 0,
			Links:      parts.Links != nil,
			API:        len(endpoints) > 0,
			Middleware: parts.Middleware != nil,
			Search:     parts.Search != nil,
			Sitemap:    parts.Sitemap != nil,
			Scheduled:  len(parts.Scheduled) > 0,
			Close:      parts.Close != nil,
		},
		routes: sortRoutes(slices.Concat(pages, feeds, endpoints, actions)),
	}

	if parts.Service != nil {
		c.parts.Service = reflect.TypeOf(parts.Service).String()
	}

	for _, section := range parts.Admin {
		c.admin = append(c.admin, moduledoc.AdminSection{
			ID:         section.ID,
			Title:      section.Title,
			Path:       adminPath(section),
			Permission: section.PermissionID(),
			Icon:       section.Icon,
			Hidden:     section.Hidden,
			AdminOnly:  section.AdminOnly,
			Badge:      section.Badge != nil,
			Group:      section.Group,
			NewLabel:   section.NewLabel,
			Recent:     section.Recent != nil,
			Scheduled:  section.Scheduled != nil,
			Trash:      section.Trash != nil,
			Review:     section.Review != nil,
		})
	}

	return c
}

// actionRoutes are the routes of the module's action router (Env.Actions),
// which lists them relative to /_/<key>, and named without the key the app
// puts in front.
func actionRoutes(env webthings.Env) []moduledoc.Route {
	routes := routerRoutes(env.Actions(), "action")

	for i := range routes {
		routes[i].Path = "/_/" + env.Key + routes[i].Path

		if routes[i].Name != "" {
			routes[i].Name = env.Key + "." + routes[i].Name
		}
	}

	return routes
}

// routeKinds is the order of the kinds of route in the list.
var routeKinds = map[string]int{"page": 0, "feed": 1, "api": 2, "action": 3}

// sortRoutes lists the page routes first, then the feeds, the API endpoints
// and the actions, each kind by path.
func sortRoutes(routes []moduledoc.Route) []moduledoc.Route {
	sort.SliceStable(routes, func(i, j int) bool {
		if routes[i].Kind != routes[j].Kind {
			return routeKinds[routes[i].Kind] < routeKinds[routes[j].Kind]
		}

		return routes[i].Path < routes[j].Path
	})

	return routes
}

// apiRoutes are the module's endpoints of the content API (Parts.API), which
// the app mounts at /api/v1/<key>/ with "api.<key>." in front of their names.
func apiRoutes(key string, endpoints *router.Router) []moduledoc.Route {
	routes := routerRoutes(endpoints, "api")

	for i := range routes {
		routes[i].Path = api.Path + "/" + key + routes[i].Path

		if routes[i].Name != "" {
			routes[i].Name = "api." + key + "." + routes[i].Name
		}
	}

	return routes
}

// feedRoutes are the routes of the module's feeds (Parts.Feeds), which the
// app adds to its page router after Setup: <Name>.atom and <Name>.rss at
// <Path>.atom and <Path>.rss below the page, with the defaults the app uses.
func feedRoutes(feeds []frontend.Feed) []moduledoc.Route {
	var found []moduledoc.Route

	for _, f := range feeds {
		name := cmp.Or(f.Name, "feed")
		path := cmp.Or(f.Path, "/feed")

		for _, format := range []string{"atom", "rss"} {
			found = append(found, moduledoc.Route{
				Kind:   "feed",
				Name:   name + "." + format,
				Method: http.MethodGet,
				Path:   path + "." + format,
			})
		}
	}

	return found
}

// routerRoutes are the routes of handler when it is a go-router; other
// handlers keep their routes to themselves.
func routerRoutes(handler http.Handler, kind string) []moduledoc.Route {
	routes, ok := handler.(*router.Router)
	if !ok || routes == nil {
		return nil
	}

	var found []moduledoc.Route

	_ = routes.Walk(func(route router.RouteInfo) error {
		found = append(found, moduledoc.Route{
			Kind:   kind,
			Name:   route.Name,
			Method: route.Method,
			Path:   display(route.Pattern),
		})

		return nil
	})

	return found
}

// display writes a pattern's path as readers expect it: "/{$}" is "/".
func display(path string) string {
	return strings.ReplaceAll(path, "{$}", "")
}

// adminPath is a section's full address under the default admin path:
// sections give theirs relative to the admin, and an empty one is "/"+ID.
func adminPath(section admin.Section) string {
	path := section.Path
	if path == "" {
		path = "/" + section.ID
	}

	if strings.HasPrefix(path, admin.DefaultPath+"/") {
		return path
	}

	return admin.DefaultPath + path
}
