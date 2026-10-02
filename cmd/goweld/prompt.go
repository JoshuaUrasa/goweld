package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/JoshuaUrasa/goweld/internal/scaffold"
)

func terminalFile(value any) bool {
	file, ok := value.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

type projectPrompt struct {
	reader *bufio.Reader
	ui     terminalUI
}

func (p projectPrompt) answer(label string) (string, error) {
	fmt.Fprint(p.ui.out, "  "+p.ui.paint(cyan, label)+" ")
	line, err := p.reader.ReadString('\n')
	if err != nil && !(errors.Is(err, io.EOF) && len(line) > 0) {
		return "", fmt.Errorf("project setup stopped while reading %s: %w", label, err)
	}
	return strings.TrimSpace(line), nil
}

func (p projectPrompt) choose(label string, choices, labels []string, defaultValue string) (string, error) {
	fmt.Fprintln(p.ui.out)
	fmt.Fprintln(p.ui.out, "  "+p.ui.paint(bold, label))
	for index, choice := range choices {
		suffix := ""
		if choice == defaultValue {
			suffix = " (default)"
		}
		fmt.Fprintf(p.ui.out, "    %s %s%s\n", p.ui.paint(purple, strconv.Itoa(index+1)+")"), labels[index], p.ui.paint(dim, suffix))
	}
	for {
		value, err := p.answer("Choose a number or name [" + defaultValue + "]:")
		if err != nil {
			return "", err
		}
		if value == "" {
			return defaultValue, nil
		}
		value = strings.ToLower(value)
		for index, choice := range choices {
			if value == choice || value == strconv.Itoa(index+1) {
				return choice, nil
			}
		}
		p.ui.info("Try again:", "choose one of the listed options.")
	}
}

func promptProject(in io.Reader, ui terminalUI, o *scaffold.Options, provided map[string]bool) error {
	p := projectPrompt{reader: bufio.NewReader(in), ui: ui}
	for o.Name == "" {
		value, err := p.answer("Project name:")
		if err != nil {
			return err
		}
		o.Name = value
		if value == "" {
			ui.info("Required:", "enter a project name, for example myapi.")
		}
	}
	var err error
	if !provided["framework"] {
		o.Framework, err = p.choose("Which framework?", []string{"std", "gin", "echo"}, []string{"net/http (standard library)", "Gin", "Echo"}, "std")
		if err != nil {
			return err
		}
	}
	if !provided["database"] {
		o.Database, err = p.choose("Which database?", []string{"none", "postgres", "mysql", "sqlite"}, []string{"No database", "PostgreSQL", "MySQL", "SQLite"}, "none")
		if err != nil {
			return err
		}
	}
	if !provided["orm"] && o.Database != "none" {
		o.ORM, err = p.choose("Which ORM / data access?", []string{"sql", "gorm"}, []string{"database/sql (no ORM)", "GORM"}, "sql")
		if err != nil {
			return err
		}
	}
	fmt.Fprintln(ui.out)
	return nil
}
