package main

import (
	"bytes"
	"testing"
)

func TestCommands(t *testing.T) {
	for _, args := range [][]string{nil, {"help"}, {"new", "--help"}, {"version"}} {
		var output bytes.Buffer
		if err := run(args, &output, &output); err != nil || output.Len() == 0 {
			t.Fatalf("%v: %v, %q", args, err, output.String())
		}
	}
	for _, args := range [][]string{{"unknown"}, {"new"}, {"new", "app", "extra"}, {"new", "app", "--framework", "invalid"}, {"generate"}, {"version", "extra"}, {"run", "extra"}, {"build", "extra"}} {
		var output bytes.Buffer
		if err := run(args, &output, &output); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}
