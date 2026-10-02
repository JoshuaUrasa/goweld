package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JoshuaUrasa/goweld/internal/scaffold"
)

func TestInteractiveProjectChoices(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		args                     []string
		answers                  string
		framework, database, orm string
	}{
		{"all choices", []string{"new", "--interactive"}, "app\n2\n2\n2\n", "gin", "postgres", "gorm"},
		{"retry invalid choice", []string{"new", "app", "--interactive"}, "invalid\n3\n4\n1\n", "echo", "sqlite", "sql"},
		{"defaults without ORM", []string{"new", "app", "--interactive"}, "\n\n", "std", "none", "none"},
		{"respect explicit flags", []string{"new", "app", "--interactive", "--framework", "gin", "--database", "mysql", "--orm", "gorm"}, "", "gin", "mysql", "gorm"},
		{"prompt missing ORM only", []string{"new", "app", "--interactive", "--framework", "echo", "--database", "sqlite"}, "gorm\n", "echo", "sqlite", "gorm"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			var output bytes.Buffer
			if err := runWithInput(tc.args, strings.NewReader(tc.answers), &output, &output); err != nil {
				t.Fatal(err)
			}
			options, err := scaffold.Load(filepath.Join(root, "app"))
			if err != nil {
				t.Fatal(err)
			}
			if options.Framework != tc.framework || options.Database != tc.database || options.ORM != tc.orm {
				t.Fatalf("wrong choices: %+v", options)
			}
			if tc.database == "none" && strings.Contains(output.String(), "Which ORM") {
				t.Fatal("asked for ORM without a database")
			}
		})
	}
}

func TestInterruptedInteractiveSetupDoesNotCreateProject(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	var output bytes.Buffer
	err := runWithInput([]string{"new", "app", "--interactive"}, strings.NewReader("2\n"), &output, &output)
	if err == nil {
		t.Fatal("accepted incomplete input")
	}
	if _, err := os.Stat(filepath.Join(root, "app")); !os.IsNotExist(err) {
		t.Fatalf("created incomplete project: %v", err)
	}
}

func TestNonInteractiveDefaults(t *testing.T) {
	t.Chdir(t.TempDir())
	var output bytes.Buffer
	if err := runWithInput([]string{"new", "app", "--no-interactive"}, strings.NewReader(""), &output, &output); err != nil {
		t.Fatal(err)
	}
	options, err := scaffold.Load("app")
	if err != nil {
		t.Fatal(err)
	}
	if options.Framework != "std" || options.Database != "none" || options.ORM != "none" {
		t.Fatalf("wrong defaults: %+v", options)
	}
	if strings.Contains(output.String(), "Which framework") {
		t.Fatal("prompted in noninteractive mode")
	}
}
