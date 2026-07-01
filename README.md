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

The deployable templates and assets live under `deploy` and are loaded from the
filesystem at runtime. When running from the repository root, the app uses
`deploy`; when running the built binary from inside `deploy`, it uses the current
directory. Set `ASSET_DIR` to override this.

The go-partial documentation is copied from `donseba/go-partial/examples/docs`
into `deploy/elements/go_partial` and rendered with `go-partial` itself.
The shared docs-family stylesheet source lives at `deploy/static/site.tailwind.css`;
the generated output is `deploy/static/site.css` and is served as
`/assets/site.css` for each docs/showcase host.
