package website

import "io/fs"

var (
	websiteFS  fs.FS = docsFileSystem()
	mainFS     fs.FS = mainFileSystem()
	showcaseFS fs.FS = showcaseFileSystem()
)

var (
	elementNames = []string{
		"go-partial",
		"go-docs",
		"go-router",
	}

	goPartialDocs = mustNewGoPartialDocs()
	goDocsDocs    = mustNewGoDocsDocs()
	goRouterDocs  = mustNewGoRouterDocs()
)

type Section string

const (
	SectionMain     Section = "main"
	SectionDocs     Section = "docs"
	SectionShowcase Section = "showcase"
)

type Element struct {
	Slug        string
	Name        string
	Description string
	Image       string
	Link        string
}

type Component struct {
	Slug        string
	Name        string
	Description string
	Image       string
	DocsURL     string
	ShowcaseURL string
	SourceURL   string
}

type PageData struct {
	Title            string
	Section          Section
	Host             string
	Element          *Element
	Elements         []Element
	Components       []Component
	Production       string
	Local            string
	Description      string
	SEO              SEOData
	ComingSoon       bool
	MainURL          string
	DocsURL          string
	ShowcaseURL      string
	SourceURL        string
	Bulletin         string
	PromptFace       string
	CurrentPath      string
	GenerateText     string
	GenerateImageURL string
}

type NavItem struct {
	Path  string
	Label string
	Group string
}

type SEOData struct {
	Title       string
	Description string
	Canonical   string
	Image       string
	SiteName    string
	Type        string
}
