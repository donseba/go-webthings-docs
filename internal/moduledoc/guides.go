package moduledoc

import (
	"encoding/json"
	"html/template"
	"io/fs"
)

// Guides are go-webthings' user guides, from its docs/ folder: README.md is
// the table of contents, every other file a guide.
type Guides struct {
	Title   string
	Summary string
	Groups  []GuideGroup
}

// GuideGroup is a heading of the table of contents with its guides.
type GuideGroup struct {
	Name   string
	Guides []Guide
}

// Guide is one guide.
type Guide struct {
	// Slug is the file name without .md, for example "getting-started".
	Slug    string
	Title   string
	Summary string
	Body    template.HTML
}

// LoadGuides reads the guides file of fsys.
func LoadGuides(fsys fs.FS, file string) (Guides, error) {
	body, err := fs.ReadFile(fsys, file)
	if err != nil {
		return Guides{}, err
	}

	var guides Guides

	err = json.Unmarshal(body, &guides)

	return guides, err
}
