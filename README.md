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
- `http://showcase.rocketweb.nl:8080/go-partial`
- `https://docs.gowebthings.com/go-partial`
- `https://docs.gowebthings.com/go-router`
- `https://showcase.gowebthings.com/go-partial`

The same router currently supports:

- `go-partial`
- `go-clue`
- `go-router`
- `go-form` — fields, metadata, submissions, validation, CSRF, themes, and translations
- `go-importmap` — packages, providers, storage, rendering, serving, and deployment
- `go-translator` — catalogues, locales, template helpers, plurals, contexts, extraction, and editing

Each component has an overview, topic pages, an API reference, and sidebar navigation.
The form, importmap, and translator examples use released versions v2.3.0, v1.4.0, and
v1.4.0 respectively. Full-page and HTMX section requests are covered by the route tests.

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
The shared docs-family stylesheet source lives at `deploy/website/docs/tailwind/main.css`;
the generated output is `deploy/website/docs/assets/css/styles.css` and is served as
`/assets/css/styles.css` for each docs/showcase host.

The main website has its own deploy files under `deploy/website/main`. Its template is
`deploy/website/main/templates/page.gohtml`, its copied image assets live in
`deploy/website/main/assets/img`, and its stylesheet source/output live at
`deploy/website/main/tailwind/main.css` and `deploy/website/main/assets/css/styles.css`.
