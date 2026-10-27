package main

import (
	"bytes"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Links between the two subdomains are written with these schemes; the
// website replaces them with the address of the subdomain it serves.
const (
	docsScheme    = "wt-docs:"
	modulesScheme = "wt-modules:"
)

// guidesPath is where the guides are on the docs subdomain.
const guidesPath = "/go-webthings"

// markdownDoc is a Markdown file of go-webthings: the title, the first
// paragraph as its summary, and the rest as HTML.
type markdownDoc struct {
	title   string
	summary string
	body    template.HTML
}

var markdown = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
)

// readREADME reads a module's README.md; a module without one has an empty
// guide.
func readREADME(dir, slug string) (markdownDoc, error) {
	doc, err := readMarkdown(filepath.Join(dir, "README.md"), "modules/"+slug)
	if os.IsNotExist(err) {
		return markdownDoc{}, nil
	}

	return doc, err
}

// readMarkdown reads file; dir is its folder in the repository, for links.
func readMarkdown(file, dir string) (markdownDoc, error) {
	source, err := os.ReadFile(file)
	if err != nil {
		return markdownDoc{}, err
	}

	doc := markdown.Parser().Parse(text.NewReader(source))

	var md markdownDoc

	if heading, ok := doc.FirstChild().(*ast.Heading); ok && heading.Level == 1 {
		md.title = plain(heading, source)
		doc.RemoveChild(doc, heading)
	}

	if paragraph, ok := doc.FirstChild().(*ast.Paragraph); ok {
		md.summary = plain(paragraph, source)
		doc.RemoveChild(doc, paragraph)
	}

	rewriteLinks(doc, dir)

	var b bytes.Buffer

	if err := markdown.Renderer().Render(&b, source, doc); err != nil {
		return markdownDoc{}, err
	}

	body := strings.TrimSpace(b.String())

	// Tables scroll inside a wrapper on the website (see .module-table).
	body = strings.ReplaceAll(body, "<table>", `<div class="module-table"><table>`)
	body = strings.ReplaceAll(body, "</table>", "</table></div>")

	//nolint:gosec // goldmark escapes raw HTML in the Markdown.
	md.body = template.HTML(body)

	return md, nil
}

func rewriteLinks(doc ast.Node, dir string) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch link := n.(type) {
		case *ast.Link:
			link.Destination = []byte(rewriteLink(string(link.Destination), dir))
		case *ast.Image:
			link.Destination = []byte(rewriteLink(string(link.Destination), dir))
		}

		return ast.WalkContinue, nil
	})
}

var (
	moduleTarget = regexp.MustCompile(`^modules/([a-z0-9-]+)(?:/README\.md)?$`)
	guideTarget  = regexp.MustCompile(`^docs(?:/([a-z0-9-]+)\.md)?$`)
)

// rewriteLink points a relative link of a file in dir at the website: module
// READMEs at their pages on the modules subdomain, guides at theirs on the
// docs subdomain, and anything else at GitHub.
func rewriteLink(dest, dir string) string {
	if dest == "" || strings.HasPrefix(dest, "#") || strings.Contains(dest, "://") || strings.HasPrefix(dest, "mailto:") {
		return dest
	}

	target, fragment, found := strings.Cut(dest, "#")
	if found {
		fragment = "#" + fragment
	}

	resolved := path.Clean(path.Join(dir, target))
	if resolved == "docs/README.md" {
		resolved = "docs"
	}

	if m := moduleTarget.FindStringSubmatch(resolved); m != nil {
		if strings.HasPrefix(dir, "modules/") {
			return "/" + m[1] + fragment
		}

		return modulesScheme + "/" + m[1] + fragment
	}

	if m := guideTarget.FindStringSubmatch(resolved); m != nil {
		page := guidesPath
		if m[1] != "" {
			page += "/" + m[1]
		}

		if dir == "docs" {
			return page + fragment
		}

		return docsScheme + page + fragment
	}

	return sourceBlob + resolved + fragment
}

// plain is the text of a node without markdown.
func plain(n ast.Node, source []byte) string {
	var b strings.Builder

	_ = ast.Walk(n, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch t := child.(type) {
		case *ast.Text:
			b.Write(t.Segment.Value(source))

			if t.SoftLineBreak() || t.HardLineBreak() {
				b.WriteString(" ")
			}
		case *ast.String:
			b.Write(t.Value)
		case *ast.CodeSpan:
			for c := t.FirstChild(); c != nil; c = c.NextSibling() {
				if segment, ok := c.(*ast.Text); ok {
					b.Write(segment.Segment.Value(source))
				}
			}

			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})

	return strings.Join(strings.Fields(b.String()), " ")
}
