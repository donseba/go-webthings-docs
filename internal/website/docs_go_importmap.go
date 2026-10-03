package website

func newGoImportmapDocs() *componentDocsApp {
	return newComponentDocsApp("go-importmap", "JavaScript and CSS", "go_importmap", []componentDocPage{
		{Path: "/", Label: "Overview", Group: "Guide", Template: "overview.gohtml", Title: "Prepare JavaScript and CSS assets", Description: "Pinned dependencies served by your Go application."},
		{Path: "/installation", Label: "Installation", Group: "Guide", Template: "installation.gohtml", Title: "Installation and first preparation", Description: "Prepare a pinned module and inspect the generated head."},
		{Path: "/packages", Label: "Packages and files", Group: "Guide", Template: "packages.gohtml", Title: "Packages, versions, and aliases", Description: "Select the browser files you want to publish."},
		{Path: "/providers", Label: "Providers", Group: "Guide", Template: "providers.gohtml", Title: "Choosing an asset provider", Description: "Use a default provider or override individual packages."},
		{Path: "/storage", Label: "Storage and caching", Group: "Guide", Template: "storage.gohtml", Title: "Filesystem roots and the cache", Description: "Separate on-disk paths from public browser URLs."},
		{Path: "/rendering", Label: "Maps and styles", Group: "Guide", Template: "rendering.gohtml", Title: "Rendering import maps and styles", Description: "Include dependency markup before importing modules."},
		{Path: "/serving", Label: "Serving and importing", Group: "Guide", Template: "serving.gohtml", Title: "Serving assets and importing modules", Description: "Connect the prepared filesystem to the browser URLs."},
		{Path: "/deployment", Label: "Build and deployment", Group: "Guide", Template: "deployment.gohtml", Title: "Asset preparation and deployment", Description: "Prepare once and ship a complete public asset tree."},
		{Path: "/troubleshooting", Label: "Troubleshooting", Group: "Reference", Template: "troubleshooting.gohtml", Title: "Troubleshooting asset loading", Description: "Trace preparation, publication, and browser resolution separately."},
		{Path: "/api", Label: "API reference", Group: "Reference", Template: "api.gohtml", Title: "go-importmap API", Description: "Configure inputs, prepare files, and choose an output format."},
	})
}
