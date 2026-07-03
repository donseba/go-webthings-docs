package website

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

func requestPath(req *http.Request) string {
	if req == nil || req.URL == nil || req.URL.Path == "" {
		return "/"
	}
	return req.URL.Path
}

func rootlessURL(rawURL string) string {
	scheme, rest, ok := strings.Cut(rawURL, "://")
	if !ok {
		return rawURL
	}
	if strings.Count(rest, "/") == 1 && strings.HasSuffix(rest, "/") {
		return scheme + "://" + strings.TrimSuffix(rest, "/")
	}
	return rawURL
}

func productionURL(req *http.Request, section Section) string {
	return fixedEnvironmentURL(req, section, "https", "gowebthings.com", "")
}

func localURL(req *http.Request, section Section) string {
	return fixedEnvironmentURL(req, section, "http", "rocketweb.nl", "8080")
}

func fixedEnvironmentURL(req *http.Request, section Section, scheme, baseDomain, port string) string {
	host := baseDomain
	if section == SectionDocs {
		host = "docs." + baseDomain
	}
	if section == SectionShowcase {
		host = "showcase." + baseDomain
	}
	if port != "" {
		host += ":" + port
	}
	return fmt.Sprintf("%s://%s%s", scheme, host, requestPath(req))
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
		return "/assets/img/logo-go-partial.png"
	case "go-docs":
		return "/assets/img/logo-go-doc.png"
	case "go-router":
		return "/assets/img/logo-go-router.png"
	default:
		return "/assets/img/logo-go-webthings.png"
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
