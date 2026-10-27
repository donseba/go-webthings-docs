package main

import (
	"fmt"
	"reflect"
	"strings"

	"donseba/go-webthings-docs/internal/moduledoc"
)

// visitorForms are the module's visitor forms (webthings.Module.Forms): the input
// types its actions map posts to, with the fields as go-form reads their
// tags.
func visitorForms(pkg *pkg, models []any) []moduledoc.Form {
	var list []moduledoc.Form

	for _, model := range models {
		t := reflect.TypeOf(model)
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}

		if t.Kind() != reflect.Struct {
			continue
		}

		form := moduledoc.Form{Type: t.String()}

		for _, docType := range pkg.doc.Types {
			if docType.Name == t.Name() {
				form.Doc = pkg.html(docType.Doc)
			}
		}

		form.Fields = formFields(t)
		list = append(list, form)
	}

	return list
}

// formFields are the fields of t that go-form renders: exported fields
// with a form tag other than "-".
func formFields(t reflect.Type) []moduledoc.FormField {
	var fields []moduledoc.FormField

	for field := range t.Fields() {
		tag := field.Tag
		control := tag.Get("form")

		if !field.IsExported() || control == "" || control == "-" {
			continue
		}

		key := tag.Get("name")
		if key == "" {
			key = field.Name
		}

		fields = append(fields, moduledoc.FormField{
			Key:         key,
			Control:     strings.ReplaceAll(control, ",", " "),
			Label:       tag.Get("label"),
			Description: tag.Get("description"),
			Required:    tag.Get("required") == "true",
			Rules:       formRules(tag),
		})
	}

	return fields
}

// formRules are the checks of a field's tags besides required, in words.
func formRules(tag reflect.StructTag) []string {
	var rules []string

	if tag.Get("line") == "true" {
		rules = append(rules, "one line")
	}

	if v := tag.Get("minLength"); v != "" {
		rules = append(rules, fmt.Sprintf("at least %s characters", v))
	}

	if v := tag.Get("maxLength"); v != "" {
		rules = append(rules, fmt.Sprintf("at most %s characters", v))
	}

	minimum, maximum := tag.Get("min"), tag.Get("max")

	switch {
	case minimum != "" && maximum != "":
		rules = append(rules, fmt.Sprintf("%s to %s", minimum, maximum))
	case minimum != "":
		rules = append(rules, "at least "+minimum)
	case maximum != "":
		rules = append(rules, "at most "+maximum)
	}

	if v := tag.Get("pattern"); v != "" {
		rules = append(rules, "matches "+v)
	}

	if v := tag.Get("accept"); v != "" {
		rules = append(rules, "accepts "+strings.ReplaceAll(v, ",", ", "))
	}

	if v := tag.Get("maxSize"); v != "" {
		rules = append(rules, "at most "+v)
	}

	return rules
}
