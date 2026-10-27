package website

import (
	"net/http"
)

// chrome is what every page of a docs subdomain shares: the header with the
// website's sections, and the rail listing the subdomain's components or
// modules.
type chrome struct {
	// section is the header link of this subdomain: "docs" or "modules".
	section   string
	title     string
	subtitle  string
	logoImage string
	gitHubURL string
	// rail lists the subdomain's items; appName is the renderer's.
	rail func(r *http.Request, appName string) DocsRailPage
}

// DocsRailPage is the first column: every component or module of the
// subdomain, by group.
type DocsRailPage struct {
	Groups []DocsRailGroup
}

// DocsRailGroup is a group in the rail; an empty Name has no heading.
type DocsRailGroup struct {
	Name  string
	Items []DocsRailItem
}

// DocsRailItem links to a component or module.
type DocsRailItem struct {
	Label string
	URL   string
	// Icon is a symbol of assets/img/icons.svg; empty shows none.
	Icon string
	// Code shows the label in monospace, for package names.
	Code   bool
	Active bool
}

var docsChrome = &chrome{
	section:   "docs",
	title:     "go-webthings docs",
	subtitle:  "html-first go components",
	logoImage: "/assets/img/logo-go-webthings.png",
	gitHubURL: "https://github.com/donseba",
	rail:      componentsRail,
}

func (c *chrome) header(r *http.Request) DocsHeaderPage {
	return DocsHeaderPage{
		Title:     c.title,
		Subtitle:  c.subtitle,
		LogoImage: c.logoImage,
		Links:     siteLinks(r, c.section),
		GitHubURL: c.gitHubURL,
	}
}

// siteLinks are the website's sections, the current one marked.
func siteLinks(r *http.Request, section string) []HeaderLink {
	return []HeaderLink{
		{Label: "webthings", URL: mainWebsiteURL(r)},
		{Label: "docs", URL: docsHomeURL(r), Active: section == "docs"},
		{Label: "modules", URL: modulesHomeURL(r), Active: section == "modules"},
		{Label: "showcase", URL: mainFamilyURL(r, "showcase", "")},
	}
}

// componentsRail lists go-webthings' guides and the components with docs.
func componentsRail(r *http.Request, appName string) DocsRailPage {
	overview := DocsRailGroup{
		Items: []DocsRailItem{
			{Label: "Overview", URL: "/", Icon: "box", Active: requestPath(r) == "/"},
		},
	}

	framework := DocsRailGroup{
		Name: "Framework",
		Items: []DocsRailItem{
			{Label: "go-webthings", URL: guidesBase, Code: true, Active: appName == "go-webthings"},
		},
	}

	components := DocsRailGroup{Name: "Components"}

	for _, element := range sectionElements() {
		components.Items = append(components.Items, DocsRailItem{
			Label:  element.Slug,
			URL:    element.Link,
			Code:   true,
			Active: element.Slug == appName,
		})
	}

	return DocsRailPage{Groups: []DocsRailGroup{overview, framework, components}}
}
