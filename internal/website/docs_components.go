package website

type componentDocsApp struct {
	docs  *docsRenderer
	pages map[string]docsPage
}

func newComponentDocs() map[string]*componentDocsApp {
	components := []struct {
		slug        string
		templateDir string
		title       string
		description string
	}{
		{"go-form", "go_form", "HTML forms from Go structs", "Render, map, and validate typed forms with request metadata, CSRF protection, and configurable themes."},
		{"go-importmap", "go_importmap", "Prepare JavaScript and CSS assets", "Fetch pinned CDN packages, reuse a local cache, and generate import maps and stylesheet tags for Go websites."},
		{"go-translator", "go_translator", "Translate Go templates with gettext", "Use PO catalogues, request-local locales, plural forms, translation contexts, and template key extraction."},
	}
	apps := make(map[string]*componentDocsApp, len(components))
	for _, component := range components {
		apps[component.slug] = &componentDocsApp{
			docs: newDocsRenderer(docsRendererConfig{
				BasePath:  "/" + component.slug,
				AppName:   component.slug,
				Title:     component.slug,
				Subtitle:  component.description,
				GitHubURL: "https://github.com/donseba/" + component.slug,
				Nav:       []NavItem{{Path: "/", Label: "Overview and usage", Group: "Documentation"}},
			}),
			pages: docsPages("templates/"+component.templateDir, map[string]docsPage{
				"/": {Template: "overview.gohtml", Title: component.title, Description: component.description, Section: "Documentation"},
			}),
		}
	}
	return apps
}
