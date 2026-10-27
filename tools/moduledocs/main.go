// Command moduledocs writes the data of the modules website: one JSON file
// per bundled go-webthings module, read from the module's code and README.
//
// Run it from the repository root with a go-webthings checkout next to it:
//
//	task modules
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"donseba/go-webthings-docs/internal/moduledoc"

	webthings "github.com/donseba/go-webthings"
	"github.com/donseba/go-webthings/modules/banners"
	"github.com/donseba/go-webthings/modules/comments"
	"github.com/donseba/go-webthings/modules/directory"
	"github.com/donseba/go-webthings/modules/documents"
	"github.com/donseba/go-webthings/modules/events"
	"github.com/donseba/go-webthings/modules/forms"
	"github.com/donseba/go-webthings/modules/gallery"
	"github.com/donseba/go-webthings/modules/lists"
	"github.com/donseba/go-webthings/modules/newsletter"
	"github.com/donseba/go-webthings/modules/people"
	"github.com/donseba/go-webthings/modules/polls"
	"github.com/donseba/go-webthings/modules/publishing"
	"github.com/donseba/go-webthings/modules/redirects"
	"github.com/donseba/go-webthings/modules/registrations"
	"github.com/donseba/go-webthings/modules/stats"
	"github.com/donseba/go-webthings/modules/submissions"
	"github.com/donseba/go-webthings/modules/testimonials"
)

const (
	corePath   = "github.com/donseba/go-webthings"
	sourceURL  = "https://github.com/donseba/go-webthings"
	sourceTree = sourceURL + "/tree/main/"
	sourceBlob = sourceURL + "/blob/main/"
)

// entry is a module the website documents, with the example instance its
// reference is generated with.
type entry struct {
	slug   string
	group  string
	module webthings.Module
	code   string
}

// entries are the documented modules. Several instances of one app, so the
// examples can refer to each other (registrations for agenda events, tips
// for news).
var entries = []entry{
	{
		slug:   "publishing",
		group:  "Content",
		module: publishing.Module("news", "News"),
		code:   `app.Use(publishing.Module("news", "News"))`,
	},
	{
		slug:   "events",
		group:  "Content",
		module: events.Module("agenda", "Agenda"),
		code:   `app.Use(events.Module("agenda", "Agenda"))`,
	},
	{
		slug:   "gallery",
		group:  "Content",
		module: gallery.Module("photos", "Photos"),
		code:   `app.Use(gallery.Module("photos", "Photos"))`,
	},
	{
		slug:   "lists",
		group:  "Content",
		module: lists.Module("faq", "FAQ"),
		code:   `app.Use(lists.Module("faq", "FAQ"))`,
	},
	{
		slug:   "people",
		group:  "Content",
		module: people.Module("team", "Team"),
		code:   `app.Use(people.Module("team", "Team"))`,
	},
	{
		slug:   "documents",
		group:  "Content",
		module: documents.Module("downloads", "Downloads"),
		code:   `app.Use(documents.Module("downloads", "Downloads"))`,
	},
	{
		slug:   "directory",
		group:  "Content",
		module: directory.Module("directory", "Directory"),
		code:   `app.Use(directory.Module("directory", "Directory"))`,
	},
	{
		slug:   "testimonials",
		group:  "Content",
		module: testimonials.Module("reviews", "Reviews"),
		code:   `app.Use(testimonials.Module("reviews", "Reviews"))`,
	},
	{
		slug:   "banners",
		group:  "Content",
		module: banners.Module("banners", "Banners"),
		code:   `app.Use(banners.Module("banners", "Banners"))`,
	},
	{
		slug:   "forms",
		group:  "Visitors",
		module: forms.Module("contact", "Contact"),
		code:   `app.Use(forms.Module("contact", "Contact"))`,
	},
	{
		slug:   "comments",
		group:  "Visitors",
		module: comments.Module("guestbook", "Guestbook"),
		code:   `app.Use(comments.Module("guestbook", "Guestbook"))`,
	},
	{
		slug:   "polls",
		group:  "Visitors",
		module: polls.Module("polls", "Polls"),
		code:   `app.Use(polls.Module("polls", "Polls"))`,
	},
	{
		slug:   "newsletter",
		group:  "Visitors",
		module: newsletter.Module("newsletter", "Newsletter"),
		code:   `app.Use(newsletter.Module("newsletter", "Newsletter"))`,
	},
	{
		slug:   "registrations",
		group:  "Visitors",
		module: registrations.Module("signup", "Sign-ups", registrations.ForEvents("agenda")),
		code: `app.Use(
	events.Module("agenda", "Agenda"),
	registrations.Module("signup", "Sign-ups", registrations.ForEvents("agenda")),
)`,
	},
	{
		slug:   "submissions",
		group:  "Visitors",
		module: submissions.Module("tips", "Tips", "news"),
		code: `app.Use(
	publishing.Module("news", "News"),
	submissions.Module("tips", "Tips", "news"),
)`,
	},
	{
		slug:   "redirects",
		group:  "Site",
		module: redirects.Module(),
		code:   `app.Use(redirects.Module())`,
	},
	{
		slug:   "stats",
		group:  "Site",
		module: stats.Module(),
		code:   `app.Use(stats.Module())`,
	},
}

