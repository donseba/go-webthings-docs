package website

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"

	partial "github.com/donseba/go-partial"
	"github.com/donseba/go-partial/connector"
	"github.com/donseba/go-partial/exp/templatehelpers"
	exterrors "github.com/donseba/go-partial/ext/errors"
)

type docsRenderer struct {
	root      *partial.Partial
	basePath  string
	appName   string
	logName   string
	logo      string
	logoImage string
	title     string
	subtitle  string
	gitHubURL string
	nav       []NavItem
}

type docsRendererConfig struct {
	BasePath  string
	AppName   string
	LogName   string
	Logo      string
	LogoImage string
	Title     string
	Subtitle  string
	GitHubURL string
	Nav       []NavItem
	Funcs     []template.FuncMap
}

type DocsArticleData struct {
	Title       string
	Description string
	Section     string
	PromptFace  string
}

type DocsHeaderPage struct {
	AppName     string
	BasePath    string
	MainURL     string
	DocsURL     string
	ShowcaseURL string
	Logo        string
	LogoImage   string
	Title       string
	Subtitle    string
	GitHubURL   string
}

type DocsNavPage struct {
	Nav    []DocsNavLink
	Groups []string
}

type DocsNavLink struct {
	Path   string
	Label  string
	Group  string
	Active bool
}

type DocsShellData struct {
	Title       string
	Description string
	Section     string
	AppName     string
	Path        string
	BasePath    string
	Header      DocsHeaderPage
	Sidebar     DocsNavPage
	Nav         []NavItem
	SEO         SEOData
}

type docsPage struct {
	Template    string
	Title       string
	Description string
	Section     string
	Configure   func(*partial.Partial)
}

const docsHeroTemplate = "templates/general/hero.gohtml"

var docsPromptFaces = []string{
	"#_>",
	"-_-",
	"^-^",
	"#_#",
	">_>",
	"<_<",
	"0_0",
	"o_o",
	"x_x",
	"^_^",
	"._.",
	">_<",
	"v_v",
	"u_u",
	"T_T",
	"n_n",
	"*_*",
	"+_+",
	"@_@",
	"$_$",
	"?_?",
	"!_!",
	"/_/",
	"_~_",
}

func newDocsRenderer(cfg docsRendererConfig) *docsRenderer {
	if cfg.LogName == "" {
		cfg.LogName = cfg.AppName
	}
	if cfg.LogoImage == "" {
		cfg.LogoImage = docsLogoImage(cfg.AppName)
	}

	funcs := []template.FuncMap{templatehelpers.FuncMap()}
	funcs = append(funcs, cfg.Funcs...)
	root := partial.NewID("shell", "templates/general/layout.gohtml").
		SetConnector(connector.NewHTMX(nil)).
		SetFileSystem(websiteFS).
		SetBasePath(cfg.BasePath).
		UseTemplateCache(true).
		Use(exterrors.Stage(exterrors.WithMode(exterrors.ModeDetailed))).
		SetFunc(funcs...)

	return &docsRenderer{
		root:      root,
		basePath:  cfg.BasePath,
		appName:   cfg.AppName,
		logName:   cfg.LogName,
		logo:      cfg.Logo,
		logoImage: cfg.LogoImage,
		title:     cfg.Title,
		subtitle:  cfg.Subtitle,
		gitHubURL: cfg.GitHubURL,
		nav:       cfg.Nav,
	}
}

func (renderer *docsRenderer) render(w http.ResponseWriter, r *http.Request, page docsPage, dot any, configure ...func(*partial.Partial)) {
	header := renderer.header(r)
	sidebar := docsNavPage(renderer.nav, renderer.basePath, r.URL.Path)
	data := DocsShellData{
		Title:       page.Title,
		Description: page.Description,
		Section:     page.Section,
		AppName:     renderer.appName,
		Path:        r.URL.Path,
		BasePath:    renderer.basePath,
		Header:      header,
		Sidebar:     sidebar,
		Nav:         renderer.nav,
		SEO:         renderer.seo(r, page),
	}

	if dot == nil {
		dot = DocsArticleData{
			Title:       page.Title,
			Description: page.Description,
			Section:     page.Section,
			PromptFace:  docsPromptFaces[rand.Intn(len(docsPromptFaces))],
		}
	}
	pageContent := page.partial(dot)
	content := partial.NewID("content", "templates/general/content.gohtml").SetDot(data).SetContent(pageContent)
	if page.Configure != nil {
		page.Configure(pageContent)
	}
	for _, fn := range configure {
		if fn != nil {
			fn(pageContent)
		}
	}

	root := renderer.root.Clone().SetDot(data).SetContent(content)
	root.WithOOB(docsOOBPartial("header", "templates/general/app_header.gohtml", header))
	root.WithOOB(docsOOBPartial("sidebar", "templates/general/app_sidebar.gohtml", sidebar))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := partial.Write(r.Context(), w, r, root); err != nil {
		log.Printf("render %s docs error: %v", renderer.logName, err)
	}
}

