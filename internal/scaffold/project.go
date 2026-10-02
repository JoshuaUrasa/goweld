package scaffold

import (
	"encoding/json"
	"fmt"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Options struct {
	Name      string `json:"name"`
	Module    string `json:"module"`
	Framework string `json:"framework"`
	Database  string `json:"database"`
	ORM       string `json:"orm"`
}

var projectName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
var modulePath = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._~/-]*$`)
var componentName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

func (o *Options) validate() error {
	if !projectName.MatchString(o.Name) {
		return fmt.Errorf("project name must start with a letter and contain only letters, numbers, hyphens or underscores")
	}
	if o.Module == "" {
		o.Module = o.Name
	}
	if !modulePath.MatchString(o.Module) {
		return fmt.Errorf("invalid module path %q", o.Module)
	}
	for _, part := range strings.Split(o.Module, "/") {
		if part == "" || part == "." || part == ".." || strings.HasSuffix(part, ".") {
			return fmt.Errorf("invalid module path %q", o.Module)
		}
	}
	if o.Framework == "" {
		o.Framework = "std"
	}
	if o.Database == "" {
		o.Database = "none"
	}
	if o.ORM == "" {
		o.ORM = "none"
		if o.Database != "none" {
			o.ORM = "sql"
		}
	}
	if o.Framework != "std" && o.Framework != "gin" && o.Framework != "echo" {
		return fmt.Errorf("unsupported framework %q: choose std, gin or echo", o.Framework)
	}
	if o.Database != "none" && o.Database != "postgres" && o.Database != "mysql" && o.Database != "sqlite" {
		return fmt.Errorf("unsupported database %q", o.Database)
	}
	if o.ORM != "none" && o.ORM != "sql" && o.ORM != "gorm" {
		return fmt.Errorf("unsupported ORM %q: choose none, sql or gorm", o.ORM)
	}
	if (o.Database == "none") != (o.ORM == "none") {
		return fmt.Errorf("select sql or gorm with a database; use orm=none only with database=none")
	}
	return nil
}

// Create validates and renders before touching the destination, and never overwrites it.
func Create(parent string, o Options) error {
	if err := o.validate(); err != nil {
		return err
	}
	files := projectFiles(o)
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}
	files["goforge.json"] = string(data) + "\n"
	for path, source := range files {
		if strings.HasSuffix(path, ".go") {
			formatted, err := format.Source([]byte(source))
			if err != nil {
				return fmt.Errorf("render %s: %w", path, err)
			}
			files[path] = string(formatted)
		}
	}
	target := filepath.Join(parent, o.Name)
	if err := os.Mkdir(target, 0750); err != nil {
		return fmt.Errorf("create project: %w", err)
	}
	complete := false
	defer func() {
		if !complete {
			_ = os.RemoveAll(target)
		}
	}()
	for path, source := range files {
		destination := filepath.Join(target, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(destination), 0750); err != nil {
			return err
		}
		if err := os.WriteFile(destination, []byte(source), 0640); err != nil {
			return err
		}
	}
	complete = true
	return nil
}

func Load(root string) (Options, error) {
	var o Options
	data, err := os.ReadFile(filepath.Join(root, "goforge.json"))
	if err != nil {
		return o, fmt.Errorf("run this command from a GoForge project root: %w", err)
	}
	if err := json.Unmarshal(data, &o); err != nil {
		return o, fmt.Errorf("invalid goforge.json: %w", err)
	}
	if err := o.validate(); err != nil {
		return o, err
	}
	return o, nil
}

// Generate adds a starter type or handler. Routes and business logic are user-owned.
func Generate(root, kind, name string) (string, error) {
	o, err := Load(root)
	if err != nil {
		return "", err
	}
	if !token.IsIdentifier(name) || !componentName.MatchString(name) {
		return "", fmt.Errorf("name must be a Go identifier, for example User")
	}
	typeName := strings.ToUpper(name[:1]) + name[1:]
	var source, dir string
	switch kind {
	case "model":
		dir = "models"
		source = "package models\ntype " + typeName + " struct { ID uint64 `json:\"id\"` }\n"
		if o.ORM == "gorm" {
			source = "package models\nimport \"gorm.io/gorm\"\ntype " + typeName + " struct { gorm.Model }\n"
		}
	case "service":
		dir = "services"
		source = "package services\n// " + typeName + "Service holds application business logic.\ntype " + typeName + "Service struct {}\n"
	case "handler":
		dir = "handlers"
		switch o.Framework {
		case "std":
			source = "package handlers\nimport \"net/http\"\nfunc " + typeName + "(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) }\n"
		case "gin":
			source = "package handlers\nimport (\"net/http\"; \"github.com/gin-gonic/gin\")\nfunc " + typeName + "(c *gin.Context) { c.Status(http.StatusNotImplemented) }\n"
		case "echo":
			source = "package handlers\nimport (\"net/http\"; \"github.com/labstack/echo/v5\")\nfunc " + typeName + "(c *echo.Context) error { return c.NoContent(http.StatusNotImplemented) }\n"
		}
	default:
		return "", fmt.Errorf("unsupported generator %q: choose model, handler or service", kind)
	}
	data, err := format.Source([]byte(source))
	if err != nil {
		return "", err
	}
	path := filepath.Join(root, "internal", dir, strings.ToLower(name)+".go")
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return "", err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
	if err != nil {
		return "", err
	}
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil {
		return "", writeErr
	}
	return path, closeErr
}
