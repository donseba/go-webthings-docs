package website

import (
	fullshowcase "donseba/go-webthings-docs/internal/showcase"

	router "github.com/donseba/go-router"
)

func registerShowcaseRoutes(r *router.Router, _ string) {
	r.Mount("/", fullshowcase.NewHandler(showcaseFS))
}
