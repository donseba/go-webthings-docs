package website

import (
	"fmt"
	"strings"
)

func pageSEO(data PageData) SEOData {
	pageTitle := data.Title
	if data.Section != "" && data.Section != SectionMain {
		pageTitle = fmt.Sprintf("%s - %s", data.Title, title(data.Section))
	}
	description := data.Description
	if description == "" {
		description = "Composable Go packages for server-rendered websites, docs, routing, and interactive partials."
	}
	image := "/assets/img/logo-go-webthings.png"
	if data.Element != nil && data.Element.Image != "" {
		image = data.Element.Image
	}
	return SEOData{
		Title:       pageTitle,
		Description: description,
		Canonical:   productionCanonical(data),
		Image:       absoluteSEOImage(data, image),
		SiteName:    "go-webthings",
		Type:        "website",
	}
}

func productionCanonical(data PageData) string {
	if data.Production == "" {
		return "https://gowebthings.com"
	}
	return rootlessURL(strings.Replace(data.Production, ":element", "", 1))
}

func absoluteSEOImage(data PageData, image string) string {
	if strings.HasPrefix(image, "http://") || strings.HasPrefix(image, "https://") {
		return image
	}
	return strings.TrimRight(originFromURL(productionCanonical(data)), "/") + "/" + strings.TrimPrefix(image, "/")
}

func originFromURL(rawURL string) string {
	if scheme, rest, ok := strings.Cut(rawURL, "://"); ok {
		host, _, _ := strings.Cut(rest, "/")
		return scheme + "://" + host
	}
	return "https://gowebthings.com"
}
