package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildLine(t *testing.T) {
	completer := Completer{}

	testCases := []struct {
		name       string
		line       string
		pos        int
		candidate  string
		nextIsArgs bool
		want       string
		wantPos    int
	}{
		{
			name:      "replace word in the middle of a line",
			line:      "add ran sub",
			pos:       len("add ran"),
			candidate: "range",
			want:      "add range sub",
			wantPos:   len("add range"),
		},
		{
			name:       "complete command and add a space",
			line:       "ad",
			pos:        len("ad"),
			candidate:  "add-dms",
			nextIsArgs: true,
			want:       "add-dms ",
			wantPos:    len("add-dms "),
		},
		{
			name:      "keep leading spaces",
			line:      "  ad",
			pos:       len("  ad"),
			candidate: "add-dms",
			want:      "  add-dms",
			wantPos:   len("  add-dms"),
		},
		{
			name:       "do not duplicate the separator when there is trailing text",
			line:       "ad rest",
			pos:        len("ad"),
			candidate:  "add-dms",
			nextIsArgs: true,
			want:       "add-dms rest",
			wantPos:    len("add-dms "),
		},
		{
			name:      "replace the rest of the word under the cursor",
			line:      "add-dms minecraft:overworld",
			pos:       len("add-dms minecraft:ov"),
			candidate: "minecraft:overworld",
			want:      "add-dms minecraft:overworld",
			wantPos:   len("add-dms minecraft:overworld"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			gotLine, gotPos, ok := completer.BuildLine(
				testCase.line,
				testCase.pos,
				testCase.candidate,
				testCase.nextIsArgs,
			)
			if !ok {
				t.Fatal("BuildLine returned ok=false")
			}
			if gotLine != testCase.want {
				t.Fatalf("line: got %q, want %q", gotLine, testCase.want)
			}
			if gotPos != testCase.wantPos {
				t.Fatalf("cursor: got %d, want %d", gotPos, testCase.wantPos)
			}
		})
	}
}

func TestCompleteArgsCommand(t *testing.T) {
	restore := saveCompletionState(t)
	defer restore()

	cmdInfo = map[string]CommandObject{
		"add-dms": cmdObject(2, true, nil),
	}

	completer := &Completer{
		Loaders: []func(string) []string{
			func(string) []string { return []string{"add-dms"} },
			func(string) []string { return nil },
		},
	}

	line, pos, ok := completer.CompleteArgs("add-dm", len("add-dm"), '\t')
	if !ok {
		t.Fatal("CompleteArgs returned ok=false")
	}
	if line != "add-dms " {
		t.Fatalf("line: got %q, want %q", line, "add-dms ")
	}
	if pos != len("add-dms ") {
		t.Fatalf("cursor: got %d, want %d", pos, len("add-dms "))
	}
}

func TestCompleteArgsAfterTrailingSpace(t *testing.T) {
	restore := saveCompletionState(t)
	defer restore()

	subCmdArgs = []string{"/tmp/world"}
	cmdInfo = map[string]CommandObject{
		"ls-range": cmdObject(2, true, nil),
	}

	var loadedWorldPath string
	completer := &Completer{
		Loaders: []func(string) []string{
			func(string) []string { return []string{"ls-range"} },
			func(worldPath string) []string {
				loadedWorldPath = worldPath
				return []string{"minecraft:overworld", "minecraft:the_nether"}
			},
		},
	}

	line, pos, ok := completer.CompleteArgs("ls-range ", len("ls-range "), '\t')
	if !ok {
		t.Fatal("CompleteArgs returned ok=false")
	}
	if line != "ls-range minecraft:" {
		t.Fatalf("line: got %q, want %q", line, "ls-range minecraft:")
	}
	if pos != len("ls-range minecraft:") {
		t.Fatalf("cursor: got %d, want %d", pos, len("ls-range minecraft:"))
	}
	if loadedWorldPath != "/tmp/world" {
		t.Fatalf("loader world: got %q, want %q", loadedWorldPath, "/tmp/world")
	}
}

func TestCompleteArgsSkipsDimensionForNonDimensionCommand(t *testing.T) {
	restore := saveCompletionState(t)
	defer restore()

	subCmdArgs = []string{"/tmp/world"}
	cmdInfo = map[string]CommandObject{
		"run": cmdObject(1, false, nil),
	}

	secondLoaderCalled := false
	completer := &Completer{
		Loaders: []func(string) []string{
			func(string) []string { return []string{"run"} },
			func(string) []string {
				secondLoaderCalled = true
				return []string{"minecraft:overworld"}
			},
		},
	}

	if _, _, ok := completer.CompleteArgs("run ", len("run "), '\t'); ok {
		t.Fatal("CompleteArgs completed a dimension for run")
	}
	if secondLoaderCalled {
		t.Fatal("dimension loader was called for run")
	}
}

func TestCompleteArgsDoesNotPanicOnWhitespace(t *testing.T) {
	restore := saveCompletionState(t)
	defer restore()

	completer := &Completer{
		Loaders: []func(string) []string{
			func(string) []string { return nil },
			func(string) []string { return nil },
		},
	}

	if _, _, ok := completer.CompleteArgs("   ", len("   "), '\t'); ok {
		t.Fatal("CompleteArgs on whitespace returned ok=true")
	}
}

func TestCompleteArgsUsesSelectedWorld(t *testing.T) {
	restore := saveCompletionState(t)
	defer restore()

	worldPath := t.TempDir()
	dimensionPath := filepath.Join(worldPath, "dimensions", "minecraft", "overworld")
	if err := os.MkdirAll(dimensionPath, 0755); err != nil {
		t.Fatal(err)
	}

	subCmdArgs = []string{worldPath}
	useLegacyMode = false
	cmdInfo = map[string]CommandObject{
		"ls-range": cmdObject(2, true, nil),
	}

	completer := &Completer{
		Loaders: []func(string) []string{
			loadFirstArgCandidate,
			loadSecondCandidate,
		},
	}

	line, pos, ok := completer.CompleteArgs("ls-range minecraft:ov", len("ls-range minecraft:ov"), '\t')
	if !ok {
		t.Fatal("CompleteArgs returned ok=false")
	}
	if line != "ls-range minecraft:overworld" {
		t.Fatalf("line: got %q, want %q", line, "ls-range minecraft:overworld")
	}
	if pos != len("ls-range minecraft:overworld") {
		t.Fatalf("cursor: got %d, want %d", pos, len("ls-range minecraft:overworld"))
	}
}

func saveCompletionState(t *testing.T) func() {
	t.Helper()

	oldCmdInfo := cmdInfo
	oldSubCmdArgs := subCmdArgs
	oldUseLegacyMode := useLegacyMode

	return func() {
		cmdInfo = oldCmdInfo
		subCmdArgs = oldSubCmdArgs
		useLegacyMode = oldUseLegacyMode
	}
}
