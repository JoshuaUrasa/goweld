package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const (
	cyan   = "\033[36m"
	purple = "\033[35m"
	green  = "\033[32m"
	red    = "\033[31m"
	dim    = "\033[2m"
	bold   = "\033[1m"
	reset  = "\033[0m"
)

type terminalUI struct {
	out     io.Writer
	color   bool
	animate bool
}

func newUI(out io.Writer) terminalUI {
	terminal := false
	if file, ok := out.(*os.File); ok {
		if info, err := file.Stat(); err == nil {
			terminal = info.Mode()&os.ModeCharDevice != 0 && os.Getenv("TERM") != "dumb"
		}
	}
	_, noColor := os.LookupEnv("NO_COLOR")
	return terminalUI{out: out, color: terminal && !noColor && os.Getenv("CLICOLOR") != "0", animate: terminal}
}

func (u terminalUI) paint(style, text string) string {
	if !u.color {
		return text
	}
	return style + text + reset
}

func (u terminalUI) banner() {
	fmt.Fprintln(u.out)
	fmt.Fprintln(u.out, u.paint(bold+cyan, "  GO")+u.paint(bold+purple, "WELD")+u.paint(dim, "  /  v"+version))
	fmt.Fprintln(u.out, u.paint(dim, "  Build your next Go idea."))
	fmt.Fprintln(u.out, u.paint(dim, "  ──────────────────────────────────────────"))
	fmt.Fprintln(u.out)
}

func (u terminalUI) info(label, value string) {
	fmt.Fprintf(u.out, "  %s %s\n", u.paint(cyan, label), value)
}

func (u terminalUI) success(message string) {
	fmt.Fprintf(u.out, "  %s %s\n", u.paint(green, "✓"), message)
}

func (u terminalUI) failure(message string) {
	fmt.Fprintf(u.out, "  %s %s\n", u.paint(red, "✗"), message)
}

// task animates only on terminals. It joins the renderer before printing a result.
// Work must not write to u.out while the renderer owns the progress line.
func (u terminalUI) task(label string, work func() error) error {
	started := time.Now()
	var stop, finished chan struct{}
	if u.animate {
		stop, finished = make(chan struct{}), make(chan struct{})
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		render := func(frame string) {
			fmt.Fprintf(u.out, "\r\033[2K  %s %s%s", u.paint(cyan, frame), label, u.paint(dim, "…"))
		}
		render(frames[0])
		go func() {
			defer close(finished)
			ticker := time.NewTicker(90 * time.Millisecond)
			defer ticker.Stop()
			index := 1
			for {
				select {
				case <-stop:
					return
				case <-ticker.C:
					render(frames[index%len(frames)])
					index++
				}
			}
		}()
	} else {
		fmt.Fprintf(u.out, "  > %s...\n", label)
	}
	err := work()
	if u.animate {
		close(stop)
		<-finished
		fmt.Fprint(u.out, "\r\033[2K")
	}
	if err != nil {
		u.failure(label + " failed")
		return err
	}
	elapsed := time.Since(started).Round(time.Millisecond)
	u.success(label + u.paint(dim, " ("+elapsed.String()+")"))
	return nil
}

func (u terminalUI) next(commands ...string) {
	fmt.Fprintln(u.out)
	fmt.Fprintln(u.out, u.paint(bold, "  Next steps"))
	for _, command := range commands {
		fmt.Fprintln(u.out, "    "+u.paint(cyan, command))
	}
	fmt.Fprintln(u.out)
}

func stackSummary(framework, database, orm string) string {
	if orm == "" {
		orm = "none"
		if database != "none" {
			orm = "sql"
		}
	}
	if framework == "std" {
		framework = "net/http"
	}
	return strings.Join([]string{framework, database, orm}, " · ")
}
