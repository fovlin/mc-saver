package main

import (
	"flag"

	"acovia.net/record"
)

var cmdMap map[string]func() error

func init() {
	cmdMap = map[string]func() error{
		"run":         run,
		"gencfg":      gencfg,
		"repl":        repl,
		"help":        help,
		"about":       about,
		"list":        list,
		"list-config": listConfig,
		"list-dms":    listDms,
		"add-dms":     addDms,
		"del-dms":     delDms,
		"mod-dms":     modDms,
		"list-range":  listRange,
		"add-range":   addRange,
		"del-range":   delRange,
		"mod-range":   modRange,
		"list-simple": listSimple,
		"add-simple":  addSimple,
		"del-simple":  delSimple,
		"mod-simple":  modSimple,
		"list-file":   listFile,
		"add-file":    addFile,
		"del-file":    delFile,
		"mod-file":    modFile,
	}
}

func main() {
	initProgram()

	function, ok := cmdMap[cmd]
	if !ok {
		record.Error("unknown command:", cmd)
	}

	if err := function(); err != nil {
		record.Error(err)
	}
}

func initProgram() {
	flag.BoolFunc("l", "legacy world mode", func(string) error {
		useLegacyMode = true
		return nil
	})
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		args = []string{"repl"}
	}

	if isPath(fmtPath(args[0])) {
		args = []string{"repl", args[0]}
	}

	cmd = args[0]
	subCmdArgs = args[1:]
}
