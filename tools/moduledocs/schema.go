package main

import (
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"

	"donseba/go-webthings-docs/internal/moduledoc"
)

// flavours are the databases in the order the website shows them; tables
// are read from the first.
var flavours = []string{"postgres", "sqlite", "mysql"}

var migrationName = regexp.MustCompile(`^(\d{14})(?:_([a-z0-9_]+))?\.(sqlite|postgres|mysql)\.sql$`)

// schema reads a module's migrations, and the tables they create.
func schema(fsys fs.FS) ([]moduledoc.Migration, []moduledoc.Table, error) {
	if fsys == nil {
		return nil, nil, nil
	}

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, nil, err
	}

	byVersion := map[string]*moduledoc.Migration{}

	for _, entry := range entries {
		match := migrationName.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}

		body, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			return nil, nil, err
		}

		migration, ok := byVersion[match[1]]
		if !ok {
			migration = &moduledoc.Migration{
				Version:     match[1],
				Description: strings.ReplaceAll(match[2], "_", " "),
			}
			byVersion[match[1]] = migration
		}

		migration.SQL = append(migration.SQL, moduledoc.Flavour{
			Name: match[3],
			File: entry.Name(),
			SQL:  strings.TrimSpace(string(body)),
		})
	}

	migrations := make([]moduledoc.Migration, 0, len(byVersion))

	for _, migration := range byVersion {
		sort.Slice(migration.SQL, func(i, j int) bool {
			return flavourOrder(migration.SQL[i].Name) < flavourOrder(migration.SQL[j].Name)
		})

		migrations = append(migrations, *migration)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	var tables []moduledoc.Table

	for _, migration := range migrations {
		if len(migration.SQL) == 0 {
			continue
		}

		tables, err = apply(tables, migration.SQL[0].SQL)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", migration.SQL[0].File, err)
		}
	}

	return migrations, tables, nil
}

func flavourOrder(name string) int {
	for i, flavour := range flavours {
		if flavour == name {
			return i
		}
	}

	return len(flavours)
}

var (
	createTable = regexp.MustCompile(`(?is)^CREATE TABLE (?:IF NOT EXISTS )?([a-z0-9_]+)\s*\((.*)\)[^)]*$`)
	createIndex = regexp.MustCompile(`(?is)^CREATE (UNIQUE )?INDEX (?:IF NOT EXISTS )?([a-z0-9_]+) ON ([a-z0-9_]+)\s*(\(.*\))`)
	addColumn   = regexp.MustCompile(`(?is)^ALTER TABLE ([a-z0-9_]+) ADD (?:COLUMN )?(.*)$`)
	dropTable   = regexp.MustCompile(`(?is)^DROP TABLE (?:IF EXISTS )?([a-z0-9_]+)`)
)

// apply reads the statements of one migration into tables.
func apply(tables []moduledoc.Table, sql string) ([]moduledoc.Table, error) {
	for _, statement := range statements(sql) {
		text := strings.Join(strings.Fields(statement.text), " ")

		if m := createTable.FindStringSubmatch(statement.text); m != nil {
			table := moduledoc.Table{Name: m[1], Note: statement.note}

			for _, item := range splitTopLevel(m[2]) {
				item = strings.Join(strings.Fields(item), " ")

				if isConstraint(item) {
					table.Indexes = append(table.Indexes, item)

					continue
				}

				table.Columns = append(table.Columns, column(item))
			}

			tables = append(tables, table)

			continue
		}

		if m := createIndex.FindStringSubmatch(text); m != nil {
			for i := range tables {
				if tables[i].Name == m[3] {
					tables[i].Indexes = append(tables[i].Indexes, strings.TrimSpace(m[1]+"INDEX "+m[2]+" "+m[4]))
				}
			}

			continue
		}

		if m := addColumn.FindStringSubmatch(text); m != nil {
			for i := range tables {
				if tables[i].Name == m[1] {
					tables[i].Columns = append(tables[i].Columns, column(m[2]))
				}
			}

			continue
		}

		if m := dropTable.FindStringSubmatch(text); m != nil {
			for i := range tables {
				if tables[i].Name == m[1] {
					tables = append(tables[:i], tables[i+1:]...)

					break
				}
			}
		}
	}

	return tables, nil
}

type statement struct {
	text string
	// note is the comment right above the statement.
	note string
}

// statements splits SQL at semicolons outside quotes and drops comments,
// keeping the comment above each statement as its note.
func statements(sql string) []statement {
	var (
		list    []statement
		current strings.Builder
		note    []string
	)

	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "--") {
			if strings.TrimSpace(current.String()) == "" {
				note = append(note, strings.TrimSpace(strings.TrimPrefix(trimmed, "--")))
			}

			continue
		}

		line = withoutComment(line)

		if trimmed == "" && strings.TrimSpace(current.String()) == "" {
			note = nil

			continue
		}

		current.WriteString(line + "\n")

		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			text := strings.TrimSuffix(strings.TrimSpace(current.String()), ";")
			list = append(list, statement{text: text, note: strings.Join(note, " ")})
			current.Reset()

			note = nil
		}
	}

	if text := strings.TrimSpace(current.String()); text != "" {
		list = append(list, statement{text: text, note: strings.Join(note, " ")})
	}

	return list
}

// withoutComment drops a -- comment that is not inside quotes.
func withoutComment(line string) string {
	quoted := false

	for i := 0; i < len(line); i++ {
		switch {
		case line[i] == '\'':
			quoted = !quoted
		case !quoted && strings.HasPrefix(line[i:], "--"):
			return strings.TrimRight(line[:i], " \t")
		}
	}

	return line
}

// splitTopLevel splits a table body at commas outside parentheses and quotes.
func splitTopLevel(body string) []string {
	var (
		parts  []string
		depth  int
		quoted bool
		start  int
	)

	for i, r := range body {
		switch {
		case r == '\'':
			quoted = !quoted
		case quoted:
		case r == '(':
			depth++
		case r == ')':
			depth--
		case r == ',' && depth == 0:
			parts = append(parts, body[start:i])
			start = i + 1
		}
	}

	parts = append(parts, body[start:])

	var list []string

	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			list = append(list, strings.TrimSpace(part))
		}
	}

	return list
}

func isConstraint(item string) bool {
	upper := strings.ToUpper(item)

	for _, prefix := range []string{"PRIMARY KEY", "UNIQUE", "FOREIGN KEY", "CONSTRAINT", "CHECK", "KEY ", "INDEX "} {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}

	return false
}

// keywords end a column's type.
var keywords = map[string]bool{
	"NOT": true, "NULL": true, "DEFAULT": true, "PRIMARY": true, "REFERENCES": true,
	"UNIQUE": true, "CHECK": true, "GENERATED": true, "COLLATE": true,
	"AUTO_INCREMENT": true, "AUTOINCREMENT": true, "CONSTRAINT": true, "ON": true,
}

// column reads "name TYPE constraints".
func column(item string) moduledoc.Column {
	words := strings.Fields(item)
	if len(words) == 0 {
		return moduledoc.Column{}
	}

	c := moduledoc.Column{Name: strings.Trim(words[0], "`\"")}

	i := 1
	for ; i < len(words) && !keywords[strings.ToUpper(words[i])]; i++ {
		if c.Type != "" {
			c.Type += " "
		}

		c.Type += words[i]
	}

	c.Notes = strings.Join(words[i:], " ")

	return c
}
