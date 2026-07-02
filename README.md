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
- `http://docs.rocketweb.nl:8080/go-docs`
- `http://docs.rocketweb.nl:8080/go-docs/install`
- `http://docs.rocketweb.nl:8080/go-router`
- `http://docs.rocketweb.nl:8080/go-router/hosts`
- `http://showcase.rocketweb.nl:8080/go-partial`
- `https://docs.go-webthings.com/go-partial`
- `https://docs.go-webthings.com/go-router`
- `https://showcase.go-webthings.com/go-partial`

The same router currently supports:

- `go-partial`
- `go-docs`
- `go-router`

Check the route behavior with:

```bash
go test ./...
go tool go-doc templates .
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
`deploy/website/docs/templates/go_doc`, and `deploy/website/docs/templates/go_router`;
shared shell templates live under `deploy/website/docs/templates/general`.
The shared docs-family stylesheet source lives at `deploy/website/docs/tailwind/main.css`;
the generated output is `deploy/website/docs/assets/css/styles.css` and is served as
`/assets/css/styles.css` for each docs/showcase host.

The main website has its own deploy files under `deploy/website/main`. Its template is
`deploy/website/main/templates/page.gohtml`, its copied image assets live in
`deploy/website/main/assets/img`, and its stylesheet source/output live at
`deploy/website/main/tailwind/main.css` and `deploy/website/main/assets/css/styles.css`.
