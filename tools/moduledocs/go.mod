module donseba/go-webthings-docs/tools/moduledocs

go 1.27.3

require (
	donseba/go-webthings-docs v0.0.0
	github.com/donseba/go-router v0.12.0
	github.com/donseba/go-webthings v0.0.0
	github.com/yuin/goldmark v1.8.6
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/donseba/go-form/v2 v2.5.0 // indirect
	github.com/donseba/go-importmap v1.3.0 // indirect
	github.com/donseba/go-partial v0.15.0 // indirect
	github.com/donseba/go-translator/v2 v2.0.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/ebitengine/purego v0.10.1 // indirect
	github.com/gen2brain/webp v0.6.4 // indirect
	github.com/go-sql-driver/mysql v1.10.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.11.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/leonelquinteros/gotext v1.7.2-0.20250311125224-bdb582003488 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mattn/go-sqlite3 v1.14.52 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/image v0.46.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	modernc.org/libc v1.77.1 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
	modernc.org/sqlite v1.60.1 // indirect
	rsc.io/qr v0.2.0 // indirect
)

replace (
	donseba/go-webthings-docs => ../..
	// go-webthings needs unreleased branches of go-form, go-importmap,
	// go-router and go-translator (feature/webthings-integration) and
	// go-partial (feature/exp-ext); remove these lines once they are
	// released and go-webthings requires them.
	github.com/donseba/go-form/v2 => ../../../go-form
	github.com/donseba/go-importmap => ../../../go-importmap
	github.com/donseba/go-partial => ../../../go-partial
	github.com/donseba/go-router => ../../../go-router
	github.com/donseba/go-translator/v2 => ../../../go-translator
	github.com/donseba/go-webthings => ../../../go-webthings
)
