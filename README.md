# GoWeld

Create Go projects with an organized folder structure and add components from the command line.

GoWeld is an open-source Go project generator. This is an early MVP with framework/database selection, component generators and terminal progress indicators.

## Install

Requires Go 1.26.5 or newer.

```sh
go install github.com/JoshuaUrasa/goweld/cmd/goweld@latest
```

Make sure the Go binary directory (`go env GOPATH` + `/bin`, unless GOBIN is configured) is on your PATH.

To build from a local checkout instead:

```sh
git clone https://github.com/JoshuaUrasa/goweld.git
cd goweld
go install ./cmd/goweld
```

## Upgrading from GoForge

The CLI is now named `goweld`. New projects use `goweld.json`; existing projects with `goforge.json` also work with GoWeld without migration. When both files exist, `goweld.json` takes precedence.

## Create a project

On a terminal, start the interactive setup:

```sh
goweld new
# Or provide the name and choose the stack interactively:
goweld new shop
```

GoWeld asks for the project name (if missing), framework, database and data access. Enter a number or option name; Enter accepts the default. ORM selection appears when a database is selected. Options already provided as flags are not asked again.

Use `--interactive` to force prompts with redirected input, or `--no-interactive` to use defaults without questions. Scripts and redirected input default to noninteractive behavior.

You can also provide all stack choices directly:

```sh
goweld new shop --module github.com/yourname/shop --framework gin --database postgres --orm gorm --install
cd shop
```

Options:

| Flag | Choices | Default |
| --- | --- | --- |
| `--framework` | `std` (net/http), `gin`, `echo` (v5) | `std` |
| `--database` | `none`, `postgres`, `mysql`, `sqlite` | `none` |
| `--orm` | `none`, `sql` (database/sql), `gorm` | `sql` when database selected; otherwise `none` |
| `--module` | Go module path | project name |
| `--interactive` | Prompt for missing choices | Automatic on terminals |
| `--no-interactive` | Skip questions and use defaults | Off |

Place the project name before the flags. The listed frameworks and data-access options are supported today; adding another framework requires a template in `internal/scaffold/templates.go` and validation in `project.go`.

`--install` installs dependencies after creating the project. You can also run `goweld install` inside an existing generated project. Both use `go mod tidy`, which records resolved versions in go.mod and go.sum. Commit both files. If installation fails, the generated project stays available so you can retry.

The CLI displays a cyan/purple banner, an animated spinner while installing/building, green success messages and red errors. Progress reflects real command execution. Redirected output uses plain text; set `NO_COLOR=1` to disable colors or `TERM=dumb` to disable terminal styling and animation. GoWeld itself uses only the standard library.

## Configure and run

Generated projects include `.env.example`. Configuration is read from the process environment; `.env` files are not loaded automatically.

```sh
cp .env.example .env
# Edit DATABASE_DSN to match your database, then export in a POSIX shell:
set -a
. ./.env
set +a
goweld run
```

For PostgreSQL/MySQL, start the database server and create your database before running the application. SQLite with GORM requires CGO and a C compiler. SQLite with `sql` uses a pure-Go driver.

The server provides `GET /health` and checks the selected database connection on startup.

## Generate components

Run from the generated project root:

```sh
goweld generate model User
goweld generate handler User
goweld generate service User
goweld build
goweld version
```

Generators create starter files. Add business logic and register routes in `internal/server/router.go`; new handlers return HTTP 501 until implemented. GORM models embed `gorm.Model`. Automatic CRUD and migrations are not included yet.

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
└── goweld.json
```

Existing project folders and component files are never overwritten. `build` writes the executable to `bin/server`.

## Development

```sh
go test ./...
go vet ./...
go build ./cmd/goweld
```

Framework and database references: [Gin quickstart](https://gin-gonic.com/en/docs/quickstart/), [Echo quickstart](https://echo.labstack.com/guide/quickstart/), [GORM connections](https://gorm.io/docs/connecting_to_the_database.html).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for setup, checks and pull request guidance.

## License

GoWeld is licensed under the [MIT License](LICENSE).