func (page docsPage) partial(dot any) *partial.Partial {
	return partial.NewID("docs-page", page.Template, docsHeroTemplate).SetDot(dot)
}

func docsPages(templateDir string, pages map[string]docsPage) map[string]docsPage {
	templateDir = strings.TrimSuffix(templateDir, "/")
	for path, page := range pages {
		page.Template = templateDir + "/" + page.Template
		pages[path] = page
	}
	return pages
}

func (renderer *docsRenderer) header(r *http.Request) DocsHeaderPage {
	return DocsHeaderPage{
		AppName:     renderer.appName,
		BasePath:    renderer.basePath,
		MainURL:     mainWebsiteURL(r),
		DocsURL:     docsHomeURL(r),
		ShowcaseURL: mainFamilyURL(r, "showcase", ""),
		Logo:        renderer.logo,
		LogoImage:   renderer.logoImage,
		Title:       renderer.title,
		Subtitle:    renderer.subtitle,
		GitHubURL:   renderer.gitHubURL,
	}
}

func docsLogoImage(appName string) string {
	switch appName {
	case "go-partial":
		return "/assets/img/logo-go-partial.png"
	case "go-docs":
		return "/assets/img/logo-go-doc.png"
	case "go-router":
		return "/assets/img/logo-go-router.png"
	default:
		return ""
	}
}

func (renderer *docsRenderer) seo(r *http.Request, page docsPage) SEOData {
	description := page.Description
	if description == "" {
		description = renderer.subtitle
	}
	return SEOData{
		Title:       fmt.Sprintf("%s - %s docs", page.Title, renderer.title),
		Description: description,
		Canonical:   docsCanonicalURL(r),
		Image:       absoluteAssetURL("https://docs.gowebthings.com", docsSEOImage(renderer.appName)),
		SiteName:    renderer.title + " docs",
		Type:        "article",
	}
}

func docsSEOImage(appName string) string {
	switch appName {
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

func docsCanonicalURL(r *http.Request) string {
	path := "/"
	if r.URL != nil && r.URL.Path != "" {
		path = r.URL.Path
	}
	return "https://docs.gowebthings.com" + path
}

func absoluteAssetURL(origin, path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return strings.TrimRight(origin, "/") + "/" + strings.TrimPrefix(path, "/")
}

func docsOOBPartial(id, tmpl string, dot any) *partial.Partial {
	return partial.NewID(id, tmpl).
		SetFileSystem(websiteFS).
		SetDot(dot).
		SetAlwaysSwapOOB(true)
}

func docsNavPage(items []NavItem, basePath, currentPath string) DocsNavPage {
	links := make([]DocsNavLink, 0, len(items))
	groups := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	currentPath = strings.TrimSuffix(currentPath, "/")
	if currentPath == "" {
		currentPath = "/"
	}

	for _, item := range items {
		if _, ok := seen[item.Group]; !ok {
			seen[item.Group] = struct{}{}
			groups = append(groups, item.Group)
		}

		path := docsFullPath(basePath, item.Path)
		links = append(links, DocsNavLink{
			Path:   path,
			Label:  item.Label,
			Group:  item.Group,
			Active: strings.TrimSuffix(path, "/") == currentPath,
		})
	}

	return DocsNavPage{
		Nav:    links,
		Groups: groups,
	}
}

func docsFullPath(basePath, path string) string {
	if path == "" || path == "/" {
		return basePath
	}
	return basePath + path
}
