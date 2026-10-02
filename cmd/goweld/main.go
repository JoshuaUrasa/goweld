package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/JoshuaUrasa/goweld/internal/scaffold"
)

const version = "0.2.0"

const help = `GoWeld — scaffold Go applications

Usage:
  goweld new [NAME] [--interactive | --no-interactive] [--install]
                   [--module PATH] [--framework std|gin|echo]
                   [--database none|postgres|mysql|sqlite] [--orm none|sql|gorm]
  goweld generate model|handler|service NAME
  goweld install
  goweld run
  goweld build
  goweld version

Run generate, install, run and build from the generated project root.
Defaults: framework=std, database=none, orm=none (sql when a database is selected).
On a terminal, new prompts for a missing name, framework, database and ORM.
`

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		newUI(os.Stderr).failure(err.Error())
		os.Exit(1)
	}
}

func run(args []string, out, stderr io.Writer) error {
	return runWithInput(args, os.Stdin, out, stderr)
}

func runWithInput(args []string, in io.Reader, out, stderr io.Writer) error {
	ui := newUI(out)
	if len(args) == 0 {
		ui.banner()
		fmt.Fprint(out, help)
		return nil
	}
	switch args[0] {
	case "help", "--help", "-h":
		ui.banner()
		fmt.Fprint(out, help)
	case "version":
		if len(args) != 1 {
			return errors.New("usage: goweld version")
		}
		fmt.Fprintln(out, version)
	case "new":
		if len(args) == 2 && (args[1] == "--help" || args[1] == "-h") {
			ui.banner()
			fmt.Fprint(out, help)
			return nil
		}
		options := scaffold.Options{}
		remaining := args[1:]
		if len(remaining) > 0 && !strings.HasPrefix(remaining[0], "-") {
			options.Name = remaining[0]
			remaining = remaining[1:]
		}
		flags := flag.NewFlagSet("new", flag.ContinueOnError)
		flags.SetOutput(stderr)
		install := flags.Bool("install", false, "install dependencies after creating the project")
		interactive := flags.Bool("interactive", false, "prompt for missing project choices")
		noInteractive := flags.Bool("no-interactive", false, "use defaults without prompting")
		flags.StringVar(&options.Module, "module", "", "Go module path")
		flags.StringVar(&options.Framework, "framework", "std", "std, gin or echo")
		flags.StringVar(&options.Database, "database", "none", "none, postgres, mysql or sqlite")
		flags.StringVar(&options.ORM, "orm", "", "none, sql or gorm")
		if err := flags.Parse(remaining); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		if flags.NArg() != 0 {
			return errors.New("provide exactly one project name before the flags")
		}
		if *interactive && *noInteractive {
			return errors.New("choose either --interactive or --no-interactive")
		}
		shouldPrompt := *interactive || (!*noInteractive && terminalFile(in) && terminalFile(out))
		if options.Name == "" && !shouldPrompt {
			return errors.New("usage: goweld new NAME [options]; use --interactive to choose a project")
		}
		ui.banner()
		if shouldPrompt {
			provided := make(map[string]bool)
			flags.Visit(func(f *flag.Flag) { provided[f.Name] = true })
			if err := promptProject(in, ui, &options, provided); err != nil {
				return err
			}
		}
		ui.info("Project", options.Name)
		ui.info("Stack  ", stackSummary(options.Framework, options.Database, options.ORM))
		fmt.Fprintln(out)
		if err := ui.task("Creating project structure", func() error { return scaffold.Create(".", options) }); err != nil {
			return err
		}
		if *install {
			if err := installDependencies(ui, options.Name); err != nil {
				return fmt.Errorf("project %s was created; retry with cd %s && goweld install: %w", options.Name, options.Name, err)
			}
		}
		ui.success("Project " + options.Name + " is ready")
		next := []string{"cd " + options.Name}
		if !*install {
			next = append(next, "goweld install")
		}
		next = append(next, "goweld run")
		ui.next(next...)
	case "generate":
		if len(args) != 3 {
			return errors.New("usage: goweld generate model|handler|service NAME")
		}
		var path string
		ui.banner()
		if err := ui.task("Generating "+args[1]+" "+args[2], func() error {
			var err error
			path, err = scaffold.Generate(".", args[1], args[2])
			return err
		}); err != nil {
			return err
		}
		ui.info("Created", path)
	case "install":
		if len(args) != 1 {
			return errors.New("usage: goweld install")
		}
		if _, err := scaffold.Load("."); err != nil {
			return err
		}
		ui.banner()
		if err := installDependencies(ui, "."); err != nil {
			return err
		}
		ui.next("goweld run")
	case "run", "build":
		if len(args) != 1 {
			return fmt.Errorf("usage: goweld %s", args[0])
		}
		if _, err := scaffold.Load("."); err != nil {
			return err
		}
		ui.banner()
		goArgs := []string{"run", "./cmd/server"}
		if args[0] == "build" {
			if err := os.MkdirAll("bin", 0750); err != nil {
				return err
			}
			goArgs = []string{"build", "-o", "bin/server", "./cmd/server"}
		}
		command := exec.Command("go", goArgs...)
		command.Stdin, command.Stdout, command.Stderr = in, out, stderr
		if args[0] == "build" {
			if err := ui.task("Building application", func() error { return capturedCommand(command) }); err != nil {
				return err
			}
			ui.info("Binary", filepath.Join("bin", "server"))
			return nil
		}
		ui.info("Starting", "development server")
		fmt.Fprintln(out)
		return command.Run()
	default:
		return fmt.Errorf("unknown command %q; run goweld help", args[0])
	}
	return nil
}

func installDependencies(ui terminalUI, dir string) error {
	command := exec.Command("go", "mod", "tidy")
	command.Dir = dir
	return ui.task("Installing dependencies", func() error { return capturedCommand(command) })
}

// Capture Go's output so dependency logs do not collide with the progress line.
// On failure, return the original diagnostics for display after the renderer stops.
func capturedCommand(command *exec.Cmd) error {
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	if err := command.Run(); err != nil {
		return &commandError{cause: err, output: output.String()}
	}
	return nil
}

type commandError struct {
	cause  error
	output string
}

func (e *commandError) Error() string { return e.cause.Error() + "\n" + e.output }
func (e *commandError) Unwrap() error { return e.cause }
