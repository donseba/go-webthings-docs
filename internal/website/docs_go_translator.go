package website

func newGoTranslatorDocs() *componentDocsApp {
	return newComponentDocsApp("go-translator", "gettext for Go", "go_translator", []componentDocPage{
		{Path: "/", Label: "Overview", Group: "Guide", Template: "overview.gohtml", Title: "Translate Go templates with gettext", Description: "PO catalogues and request-local translations."},
		{Path: "/installation", Label: "Installation", Group: "Guide", Template: "installation.gohtml", Title: "Installation and first render", Description: "Load a catalogue, then render a translated template."},
		{Path: "/catalogues", Label: "Catalogues", Group: "Guide", Template: "catalogues.gohtml", Title: "PO catalogues and languages", Description: "Keep one PO file per locale and one shared POT file."},
		{Path: "/locales", Label: "Request-local locales", Group: "Guide", Template: "locales.gohtml", Title: "Choosing a locale per request", Description: "Select a loaded language without changing shared state."},
		{Path: "/templates", Label: "Template helpers", Group: "Guide", Template: "templates.gohtml", Title: "Template helpers and formatting", Description: "Translate literal keys and interpolate typed values."},
		{Path: "/plurals", Label: "Plural forms", Group: "Guide", Template: "plurals.gohtml", Title: "Plural forms and count arguments", Description: "Let the catalogue select the language\u2019s plural form."},
		{Path: "/contexts", Label: "Contexts and prefixes", Group: "Guide", Template: "contexts.gohtml", Title: "Translation contexts and prefixes", Description: "Disambiguate the same text used in different places."},
		{Path: "/extraction", Label: "Key extraction", Group: "Guide", Template: "extraction.gohtml", Title: "Extracting template keys", Description: "Maintain a POT file from literal template calls."},
		{Path: "/integrations", Label: "Component integration", Group: "Integration", Template: "integrations.gohtml", Title: "Using go-partial and go-form", Description: "Share a request localizer through small adapters."},
		{Path: "/editing", Label: "Catalogue editing", Group: "Reference", Template: "editing.gohtml", Title: "Creating and editing catalogues", Description: "Edit loaded catalogue entries and write them explicitly."},
		{Path: "/api", Label: "API reference", Group: "Reference", Template: "api.gohtml", Title: "go-translator API", Description: "Construct, look up, extract, and maintain catalogues."},
	})
}