func main() {
	root := flag.String("webthings", "../go-webthings", "go-webthings checkout, for the module sources, README files and guides")
	out := flag.String("out", "deploy/website/docs/modules", "directory the module JSON files are written to")
	guidesOut := flag.String("guides", "deploy/website/docs/guides/go-webthings.json", "file the guides are written to")

	flag.Parse()

	if err := run(*root, *out); err != nil {
		log.Fatal(err)
	}

	if err := writeGuides(*root, *guidesOut); err != nil {
		log.Fatal(err)
	}
}

// writeGuides writes go-webthings' user guides to file.
func writeGuides(root, file string) error {
	result, err := guides(root)
	if err != nil {
		return err
	}

	body, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}

	count := 0
	for _, group := range result.Groups {
		count += len(group.Guides)
	}

	fmt.Printf("%-14s %s (%d guides)\n", "guides", file, count)

	return os.WriteFile(file, append(body, '\n'), 0o644)
}

func run(root, out string) error {
	// Starting the app logs missing templates; that is expected here.
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	parts, err := inspect(context.Background(), entries)
	if err != nil {
		return err
	}

	sources := newSources(root)

	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}

	stale, err := filepath.Glob(filepath.Join(out, "*.json"))
	if err != nil {
		return err
	}

	for _, file := range stale {
		if err := os.Remove(file); err != nil {
			return err
		}
	}

	for _, e := range entries {
		module, err := describe(sources, e, parts[e.module.Key])
		if err != nil {
			return fmt.Errorf("%s: %w", e.slug, err)
		}

		body, err := json.MarshalIndent(module, "", "  ")
		if err != nil {
			return err
		}

		file := filepath.Join(out, e.slug+".json")

		if err := os.WriteFile(file, append(body, '\n'), 0o644); err != nil {
			return err
		}

		fmt.Printf("%-14s %s\n", e.slug, file)
	}

	return nil
}

// describe collects everything about one module.
func describe(sources *sources, e entry, parts captured) (moduledoc.Module, error) {
	importPath := corePath + "/modules/" + e.slug

	pkg, err := sources.load(importPath)
	if err != nil {
		return moduledoc.Module{}, err
	}

	readme, err := readREADME(sources.dir(importPath), e.slug)
	if err != nil {
		return moduledoc.Module{}, err
	}

	module := moduledoc.Module{
		Slug:       e.slug,
		Name:       readme.title,
		Group:      e.group,
		Summary:    readme.summary,
		Deprecated: pkg.deprecated(),
		ImportPath: importPath,
		SourceURL:  sourceTree + "modules/" + e.slug,
		Type:       e.module.Type,
		Example: moduledoc.Instance{
			Key:   e.module.Key,
			Title: e.module.Title,
			Code:  e.code,
		},
		Guide:        readme.body,
		Parts:        parts.parts,
		Admin:        parts.admin,
		Translations: translations(sources.dir(importPath)),
	}

	if module.Name == "" {
		module.Name = strings.ToUpper(e.slug[:1]) + e.slug[1:]
	}

	if module.Summary == "" {
		module.Summary = pkg.synopsis()
	}

	module.Constructor, module.Options = pkg.constructor()
	module.Settings = pkg.settings()
	module.API = pkg.api()
	module.Routes = parts.routes
	module.Funcs = pkg.templateFuncs(e.module.Funcs)
	module.Templates = templates(sources, e.module, e.module.Templates, "")
	module.Components = templates(sources, e.module, e.module.Components, "")
	module.Mail = templates(sources, e.module, e.module.Mail, "mail/")
	module.Forms = visitorForms(pkg, e.module.Forms)

	module.Migrations, module.Tables, err = schema(e.module.Schema)
	if err != nil {
		return moduledoc.Module{}, err
	}

	return module, nil
}
