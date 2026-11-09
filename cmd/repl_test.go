package main

import "testing"

func TestBuildLineKeepsTheRest(t *testing.T) {
	completer := Completer{}

	line, pos, ok := completer.BuildLine("add ran sub", 7, "range", false)
	if !ok || line != "add range sub" || pos != 9 {
		t.Fatalf("mid line: got %q pos %d ok %v", line, pos, ok)
	}

	line, pos, ok = completer.BuildLine("ad", 2, "add-dms", true)
	if !ok || line != "add-dms " || pos != 8 {
		t.Fatalf("tail: got %q pos %d ok %v", line, pos, ok)
	}

	line, pos, ok = completer.BuildLine("  ad", 4, "add-dms", false)
	if !ok || line != "  add-dms" || pos != 9 {
		t.Fatalf("leading spaces: got %q pos %d ok %v", line, pos, ok)
	}
}

func TestCompleteArgsUsesLoaders(t *testing.T) {
	cmdInfo = map[string]CommandObject{
		"add-dms": cmdObject(2, true, func() error { return nil }),
	}

	calls := 0
	completer := Completer{
		Loaders: []func(field string) []string{
			func(field string) []string {
				calls++
				return []string{"add-dms"}
			},
			func(field string) []string {
				calls++
				return []string{"minecraft:overworld", "minecraft:mine"}
			},
		},
	}

	line, pos, ok := completer.CompleteArgs("add-dm", 6, '\t')
	if !ok || line != "add-dms " || pos != 8 {
		t.Fatalf("command completion: got %q pos %d ok %v", line, pos, ok)
	}

	if calls != 1 {
		t.Fatalf("loader should be called once per completion, got %d", calls)
	}

	line, pos, ok = completer.CompleteArgs("add-dms minecraft:m", 19, '\t')
	if !ok || line != "add-dms minecraft:mine" || pos != len("add-dms minecraft:mine") {
		t.Fatalf("arg completion: got %q pos %d ok %v", line, pos, ok)
	}

	if calls != 2 {
		t.Fatalf("second loader should run on the second completion, got %d", calls)
	}
}
