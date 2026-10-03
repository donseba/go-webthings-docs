package website

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/donseba/go-router"
)

type componentDocsApp struct {
	docs  *docsRenderer
	pages map[string]docsPage
}

func newComponentDocs() map[string]*componentDocsApp {
	return map[string]*componentDocsApp{
		"go-form":       newGoFormDocs(),
		"go-importmap":  newGoImportmapDocs(),
		"go-translator": newGoTranslatorDocs(),
	}
}

type componentDocPage struct {
	Path        string
	Label       string
	Group       string
	Template    string
	Title       string
	Description string
}

func newComponentDocsApp(slug, subtitle, templateDir string, sections []componentDocPage) *componentDocsApp {
	nav := make([]NavItem, 0, len(sections))
	pages := make(map[string]docsPage, len(sections))
	for _, section := range sections {
		nav = append(nav, NavItem{Path: section.Path, Label: section.Label, Group: section.Group})
		pages[section.Path] = docsPage{
			Template: section.Template, Title: section.Title,
			Description: section.Description, Section: section.Group,
		}
	}
	return &componentDocsApp{
		docs: newDocsRenderer(docsRendererConfig{
			BasePath: "/" + slug, AppName: slug, Title: slug, Subtitle: subtitle,
			GitHubURL: "https://github.com/donseba/" + slug, Nav: nav,
		}),
		pages: docsPages("templates/"+templateDir, pages),
	}
}

func registerComponentDocsRoutes(r *router.Router, domain string) {
	for slug, app := range componentDocs {
		for path, page := range app.pages {
			if path == "/" {
				continue
			}
			r.Get("/"+slug+path, func(w http.ResponseWriter, req *http.Request) {
				app.docs.render(w, req, page, nil)
			}).As(fmt.Sprintf("%s.%s.%s", domain, slug, strings.TrimPrefix(path, "/")))
		}
	}
}
