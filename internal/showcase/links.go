package showcase

import (
	"fmt"
	"net/http"
	"strings"
)

func showcaseMainURL(req *http.Request) string {
	return showcaseFamilyURL(req, "", "")
}

func showcaseDocsURL(req *http.Request) string {
	return showcaseFamilyURL(req, "docs", "/go-partial")
}

func showcaseFamilyURL(req *http.Request, subdomain, path string) string {
	host, port := showcaseHostParts(req)
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

func showcaseHostParts(req *http.Request) (string, string) {
	host := ""
	if req != nil {
		host = req.Host
	}
	if host == "" {
		return "gowebthings.com", ""
	}
	if h, port, ok := strings.Cut(host, ":"); ok {
		return h, port
	}
	return host, ""
}
