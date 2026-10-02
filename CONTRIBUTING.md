# Contributing to GoWeld

Bug reports, documentation improvements and new framework/database templates are welcome.

## Local setup

Requires Go 1.26.5 or newer.

```sh
git clone https://github.com/JoshuaUrasa/goweld.git
cd goweld
go build -o bin/goweld ./cmd/goweld
./bin/goweld help
```

## Before submitting a pull request

1. Create a branch for your change.
2. Keep the change focused and explain the user-facing behavior in your PR.
3. Format Go files with `gofmt`.
4. Run `go test ./...` and `go vet ./...`.
5. Update the README when commands or supported options change.

Framework/database templates live in `internal/scaffold/templates.go`. Option validation and component generators live in `internal/scaffold/project.go`. When adding a template, verify that a generated project compiles and document any setup requirements.

For bug reports, include your operating system, `go version`, command used, expected result and actual error. Remove credentials from any output you share.


Contributions are provided under the project's [MIT license](LICENSE).
