# go-webthings-docs
Go Webthings Documentation

## Local routing proof

Run the docs server:

```bash
go run ./cmd/website
```

Routes are selected by host:

- `http://docs.rocketweb.nl:8080/go-partial`
- `http://docs.rocketweb.nl:8080/go-partial/rendering`
- `http://docs.rocketweb.nl:8080/go-clue`
- `http://docs.rocketweb.nl:8080/go-clue/install`
- `http://docs.rocketweb.nl:8080/go-router`
- `http://docs.rocketweb.nl:8080/go-router/hosts`
- `http://docs.rocketweb.nl:8080/go-form/submissions`
- `http://docs.rocketweb.nl:8080/go-importmap/serving`
- `http://docs.rocketweb.nl:8080/go-translator/extraction`
- `http://docs.rocketweb.nl:8080/go-webthings`
- `http://showcase.rocketweb.nl:8080/go-partial`
- `http://modules.rocketweb.nl:8080/publishing`
- `http://modules.localhost:8080/publishing` (no DNS needed)
- `https://docs.gowebthings.com/go-partial`
- `https://docs.gowebthings.com/go-router`
- `https://docs.gowebthings.com/go-webthings`
- `https://showcase.gowebthings.com/go-partial`
- `https://modules.gowebthings.com/publishing`

The same router currently supports:

- `go-partial`
- `go-clue`
- `go-router`
- `go-form` — fields, metadata, submissions, validation, CSRF, themes, and translations
- `go-importmap` — vendoring, packages, providers, storage, rendering, serving, and deployment
- `go-translator` — catalogues, locales, template helpers, plurals, contexts, extraction, the extract package, and editing
- `go-webthings` — the user guides, generated from the `docs/` folder of go-webthings

go-partial, go-form, go-importmap and go-translator each have an overview, topic pages, an API reference,
and sidebar navigation. The form, importmap, and translator examples use versions v2.6.0,
v1.5.0, and v2.0.0 (module `github.com/donseba/go-translator/v2`) respectively. Full-page
and HTMX section requests are covered by the route tests.

Check the route behavior with:

```bash
go test ./...
go vet ./...
golangci-lint run
go tool go-clue templates .
```

Build the shared docs stylesheet from its Tailwind source with:

```bash
task build-css
```

The deployable website lives under `deploy/website` and is split into deploy sections:

- `deploy/website/docs`
- `deploy/website/main`
- `deploy/website/showcase`

The app loads deploy files from the filesystem at runtime. When running from the repository
root, it uses `deploy/website/docs`; when running the built binary from `deploy/website`,
it uses the `docs` directory next to the executable. Set `ASSET_DIR` to override this.

The element documentation templates live under `deploy/website/docs/templates/go_partial`,
`go_clue`, `go_router`, `go_form`, `go_importmap`, and `go_translator`;
shared shell templates live under `deploy/website/docs/templates/general`.
The go-webthings guides and the module pages use `deploy/website/docs/templates/guides`
and `deploy/website/docs/templates/modules`.
The shared docs-family stylesheet source lives at `deploy/website/docs/tailwind/main.css`;
the generated output is `deploy/website/docs/assets/css/styles.css` and is served as
`/assets/css/styles.css` for each docs/showcase host.

The main website has its own deploy files under `deploy/website/main`. Its templates are in
`deploy/website/main/templates`: `layout.gohtml` around `page.gohtml` (the home page),
`components.gohtml`, `modules.gohtml` and `generate.gohtml`. Its copied image assets live in
`deploy/website/main/assets/img`, and its stylesheet source/output live at
`deploy/website/main/tailwind/main.css` and `deploy/website/main/assets/css/styles.css`.

## Modules

`gowebthings.com/modules` lists the modules bundled with go-webthings and
`modules.gowebthings.com/<module>` documents each one: its guide, then
Configuration, Templates, Database and Go API pages.

The pages are rendered from `deploy/website/docs/modules/<module>.json`. Those
files are generated, not edited: `tools/moduledocs` (its own Go module, with a
`replace` to a go-webthings checkout next to this repository) starts every
module on a throwaway SQLite database and reads its code, then writes the JSON.
The guide is the module's `README.md` in go-webthings; everything else comes
from the code: constructor and options, admin sections, settings, addresses,
template functions, templates and their data, e-mail templates, migrations,
tables and the Go API.

After changing a module, its README or a guide in the `docs/` folder of
go-webthings, regenerate and commit the JSON:

```bash
task modules
```

It also writes `deploy/website/docs/guides/go-webthings.json`, the guides at
`docs.gowebthings.com/go-webthings`.

Use `WEBTHINGS=/path/to/go-webthings task modules` for a checkout elsewhere.
The deployed website itself does not depend on go-webthings.

