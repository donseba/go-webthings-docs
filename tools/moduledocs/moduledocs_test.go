package main

import (
	"reflect"
	"testing"
	"testing/fstest"

	"donseba/go-webthings-docs/internal/moduledoc"
)

func TestSchema(t *testing.T) {
	fsys := fstest.MapFS{
		"20261004000000_init.postgres.sql": {Data: []byte(`-- Posts are revisioned content.
CREATE TABLE wt_post (
    id       UUID PRIMARY KEY,
    instance VARCHAR(64) NOT NULL,
    title    VARCHAR(200) NOT NULL DEFAULT '', -- shown in lists
    note     TEXT NOT NULL DEFAULT 'a, b',
    UNIQUE (instance, title)
);

CREATE INDEX wt_post_instance ON wt_post (instance);
`)},
		"20261004000000_init.sqlite.sql": {Data: []byte("CREATE TABLE wt_post (id TEXT PRIMARY KEY);")},
		"20261104000000_add_slug.postgres.sql": {Data: []byte(
			"ALTER TABLE wt_post ADD COLUMN slug VARCHAR(190) NOT NULL DEFAULT '';",
		)},
		"README.md": {Data: []byte("not a migration")},
	}

	migrations, tables, err := schema(fsys)
	if err != nil {
		t.Fatal(err)
	}

	if len(migrations) != 2 || migrations[1].Description != "add slug" {
		t.Fatalf("migrations: %+v", migrations)
	}

	if names := []string{migrations[0].SQL[0].Name, migrations[0].SQL[1].Name}; !reflect.DeepEqual(names, []string{"postgres", "sqlite"}) {
		t.Fatalf("flavours in order: %v", names)
	}

	want := moduledoc.Table{
		Name: "wt_post",
		Note: "Posts are revisioned content.",
		Columns: []moduledoc.Column{
			{Name: "id", Type: "UUID", Notes: "PRIMARY KEY"},
			{Name: "instance", Type: "VARCHAR(64)", Notes: "NOT NULL"},
			{Name: "title", Type: "VARCHAR(200)", Notes: "NOT NULL DEFAULT ''"},
			{Name: "note", Type: "TEXT", Notes: "NOT NULL DEFAULT 'a, b'"},
			{Name: "slug", Type: "VARCHAR(190)", Notes: "NOT NULL DEFAULT ''"},
		},
		Indexes: []string{"UNIQUE (instance, title)", "INDEX wt_post_instance (instance)"},
	}

	if len(tables) != 1 || !reflect.DeepEqual(tables[0], want) {
		t.Fatalf("tables:\n got %+v\nwant %+v", tables, want)
	}
}

func TestRewriteLink(t *testing.T) {
	for _, tc := range []struct {
		dir, dest, want string
	}{
		{"modules/publishing", "../events/README.md", "/events"},
		{"modules/publishing", "../events/README.md#editing", "/events#editing"},
		{"modules/publishing", "../events/", "/events"},
		{"modules/publishing", "#editing", "#editing"},
		{"modules/publishing", "https://example.com/x", "https://example.com/x"},
		{"modules/publishing", "schema/", sourceBlob + "modules/publishing/schema"},
		{"modules/publishing", "../../README.md#templates", sourceBlob + "README.md#templates"},
		{"modules/publishing", "../../docs/templates.md#layout", "wt-docs:/go-webthings/templates#layout"},
		{"modules/publishing", "../../docs/README.md", "wt-docs:/go-webthings"},
		{"docs", "templates.md#layout", "/go-webthings/templates#layout"},
		{"docs", "README.md", "/go-webthings"},
		{"docs", "../modules/publishing/README.md", "wt-modules:/publishing"},
		{"docs", "../frontend/view.go", sourceBlob + "frontend/view.go"},
	} {
		if got := rewriteLink(tc.dest, tc.dir); got != tc.want {
			t.Errorf("rewriteLink(%q, %q) = %q, want %q", tc.dest, tc.dir, got, tc.want)
		}
	}
}

func TestDisplay(t *testing.T) {
	if got := display("/{$}"); got != "/" {
		t.Fatalf("display = %q", got)
	}
}

func TestFormFields(t *testing.T) {
	type input struct {
		Subject string `form:"input,hidden" name:"subject" required:"true" maxLength:"64" line:"true"`
		Adults  int    `form:"input,number" name:"adults" label:"Adults" min:"0" max:"50"`
		Notes   string `form:"textarea" label:"Remarks" description:"Optional."`
		Skipped string `form:"-"`
		Plain   string
		hidden  string `form:"input,text"`
	}

	_ = input{}.hidden

	got := formFields(reflect.TypeFor[input]())
	want := []moduledoc.FormField{
		{Key: "subject", Control: "input hidden", Required: true, Rules: []string{"one line", "at most 64 characters"}},
		{Key: "adults", Control: "input number", Label: "Adults", Rules: []string{"0 to 50"}},
		{Key: "Notes", Control: "textarea", Label: "Remarks", Description: "Optional."},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("formFields:\n got %+v\nwant %+v", got, want)
	}
}
