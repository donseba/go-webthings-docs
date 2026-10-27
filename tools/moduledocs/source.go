package main

import (
	"bytes"
	"go/ast"
	"go/build"
	"go/doc"
	"go/doc/comment"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"donseba/go-webthings-docs/internal/moduledoc"
)

// sources parses go-webthings packages on demand.
type sources struct {
	root     string
	packages map[string]*pkg
}

func newSources(root string) *sources {
	return &sources{root: root, packages: map[string]*pkg{}}
}

// dir is the directory of a go-webthings package; "" for other packages.
func (s *sources) dir(importPath string) string {
	if importPath == corePath {
		return s.root
	}

	rel, ok := strings.CutPrefix(importPath, corePath+"/")
	if !ok {
		return ""
	}

	return filepath.Join(s.root, filepath.FromSlash(rel))
}

// load parses a go-webthings package, without its tests.
func (s *sources) load(importPath string) (*pkg, error) {
	if p, ok := s.packages[importPath]; ok {
		return p, nil
	}

	dir := s.dir(importPath)
	if dir == "" {
		return nil, os.ErrNotExist
	}

	fset := token.NewFileSet()

	files, err := parseDir(fset, dir)
	if err != nil {
		return nil, err
	}

	// go/doc drops unexported declarations from the files it is given, so
	// it gets its own copy.
	docFiles, err := parseDir(fset, dir)
	if err != nil {
		return nil, err
	}

	docs, err := doc.NewFromFiles(fset, docFiles, importPath)
	if err != nil {
		return nil, err
	}

	p := &pkg{
		fset:   fset,
		files:  files,
		doc:    docs,
		funcs:  map[string]*ast.FuncDecl{},
		types:  map[string]*ast.TypeSpec{},
		parser: docs.Parser(),
	}

	p.printer = docs.Printer()
	p.printer.HeadingLevel = 4
	p.printer.DocLinkURL = func(*comment.DocLink) string {
		return ""
	}

	for _, file := range files {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil {
					p.funcs[d.Name.Name] = d
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						p.types[ts.Name.Name] = ts
					}
				}
			}
		}
	}

	s.packages[importPath] = p

	return p, nil
}

// parseDir parses the Go files of the package in dir that the build uses,
// without its tests, in file name order.
func parseDir(fset *token.FileSet, dir string) ([]*ast.File, error) {
	info, err := build.ImportDir(dir, 0)
	if err != nil {
		return nil, err
	}

	names := append([]string(nil), info.GoFiles...)
	sort.Strings(names)

	files := make([]*ast.File, 0, len(names))

	for _, name := range names {
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		files = append(files, file)
	}

	return files, nil
}

// fieldDoc is the comment of a struct field in a go-webthings package.
func (s *sources) fieldDoc(importPath, typeName, field string) string {
	p, err := s.load(importPath)
	if err != nil {
		return ""
	}

	spec, ok := p.types[typeName]
	if !ok {
		return ""
	}

	st, ok := spec.Type.(*ast.StructType)
	if !ok {
		return ""
	}

	for _, f := range st.Fields.List {
		for _, name := range f.Names {
			if name.Name == field {
				return fieldComment(f)
			}
		}

		if len(f.Names) == 0 && embeddedName(f.Type) == field {
			return fieldComment(f)
		}
	}

	return ""
}

// pkg is a parsed package.
type pkg struct {
	fset    *token.FileSet
	files   []*ast.File
	doc     *doc.Package
	funcs   map[string]*ast.FuncDecl
	types   map[string]*ast.TypeSpec
	parser  *comment.Parser
	printer *comment.Printer
}

// html renders a doc comment.
func (p *pkg) html(text string) template.HTML {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	//nolint:gosec // go/doc/comment escapes the comment text.
	return template.HTML(p.printer.HTML(p.parser.Parse(text)))
}

func (p *pkg) synopsis() string {
	return p.doc.Synopsis(p.doc.Doc)
}

// deprecated is the package doc's "Deprecated:" paragraph.
func (p *pkg) deprecated() string {
	for _, paragraph := range strings.Split(p.doc.Doc, "\n\n") {
		if text, ok := strings.CutPrefix(strings.TrimSpace(paragraph), "Deprecated:"); ok {
			return strings.Join(strings.Fields(text), " ")
		}
	}

	return ""
}

// constructor is the Module function and the functions that make its options.
func (p *pkg) constructor() (moduledoc.Func, []moduledoc.Func) {
	var module moduledoc.Func

	for _, f := range p.doc.Funcs {
		if f.Name == "Module" {
			module = p.function(f.Decl)
		}
	}

	var options []moduledoc.Func

	for _, t := range p.doc.Types {
		if t.Name != "Option" {
			continue
		}

		for _, f := range t.Funcs {
			options = append(options, p.function(f.Decl))
		}
	}

	return module, options
}

