# GoForge

CLI ya kutengeneza Go projects zenye folder structure na starter code, na kuongeza components kwa commands.

GoForge is an open-source Go project generator. This is an early MVP with framework/database selection, component generators and terminal progress indicators.

## Install

Requires Go 1.26.5 or newer.

```sh
go install github.com/JoshuaUrasa/goforge/cmd/goforge@latest
```

Hakikisha Go binary directory (`go env GOPATH` + `/bin`) ipo kwenye PATH.

To build from a local checkout instead:

```sh
git clone https://github.com/JoshuaUrasa/goforge.git
cd goforge
go install ./cmd/goforge
```

## Create a project

```sh
goforge new shop --module github.com/yourname/shop --framework gin --database postgres --orm gorm --install
cd shop
```

Options:

| Flag | Choices | Default |
| --- | --- | --- |
| `--framework` | `std` (net/http), `gin`, `echo` (v5) | `std` |
| `--database` | `none`, `postgres`, `mysql`, `sqlite` | `none` |
| `--orm` | `none`, `sql` (database/sql), `gorm` | `sql` when database selected; otherwise `none` |
| `--module` | Go module path | project name |

Weka project name kabla ya flags. Framework na data-access choices hizi ndizo zinazosapotiwa kwa sasa; framework nyingine zinahitaji template mpya kwenye `internal/scaffold/templates.go` na validation kwenye `project.go`.

`--install` installs dependencies after creating the project. You can also run `goforge install` inside an existing generated project. Both use `go mod tidy`, which records resolved versions in go.mod and go.sum. Commit both files. If installation fails, the generated project stays available so you can retry.

The CLI displays a cyan/purple banner, an animated spinner while installing/building, green success messages and red errors. Progress reflects real command execution. Redirected output uses plain text; set `NO_COLOR=1` to disable colors or `TERM=dumb` to disable terminal styling and animation. GoForge itself uses only the standard library.

## Configure and run

Project ina `.env.example`; configuration inasomwa kutoka environment, si `.env` automatically.

```sh
cp .env.example .env
# Edit DATABASE_DSN to match your database, then export in a POSIX shell:
set -a
. ./.env
set +a
goforge run
```

Kwa PostgreSQL/MySQL, anza database server na tengeneza database kabla ya ku-run. SQLite na GORM inahitaji CGO na C compiler. SQLite na `sql` inatumia pure-Go driver.

Server ina `GET /health`. Connection ya database inafunguliwa na kuangaliwa wakati wa startup.

## Generate components

Run from the generated project root:

```sh
goforge generate model User
goforge generate handler User
goforge generate service User
goforge build
goforge version
```

Generators hutengeneza starter files. Ongeza business logic na register routes kwenye `internal/server/router.go`; handler mpya hurudisha HTTP 501 mpaka uiimplement. Model ya GORM inatumia `gorm.Model`. Hakuna automatic CRUD au migrations bado.

```text
shop/
├── cmd/server/main.go
├── internal/
│   ├── config/
│   ├── database/        # when selected
│   ├── handlers/
│   ├── models/
│   ├── repositories/
│   ├── server/
│   └── services/
├── migrations/
├── .env.example
├── go.mod
└── goforge.json
```

Existing project folders and component files are never overwritten. `build` writes the executable to `bin/server`.

## Development

```sh
go test ./...
go vet ./...
go build ./cmd/goforge
```

Framework and database references: [Gin quickstart](https://gin-gonic.com/en/docs/quickstart/), [Echo quickstart](https://echo.labstack.com/guide/quickstart/), [GORM connections](https://gorm.io/docs/connecting_to_the_database.html).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for setup, checks and pull request guidance.

## License

GoForge is licensed under the [MIT License](LICENSE).
