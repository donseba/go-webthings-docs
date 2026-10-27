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
	chrome    *chrome
	origin    string
	basePath  string
	appName   string
	logName   string
	logoImage string
	title     string
	subtitle  string
	gitHubURL string
	nav       []NavItem
}

type docsRendererConfig struct {
	// Chrome is the header and rail of the subdomain; nil is the docs'.
	Chrome *chrome
	// Origin is the production address of the site, for canonical URLs;
	// empty is https://docs.gowebthings.com.
	Origin    string
	BasePath  string
	AppName   string
	LogName   string
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

// DocsHeaderPage is the header: the subdomain's name and the website's
// sections.
type DocsHeaderPage struct {
	Title     string
	Subtitle  string
	LogoImage string
	Links     []HeaderLink
	GitHubURL string
}

// HeaderLink is a link to another part of the website in the docs header.
type HeaderLink struct {
	Label  string
	URL    string
	Active bool
}

// DocsNavPage is the second column: the pages of the current component or
// module, under its name. A page without one has no Nav and hides it.
type DocsNavPage struct {
	Head   *DocsNavHead
	Nav    []DocsNavLink
	Groups []string
}

// DocsNavHead names the component or module the pages belong to.
type DocsNavHead struct {
	Title     string
	Subtitle  string
	URL       string
	Image     string
	Icon      string
	SourceURL string
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
	Rail        DocsRailPage
	Sidebar     DocsNavPage
	Nav         []NavItem
	SEO         SEOData
}

type docsPage struct {
	Template string
	// Shared are more template files the page uses, such as defined blocks.
	Shared      []string
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
	if cfg.Origin == "" {
		cfg.Origin = "https://docs.gowebthings.com"
	}
	if cfg.Chrome == nil {
		cfg.Chrome = docsChrome
	}

	funcs := []template.FuncMap{templatehelpers.FuncMap()}
	funcs = append(funcs, cfg.Funcs...)
	root := partial.NewID("shell", "templates/general/layout.gohtml").
		SetConnector(connector.NewHTMX4(nil)).
		SetFileSystem(websiteFS).
		SetBasePath(cfg.BasePath).
		UseTemplateCache(true).
		Use(exterrors.Stage(exterrors.WithMode(exterrors.ModeDetailed))).
		SetFunc(funcs...)

	return &docsRenderer{
		root:      root,
		chrome:    cfg.Chrome,
		origin:    cfg.Origin,
		basePath:  cfg.BasePath,
		appName:   cfg.AppName,
		logName:   cfg.LogName,
		logoImage: cfg.LogoImage,
		title:     cfg.Title,
		subtitle:  cfg.Subtitle,
		gitHubURL: cfg.GitHubURL,
		nav:       cfg.Nav,
	}
}

func (renderer *docsRenderer) render(w http.ResponseWriter, r *http.Request, page docsPage, dot any, configure ...func(*partial.Partial)) {
	sidebar := docsNavPage(renderer.nav, renderer.basePath, r.URL.Path)
	sidebar.Head = renderer.navHead()
	renderer.write(w, r, http.StatusOK, page, dot, sidebar, configure...)
}

// write renders a page with the given sidebar and status.
func (renderer *docsRenderer) write(w http.ResponseWriter, r *http.Request, status int, page docsPage, dot any, sidebar DocsNavPage, configure ...func(*partial.Partial)) {
	header := renderer.chrome.header(r)
	rail := renderer.chrome.rail(r, renderer.appName)
	data := DocsShellData{
		Title:       page.Title,
		Description: page.Description,
		Section:     page.Section,
		AppName:     renderer.appName,
		Path:        r.URL.Path,
		BasePath:    renderer.basePath,
		Header:      header,
		Rail:        rail,
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
	root.WithOOB(docsOOBPartial("rail", "templates/general/app_rail.gohtml", rail))
	root.WithOOB(docsOOBPartial("sidebar", "templates/general/app_sidebar.gohtml", sidebar))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	if err := partial.Write(r.Context(), w, r, root); err != nil {
		log.Printf("render %s docs error: %v", renderer.logName, err)
	}
}

func (page docsPage) partial(dot any) *partial.Partial {
	templates := append([]string{page.Template, docsHeroTemplate}, page.Shared...)

	return partial.NewID("docs-page", templates...).SetDot(dot)
}

func docsPages(templateDir string, pages map[string]docsPage) map[string]docsPage {
	templateDir = strings.TrimSuffix(templateDir, "/")
	for path, page := range pages {
		page.Template = templateDir + "/" + page.Template
		pages[path] = page
	}
	return pages
}

// navHead names the component above its pages.
func (renderer *docsRenderer) navHead() *DocsNavHead {
	if len(renderer.nav) == 0 {
		return nil
	}

	return &DocsNavHead{
		Title:     renderer.title,
		Subtitle:  renderer.subtitle,
		URL:       renderer.basePath,
		Image:     renderer.logoImage,
		SourceURL: renderer.gitHubURL,
	}
}

func docsLogoImage(appName string) string {
	switch appName {
	case "go-partial":
		return "/assets/img/logo-go-partial.png"
	case "go-clue":
		return "/assets/img/logo-go-clue.png"
	case "go-router":
		return "/assets/img/logo-go-router.png"
	case "go-form", "go-importmap", "go-translator":
		return "/assets/img/logo-" + appName + ".png"
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
		Canonical:   canonicalURL(renderer.origin, r),
		Image:       absoluteAssetURL(renderer.origin, docsSEOImage(renderer.appName)),
		SiteName:    renderer.title + " docs",
		Type:        "article",
	}
}

func docsSEOImage(appName string) string {
	switch appName {
	case "go-partial":
		return "/assets/img/logo-go-partial.png"
	case "go-clue":
		return "/assets/img/logo-go-clue.png"
	case "go-router":
		return "/assets/img/logo-go-router.png"
	case "go-form", "go-importmap", "go-translator":
		return "/assets/img/logo-" + appName + ".png"
	default:
		return "/assets/img/logo-go-webthings.png"
	}
}

func canonicalURL(origin string, r *http.Request) string {
	path := "/"
	if r.URL != nil && r.URL.Path != "" {
		path = r.URL.Path
	}
	return origin + path
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
