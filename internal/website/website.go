package website

import (
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"

	router "github.com/donseba/go-router"
)

func NewRouter() http.Handler {
	r := router.New(http.NewServeMux(), "go-webthings docs", "0.1.0")
	r.HandleStatus(http.StatusNotFound, renderNotFound)

	registerDomain(r, "rocketweb.nl")
	registerDomain(r, "gowebthings.com")

	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "https://docs.gowebthings.com/go-router", http.StatusTemporaryRedirect)
	}).As("home")

	return r
}

func registerDomain(r *router.Router, domain string) {
	r.Host(domain, func(main *router.Router) {
		registerMainRoutes(main, domain, "main")
	})

	r.Subdomain("www", domain, func(main *router.Router) {
		registerMainRoutes(main, domain, "www.main")
	})

	r.Subdomain("docs", domain, func(docs *router.Router) {
		registerStaticRoutes(docs, websiteFS)
		registerDocsRoutes(docs, domain)
	})

	r.Subdomain("showcase", domain, func(showcase *router.Router) {
		registerStaticRoutes(showcase, showcaseFS)
		registerShowcaseRoutes(showcase, domain)
	})
}

func registerMainRoutes(r *router.Router, domain, routeScope string) {
	registerStaticRoutes(r, mainFS)
	r.Get("/", mainIndex).As(fmt.Sprintf("%s.%s.index", domain, routeScope))
	r.Get("/bulletin", mainBulletin).As(fmt.Sprintf("%s.%s.bulletin", domain, routeScope))
	r.Get("/components", mainComponentsPage).As(fmt.Sprintf("%s.%s.components", domain, routeScope))
	r.Get("/generate", mainGenerate).As(fmt.Sprintf("%s.%s.generate", domain, routeScope))
	r.Get("/generate/preview", mainGeneratePreview).As(fmt.Sprintf("%s.%s.generate.preview", domain, routeScope))
	r.Get("/generate/image", mainGenerateImage).As(fmt.Sprintf("%s.%s.generate.image", domain, routeScope))
	r.Get("/{element}", mainElement).As(fmt.Sprintf("%s.%s.element", domain, routeScope))
}

func registerStaticRoutes(r *router.Router, fsys fs.FS) {
	r.ServeFiles("/assets/", http.FS(mustSubFS(fsys, "assets")))
	r.Get("/favicon.ico", serveAssetFile(fsys, "assets/img/favicon.ico"))
}

func serveAssetFile(fsys fs.FS, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		body, err := fs.ReadFile(fsys, path)
		if err != nil {
			renderNotFound(w, req)
			return
		}
		if contentType := mime.TypeByExtension(filepath.Ext(path)); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		_, _ = w.Write(body)
	}
}
