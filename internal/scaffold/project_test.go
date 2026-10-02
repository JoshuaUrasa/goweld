package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCreateMatrix(t *testing.T) {
	for _, framework := range []string{"std", "gin", "echo"} {
		for _, database := range []string{"none", "postgres", "mysql", "sqlite"} {
			orms := []string{"sql", "gorm"}
			if database == "none" {
				orms = []string{"none"}
			}
			for _, orm := range orms {
				t.Run(framework+"/"+database+"/"+orm, func(t *testing.T) {
					root := t.TempDir()
					options := Options{Name: "app", Module: "example.com/app", Framework: framework, Database: database, ORM: orm}
					if err := Create(root, options); err != nil {
						t.Fatal(err)
					}
					project := filepath.Join(root, "app")
					got, err := Load(project)
					if err != nil || got != options {
						t.Fatalf("load: %+v, %v", got, err)
					}
					for _, kind := range []string{"model", "handler", "service"} {
						if _, err := Generate(project, kind, "User"); err != nil {
							t.Fatal(err)
						}
					}
				})
			}
		}
	}
}

func TestRejectInvalidOptionsBeforeCreatingFiles(t *testing.T) {
	for _, options := range []Options{
		{Name: "../escape"}, {Name: "app", Framework: "unknown"},
		{Name: "app", Database: "mongodb"}, {Name: "app", ORM: "gorm"},
		{Name: "app", Database: "postgres", ORM: "none"}, {Name: "app", Module: "bad\nmodule"},
		{Name: "app", Module: "example.com/../bad"},
	} {
		root := t.TempDir()
		if err := Create(root, options); err == nil {
			t.Fatalf("accepted %+v", options)
		}
		entries, err := os.ReadDir(root)
		if err != nil || len(entries) != 0 {
			t.Fatalf("unexpected files: %v %v", entries, err)
		}
	}
}

func TestNeverOverwrite(t *testing.T) {
	root := t.TempDir()
	if err := Create(root, Options{Name: "app"}); err != nil {
		t.Fatal(err)
	}
	project := filepath.Join(root, "app")
	if err := Create(root, Options{Name: "app"}); err == nil {
		t.Fatal("overwrote project")
	}
	path, err := Generate(project, "model", "User")
	if err != nil {
		t.Fatal(err)
	}
	original, _ := os.ReadFile(path)
	if _, err := Generate(project, "model", "User"); err == nil {
		t.Fatal("overwrote model")
	}
	after, _ := os.ReadFile(path)
	if string(original) != string(after) {
		t.Fatal("model changed")
	}
	for _, name := range []string{"../User", "func", "_", "123", "a/b"} {
		if _, err := Generate(project, "model", name); err == nil {
			t.Fatalf("accepted %q", name)
		}
	}
	if _, err := Generate(root, "model", "User"); err == nil {
		t.Fatal("generated outside a project")
	}
}

func TestGeneratedStandardProjectBuilds(t *testing.T) {
	root := t.TempDir()
	if err := Create(root, Options{Name: "app"}); err != nil {
		t.Fatal(err)
	}
	project := filepath.Join(root, "app")
	for _, kind := range []string{"model", "handler", "service"} {
		if _, err := Generate(project, kind, "User"); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = project
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated project: %v\n%s", err, output)
	}
}
