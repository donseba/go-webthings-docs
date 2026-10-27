package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"donseba/go-webthings-docs/internal/moduledoc"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var guideFile = regexp.MustCompile(`^([a-z0-9-]+)\.md$`)

// guides reads go-webthings' docs/: README.md is the table of contents, with
// a heading per group and a list of links to the guides in reading order.
func guides(root string) (moduledoc.Guides, error) {
	dir := filepath.Join(root, "docs")

	source, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		return moduledoc.Guides{}, err
	}

	doc := markdown.Parser().Parse(text.NewReader(source))

	var (
		result moduledoc.Guides
		group  *moduledoc.GuideGroup
	)

	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		switch node := n.(type) {
		case *ast.Heading:
			if node.Level == 1 {
				result.Title = plain(node, source)

				continue
			}

			result.Groups = append(result.Groups, moduledoc.GuideGroup{Name: plain(node, source)})
			group = &result.Groups[len(result.Groups)-1]
		case *ast.Paragraph:
			if result.Summary == "" && group == nil {
				result.Summary = plain(node, source)
			}
		case *ast.List:
			if group == nil {
				continue
			}

			for _, slug := range listedGuides(node) {
				guide, err := readGuide(dir, slug)
				if os.IsNotExist(err) {
					fmt.Printf("guide %s.md is in the table of contents but does not exist yet\n", slug)

					continue
				}

				if err != nil {
					return moduledoc.Guides{}, fmt.Errorf("%s.md: %w", slug, err)
				}

				group.Guides = append(group.Guides, guide)
			}
		}
	}

	return result, nil
}

// listedGuides are the guides a table of contents list links to.
func listedGuides(list *ast.List) []string {
	var slugs []string

	_ = ast.Walk(list, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		link, ok := n.(*ast.Link)
		if !entering || !ok {
			return ast.WalkContinue, nil
		}

		if m := guideFile.FindStringSubmatch(string(link.Destination)); m != nil {
			slugs = append(slugs, m[1])
		}

		return ast.WalkSkipChildren, nil
	})

	return slugs
}

func readGuide(dir, slug string) (moduledoc.Guide, error) {
	doc, err := readMarkdown(filepath.Join(dir, slug+".md"), "docs")
	if err != nil {
		return moduledoc.Guide{}, err
	}

	if doc.title == "" {
		return moduledoc.Guide{}, fmt.Errorf("no title")
	}

	return moduledoc.Guide{
		Slug:    slug,
		Title:   doc.title,
		Summary: doc.summary,
		Body:    doc.body,
	}, nil
}
