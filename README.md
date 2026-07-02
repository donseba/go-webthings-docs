# go-webthings-docs
Go Webthings Documentation

## Local routing proof

Run the docs server:

```bash
go run ./cmd/docs
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

The deployable docs templates and assets live under `deploy/docs` and are loaded from the
filesystem at runtime. When running from the repository root, the app uses
`deploy/docs`; when running the built binary from inside `deploy/docs`, it uses the current
directory. Set `ASSET_DIR` to override this.

The element documentation templates live under `deploy/docs/templates/go_partial`,
`deploy/docs/templates/go_doc`, and `deploy/docs/templates/go_router`; shared shell
templates live under `deploy/docs/templates/general`.
The shared docs-family stylesheet source lives at `deploy/docs/tailwind/main.css`;
the generated output is `deploy/docs/assets/css/styles.css` and is served as
`/assets/css/styles.css` for each docs/showcase host.
