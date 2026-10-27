// Package moduledoc describes go-webthings for the website: its bundled
// modules and its user guides. tools/moduledocs reads them from a
// go-webthings checkout (code, README files, docs/) and writes JSON files;
// the website loads those files.
package moduledoc

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// Groups are the module groups in the order the website shows them.
var Groups = []string{"Content", "Visitors", "Site"}

// Module is everything the website shows about one module.
type Module struct {
	// Slug is the package name, for example "publishing".
	Slug string
	// Name is the README's title, for example "Publishing".
	Name string
	// Group is one of Groups.
	Group string
	// Summary is the README's first paragraph as plain text.
	Summary string
	// Deprecated explains why the module is going away; empty when it is not.
	Deprecated string
	ImportPath string
	SourceURL  string
	// Type is webthings.Module.Type: the schema owner and template folder.
	Type string
	// Example is the instance the reference was generated with; admin paths
	// and template folders use its key.
	Example Instance
	// Guide is the README after the title and summary.
	Guide template.HTML

	Constructor Func
	Options     []Func
	Parts       Parts
	Admin       []AdminSection
	Settings    *Type
	Routes      []Route

	Funcs     []Func
	Templates []Template
	// Components are parts of a page the module's template functions render,
	// with a default markup themes can replace.
	Components []Template
	Mail       []Template
	// Forms are the visitor forms themes render with the form_* template
	// functions (webthings.Module.Forms).
	Forms []Form

	Migrations []Migration
	Tables     []Table

	Translations []string
	API          API
}

// Instance is an example instance of a module.
type Instance struct {
	Key   string
	Title string
	// Code adds the instance to an app, for example
	// app.Use(publishing.Module("news", "News")).
	Code string
}

// Func is a function, method or template function.
type Func struct {
	Name string
	// Signature is the declaration without the body; template functions
	// leave out their context.Context parameter.
	Signature string
	Doc       template.HTML
}

// Parts are the webthings.Parts a module returns from Setup.
type Parts struct {
	// Public means the module can be placed on a page.
	Public bool
	// Actions means visitors post forms to /_/<key>/.
	Actions bool
	// Feeds means the module serves Atom and RSS feeds below its page.
	Feeds bool
	// Links means the module reports the links in its content to the
	// broken-link check.
	Links bool
	// API means the module has endpoints in the content API.
	API        bool
	Middleware bool
	Search     bool
	Sitemap    bool
	// Scheduled means editors can plan publication.
	Scheduled bool
	Close     bool
	// Service is the type of the module's API, for example "*publishing.Service".
	Service string
}

// AdminSection is a section the module adds to the admin.
type AdminSection struct {
	ID         string
	Title      string
	Path       string
	Permission string
	Icon       string
	Hidden     bool
	AdminOnly  bool
	Badge      bool
	// Group is the navigation heading the section sets itself; empty lists
	// it under the module's group, or under Modules.
	Group string
	// NewLabel is the section's quick action on the dashboard.
	NewLabel string
	// Recent and Scheduled mean the section feeds the dashboard's recently
	// edited and scheduled lists.
	Recent    bool
	Scheduled bool
	// Trash means deleted items of the section go to the admin's trash.
	Trash bool
	// Review means the section feeds the dashboard's Needs review list,
	// when the review workflow is on for it.
	Review bool
}

// Route is a public address the module serves.
type Route struct {
	// Kind is "page" for addresses below the page the module is placed on,
	// "feed" for its Atom and RSS feeds there, "api" for its endpoints of the
	// content API at /api/v1/<key>/, and "action" for form posts to
	// /_/<key>/.
	Kind string
	// Name is the route's name, for building its address (RoutePath); an
	// action route's name has the instance key in front, such as
	// "guestbook.post".
	Name   string
	Method string
	// Path is relative to the page, or the full /_/<key>/... address.
	Path string
}

// Template is a template file the module renders, or an e-mail template.
type Template struct {
	Name string
	// Files are the lookup paths in a theme folder, the first found wins.
	Files []string
	// DataType is the template's @dot, for example "publishing.Post".
	DataType string
	Fields   []Field
	// Starter is the module's hand-written example, or the scaffold
	// frontend.Scaffold generates from the data type. For components it is
	// the default markup.
	Starter string
}

// Form is a visitor form of a module: the input type its action maps a post
// to.
type Form struct {
	// Type is the input type, for example "comments.PostInput".
	Type   string
	Doc    template.HTML
	Fields []FormField
}

// FormField is a field of a visitor form.
type FormField struct {
	// Key is the field's name in the form, which form_field takes.
	Key string
	// Control is the go-form control, for example "input email" or
	// "textarea".
	Control     string
	Label       string
	Description string
	Required    bool
	// Rules are the other checks in words, for example "at most 100
	// characters".
	Rules []string
}

// Field is a field of a template's data, with its own fields when it is a
// struct.
type Field struct {
	Name   string
	Type   string
	Doc    string
	Fields []Field
}

// Type is a documented Go type.
type Type struct {
	Name   string
	Doc    template.HTML
	Fields []Field
	// Decl is the Go declaration, with its comments.
	Decl string
}

// Migration is one version of the module's schema.
type Migration struct {
	Version     string
	Description string
	// SQL is the migration per flavour: sqlite, postgres and mysql.
	SQL []Flavour
}

// Flavour is a migration file for one database.
type Flavour struct {
	Name string
	File string
	SQL  string
}

// Table is a table the module's schema creates.
type Table struct {
	Name    string
	Note    string
	Columns []Column
	Indexes []string
}

// Column is a column of a table.
type Column struct {
	Name string
	Type string
	// Notes are the constraints, for example "NOT NULL DEFAULT ''".
	Notes string
}

// API is the module's Go API.
type API struct {
	Funcs   []Func
	Methods []Func
	Types   []Type
}

// Pages are the pages of a module on the website, in navigation order.
var Pages = []Page{
	{Path: "", Label: "Overview"},
	{Path: "configuration", Label: "Configuration"},
	{Path: "templates", Label: "Templates"},
	{Path: "database", Label: "Database"},
	{Path: "api", Label: "Go API"},
}

// Page is a page of a module.
type Page struct {
	Path  string
	Label string
}

// Load reads every module in dir of fsys, sorted by group and name.
func Load(fsys fs.FS, dir string) ([]Module, error) {
	files, err := fs.Glob(fsys, path.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}

	modules := make([]Module, 0, len(files))

	for _, file := range files {
		body, err := fs.ReadFile(fsys, file)
		if err != nil {
			return nil, err
		}

		var module Module

		if err := json.Unmarshal(body, &module); err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}

		modules = append(modules, module)
	}

	Sort(modules)

	return modules, nil
}

// Sort orders modules by group, then by name.
func Sort(modules []Module) {
	order := map[string]int{}

	for i, group := range Groups {
		order[group] = i
	}

	sort.SliceStable(modules, func(i, j int) bool {
		a, b := modules[i], modules[j]

		if order[a.Group] != order[b.Group] {
			return order[a.Group] < order[b.Group]
		}

		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})
}