// api is the package's exported API besides Module and its options.
func (p *pkg) api() moduledoc.API {
	var api moduledoc.API

	for _, f := range p.doc.Funcs {
		if f.Name != "Module" {
			api.Funcs = append(api.Funcs, p.function(f.Decl))
		}
	}

	for _, t := range p.doc.Types {
		if t.Name == "Option" {
			continue
		}

		for _, f := range t.Funcs {
			api.Funcs = append(api.Funcs, p.function(f.Decl))
		}

		if t.Name == "Service" {
			for _, m := range t.Methods {
				if ast.IsExported(m.Name) {
					api.Methods = append(api.Methods, p.function(m.Decl))
				}
			}
		}

		api.Types = append(api.Types, p.typeDoc(t))
	}

	return api
}

// settings is the package's Settings type, if it has one.
func (p *pkg) settings() *moduledoc.Type {
	for _, t := range p.doc.Types {
		if t.Name != "Settings" {
			continue
		}

		settings := p.typeDoc(t)
		settings.Fields = p.structFields(t.Name)

		return &settings
	}

	return nil
}

// structFields are the exported fields of a struct type in the package.
func (p *pkg) structFields(name string) []moduledoc.Field {
	spec, ok := p.types[name]
	if !ok {
		return nil
	}

	st, ok := spec.Type.(*ast.StructType)
	if !ok {
		return nil
	}

	var fields []moduledoc.Field

	for _, f := range st.Fields.List {
		for _, ident := range f.Names {
			if !ident.IsExported() {
				continue
			}

			fields = append(fields, moduledoc.Field{
				Name: ident.Name,
				Type: p.node(f.Type),
				Doc:  fieldComment(f),
			})
		}
	}

	return fields
}

func (p *pkg) typeDoc(t *doc.Type) moduledoc.Type {
	return moduledoc.Type{
		Name: t.Name,
		Doc:  p.html(t.Doc),
		Decl: p.declaration(t),
	}
}

// function documents a function or method, without its body.
func (p *pkg) function(decl *ast.FuncDecl) moduledoc.Func {
	signature := *decl
	signature.Body = nil
	signature.Doc = nil

	return moduledoc.Func{
		Name:      decl.Name.Name,
		Signature: p.node(&signature),
		Doc:       p.html(decl.Doc.Text()),
	}
}

// templateFuncs documents the functions a module offers to templates. The
// names come from the module; signatures and comments from the functions
// its FuncMap literal names.
func (p *pkg) templateFuncs(funcs template.FuncMap) []moduledoc.Func {
	implementations := map[string]string{}

	for _, file := range p.files {
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok || !isFuncMap(lit.Type) {
				return true
			}

			for _, elt := range lit.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				key, ok := kv.Key.(*ast.BasicLit)
				if !ok || key.Kind != token.STRING {
					continue
				}

				name, err := strconv.Unquote(key.Value)
				if err != nil {
					continue
				}

				if ident, ok := kv.Value.(*ast.Ident); ok {
					implementations[name] = ident.Name
				}
			}

			return true
		})
	}

	names := make([]string, 0, len(funcs))
	for name := range funcs {
		names = append(names, name)
	}

	sort.Strings(names)

	list := make([]moduledoc.Func, 0, len(names))

	for _, name := range names {
		decl, ok := p.funcs[implementations[name]]
		if !ok {
			list = append(list, moduledoc.Func{
				Name:      name,
				Signature: name + " " + reflect.TypeOf(funcs[name]).String(),
			})

			continue
		}

		list = append(list, moduledoc.Func{
			Name:      name,
			Signature: name + strings.TrimPrefix(p.node(withoutContext(decl.Type)), "func"),
			Doc:       p.html(decl.Doc.Text()),
		})
	}

	return list
}

func isSelector(expr ast.Expr, pkg, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != name {
		return false
	}

	x, ok := sel.X.(*ast.Ident)

	return ok && x.Name == pkg
}

