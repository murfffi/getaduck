package enumflag

// Adapted from https://github.com/creachadair/goflags/blob/main/enumflag/flag.go
// under BSD 3-Clause license

import (
	"bytes"
	"flag"
	"io"
	"testing"
)

func newFlagSet(name string, w io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(w)
	return fs
}

func TestFlagBits(t *testing.T) {
	color := New("red", "orange", "yellow", "green", "blue")

	const initial = "red"
	const flagged = "green"
	const flaggedIndex = 3

	var buf bytes.Buffer
	fs := newFlagSet("color", &buf)
	fs.Var(color, "color", color.Help("The color to paint the bike shed"))
	fs.PrintDefaults()
	t.Logf("Color flag set:\n%s", buf.String())
	buf.Reset()

	if key := color.Key(); key != initial {
		t.Errorf("Initial value for -color: got %q, want %q", key, initial)
	}

	if err := fs.Parse([]string{"-color", "GREEN"}); err != nil {
		t.Fatalf("Argument parsing failed: %v", err)
	}

	if key := color.Key(); key != flagged {
		t.Errorf("Value for -color: got %q, want %q", key, flagged)
	}
	if idx := color.Index(); idx != flaggedIndex {
		t.Errorf("Index for -color: got %d, want %d", idx, flaggedIndex)
	}

	taste := New("", "sweet", "sour")
	fs = newFlagSet("taste", &buf)
	fs.Var(taste, "taste", taste.Help("The flavour of the ice cream"))
	fs.PrintDefaults()
	t.Logf("Taste flag set:\n%s", buf.String())

	if err := fs.Parse([]string{"-taste", "crud"}); err == nil {
		t.Error("Expected error from bogus flag, but got none")
	} else {
		t.Logf("Got expected error from bogus -taste: %v", err)
	}
}
