package main

import (
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"donseba/go-webthings-docs/internal/moduledoc"

	webthings "github.com/donseba/go-webthings"
	"github.com/donseba/go-webthings/frontend"
)

// maxDepth limits how deep template data is expanded.
const maxDepth = 3

// templates documents a module's page or e-mail templates; prefix is "" for
// pages and "mail/" for e-mail.
func templates(s *sources, module webthings.Module, list []frontend.Template, prefix string) []moduledoc.Template {
	docs := make([]moduledoc.Template, 0, len(list))

	for _, t := range list {
		doc := moduledoc.Template{
			Name: t.Name,
			Files: []string{
				prefix + module.Key + "/" + t.Name + ".gohtml",
				prefix + module.Type + "/" + t.Name + ".gohtml",
			},
			Starter: strings.TrimSpace(t.Example),
		}

		if doc.Starter == "" {
			doc.Starter = strings.TrimSpace(frontend.Scaffold(t))
		}

		if t.Data != nil {
			rt := reflect.TypeOf(t.Data)
			doc.DataType = rt.String()
			doc.Fields = fields(s, rt, 0, map[reflect.Type]bool{})
		}

		docs = append(docs, doc)
	}

	return docs
}

// fields are the exported fields of a struct type, expanded into the
// go-webthings structs they hold.
func fields(s *sources, rt reflect.Type, depth int, seen map[reflect.Type]bool) []moduledoc.Field {
	rt = element(rt)
	if rt.Kind() != reflect.Struct || seen[rt] || depth >= maxDepth {
		return nil
	}

	seen[rt] = true
	defer delete(seen, rt)

	var list []moduledoc.Field

	for i := range rt.NumField() {
		f := rt.Field(i)
		if !f.IsExported() {
			continue
		}

		field := moduledoc.Field{
			Name: f.Name,
			Type: f.Type.String(),
			Doc:  s.fieldDoc(rt.PkgPath(), rt.Name(), f.Name),
		}

		if expandable(f.Type) {
			field.Fields = fields(s, f.Type, depth+1, seen)
		}

		list = append(list, field)
	}

	return list
}

// element is the type a slice, array, pointer or map holds.
func element(rt reflect.Type) reflect.Type {
	for {
		switch rt.Kind() {
		case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Map:
			rt = rt.Elem()
		default:
			return rt
		}
	}
}

// expandable reports whether a field's fields are worth listing: structs of
// go-webthings itself, not time.Time and the like.
func expandable(rt reflect.Type) bool {
	rt = element(rt)

	return rt.Kind() == reflect.Struct && strings.HasPrefix(rt.PkgPath(), corePath)
}

// translations are the admin languages a module ships, from translations/.
func translations(dir string) []string {
	files, _ := filepath.Glob(filepath.Join(dir, "translations", "*.po"))

	locales := make([]string, 0, len(files))

	for _, file := range files {
		locales = append(locales, strings.TrimSuffix(filepath.Base(file), ".po"))
	}

	sort.Strings(locales)

	return locales
}
