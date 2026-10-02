package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProgressReportsResultWithoutTerminalEscapes(t *testing.T) {
	for _, fails := range []bool{false, true} {
		var output bytes.Buffer
		ui := newUI(&output)
		err := ui.task("Installing dependencies", func() error {
			if fails {
				return errors.New("download failed")
			}
			return nil
		})
		if (err != nil) != fails {
			t.Fatalf("unexpected result %v", err)
		}
		text := output.String()
		if strings.Contains(text, "\033") {
			t.Fatalf("terminal escapes in redirected output: %q", text)
		}
		if fails && (strings.Contains(text, "✓") || !strings.Contains(text, "failed")) {
			t.Fatalf("false success: %s", text)
		}
		if !fails && !strings.Contains(text, "✓ Installing dependencies") {
			t.Fatalf("missing success: %s", text)
		}
	}
}

func TestNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	ui := newUI(os.Stdout)
	if ui.color || strings.Contains(ui.paint(cyan, "GoWeld"), "\033") {
		t.Fatal("NO_COLOR was ignored")
	}
}

func TestNewInstallRunsTidyAndPreservesProjectOnFailure(t *testing.T) {
	for _, fails := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[fails], func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			bin := filepath.Join(root, "tools")
			if err := os.Mkdir(bin, 0750); err != nil {
				t.Fatal(err)
			}
			script := "#!/bin/sh\nprintf '%s %s' \"$1\" \"$2\" > install-command\n"
			if fails {
				script += "echo 'dependency download failed' >&2\nexit 1\n"
			}
			if err := os.WriteFile(filepath.Join(bin, "go"), []byte(script), 0750); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			var output bytes.Buffer
			err := run([]string{"new", "app", "--install"}, &output, &output)
			if (err != nil) != fails {
				t.Fatalf("unexpected error %v", err)
			}
			command, readErr := os.ReadFile(filepath.Join(root, "app", "install-command"))
			if readErr != nil || string(command) != "mod tidy" {
				t.Fatalf("install command: %q %v", command, readErr)
			}
			if _, err := os.Stat(filepath.Join(root, "app", "goweld.json")); err != nil {
				t.Fatal("project lost after install", err)
			}
			if fails {
				if !strings.Contains(err.Error(), "dependency download failed") {
					t.Fatalf("lost diagnostic: %v", err)
				}
				if strings.Contains(output.String(), "is ready") {
					t.Fatal("reported ready after failure")
				}
			} else if !strings.Contains(output.String(), "is ready") {
				t.Fatal("missing project success")
			}
		})
	}
}