// declaration is a type's Go declaration with its exported fields and
// associated constants, formatted like gofmt.
func (p *pkg) declaration(t *doc.Type) string {
	var b strings.Builder

	spec := p.types[t.Name]
	if spec == nil {
		return ""
	}

	switch typ := spec.Type.(type) {
	case *ast.StructType:
		b.WriteString("type " + t.Name + " struct {\n")

		hidden := false

		for _, f := range typ.Fields.List {
			names := exportedNames(f)

			embedded := len(f.Names) == 0 && ast.IsExported(embeddedName(f.Type))

			if len(names) == 0 && !embedded {
				hidden = true

				continue
			}

			if f.Doc != nil {
				writeComment(&b, f.Doc.Text(), "\t")
			}

			b.WriteString("\t")

			if len(names) > 0 {
				b.WriteString(strings.Join(names, ", ") + " ")
			}

			b.WriteString(p.node(f.Type))

			if f.Tag != nil {
				b.WriteString(" " + f.Tag.Value)
			}

			if f.Comment != nil {
				b.WriteString(" // " + strings.TrimSpace(f.Comment.Text()))
			}

			b.WriteString("\n")
		}

		if hidden {
			b.WriteString("\t// contains unexported fields\n")
		}

		b.WriteString("}\n")
	default:
		b.WriteString("type " + t.Name)

		if spec.Assign.IsValid() {
			b.WriteString(" =")
		}

		b.WriteString(" " + p.node(spec.Type) + "\n")
	}

	for _, values := range [][]*doc.Value{t.Consts, t.Vars} {
		for _, value := range values {
			b.WriteString("\n" + p.values(value.Decl))
		}
	}

	return gofmt(b.String())
}

// values writes the exported names of a const or var block.
func (p *pkg) values(decl *ast.GenDecl) string {
	var b strings.Builder

	b.WriteString(decl.Tok.String() + " (\n")

	for _, spec := range decl.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		var names []string

		for _, name := range vs.Names {
			if name.IsExported() {
				names = append(names, name.Name)
			}
		}

		if len(names) == 0 {
			continue
		}

		if vs.Doc != nil {
			writeComment(&b, vs.Doc.Text(), "\t")
		}

		b.WriteString("\t" + strings.Join(names, ", "))

		if vs.Type != nil {
			b.WriteString(" " + p.node(vs.Type))
		}

		if len(vs.Values) > 0 && len(names) == len(vs.Names) {
			values := make([]string, 0, len(vs.Values))
			for _, v := range vs.Values {
				values = append(values, p.node(v))
			}

			b.WriteString(" = " + strings.Join(values, ", "))
		}

		if vs.Comment != nil {
			b.WriteString(" // " + strings.TrimSpace(vs.Comment.Text()))
		}

		b.WriteString("\n")
	}

	b.WriteString(")\n")

	return b.String()
}

// node prints an AST node.
func (p *pkg) node(n any) string {
	var b bytes.Buffer

	config := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	if err := config.Fprint(&b, p.fset, n); err != nil {
		return ""
	}

	return b.String()
}

// withoutContext drops a leading context.Context parameter: templates do not
// pass it.
func withoutContext(ft *ast.FuncType) *ast.FuncType {
	if ft.Params == nil || len(ft.Params.List) == 0 {
		return ft
	}

	sel, ok := ft.Params.List[0].Type.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Context" {
		return ft
	}

	if x, ok := sel.X.(*ast.Ident); !ok || x.Name != "context" {
		return ft
	}

	params := *ft.Params
	first := *params.List[0]

	if len(first.Names) > 1 {
		first.Names = first.Names[1:]
		params.List = append([]*ast.Field{&first}, params.List[1:]...)
	} else {
		params.List = params.List[1:]
	}

	copied := *ft
	copied.Params = &params

	return &copied
}

func isFuncMap(expr ast.Expr) bool {
	return isSelector(expr, "template", "FuncMap")
}

func exportedNames(f *ast.Field) []string {
	var names []string

	for _, name := range f.Names {
		if name.IsExported() {
			names = append(names, name.Name)
		}
	}

	return names
}

// embeddedName is the type name of an embedded field.
func embeddedName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return embeddedName(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	case *ast.IndexExpr:
		return embeddedName(t.X)
	}

	return ""
}

// fieldComment is a field's doc comment, or its line comment.
func fieldComment(f *ast.Field) string {
	text := f.Doc.Text()
	if text == "" {
		text = f.Comment.Text()
	}

	return strings.Join(strings.Fields(text), " ")
}

func writeComment(b *strings.Builder, text, indent string) {
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		if line == "" {
			b.WriteString(indent + "//\n")

			continue
		}

		b.WriteString(indent + "// " + line + "\n")
	}
}

// gofmt formats declarations; unformatted on a parse error.
func gofmt(decls string) string {
	formatted, err := format.Source([]byte("package p\n\n" + decls))
	if err != nil {
		return strings.TrimSpace(decls)
	}

	return strings.TrimSpace(strings.TrimPrefix(string(formatted), "package p\n"))
}
