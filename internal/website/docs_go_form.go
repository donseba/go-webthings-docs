package website

func newGoFormDocs() *componentDocsApp {
	return newComponentDocsApp("go-form", "typed HTML forms", "go_form", []componentDocPage{
		{Path: "/", Label: "Overview", Group: "Guide", Template: "overview.gohtml", Title: "HTML forms from Go structs", Description: "Typed forms, validation, and CSRF protection."},
		{Path: "/installation", Label: "Installation", Group: "Guide", Template: "installation.gohtml", Title: "Installation and first render", Description: "Render an ordinary Go struct as an HTML form."},
		{Path: "/fields", Label: "Fields and choices", Group: "Guide", Template: "fields.gohtml", Title: "Fields, tags, and choices", Description: "Choose controls and give submitted fields stable names."},
		{Path: "/metadata", Label: "Form metadata", Group: "Guide", Template: "metadata.gohtml", Title: "Form metadata and request tokens", Description: "Keep actions and CSRF values outside the input model."},
		{Path: "/submissions", Label: "Handling submissions", Group: "Guide", Template: "submissions.gohtml", Title: "Mapping and handling submissions", Description: "Parse input, validate it, then call your application service."},
		{Path: "/validation", Label: "Validation", Group: "Guide", Template: "validation.gohtml", Title: "Built-in and custom validation", Description: "Return field errors without coupling them to persistence."},
		{Path: "/csrf", Label: "CSRF protection", Group: "Guide", Template: "csrf.gohtml", Title: "CSRF protection and shared stores", Description: "Protect a form submission and keep other open forms usable."},
		{Path: "/themes", Label: "Themes", Group: "Guide", Template: "themes.gohtml", Title: "Themes and custom templates", Description: "Use a built-in theme or supply your own field markup."},
		{Path: "/translations", Label: "Translations", Group: "Guide", Template: "translations.gohtml", Title: "Localized forms and validation", Description: "Translate labels and built-in errors for each request."},
		{Path: "/integrations", Label: "Component integration", Group: "Integration", Template: "integrations.gohtml", Title: "Forms with go-partial and HTMX", Description: "Render a full page or refreshed form through the same pipeline."},
		{Path: "/api", Label: "API reference", Group: "Reference", Template: "api.gohtml", Title: "go-form API", Description: "Renderer, metadata, mapping, validation, and token helpers."},
	})
}
