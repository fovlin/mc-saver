package main

import (
	"os"

	"acovia.net/opt"
	"acovia.net/record"
)

type CommandObject struct {
	MinArgsNumber    int
	NeedDimensionArg bool
	Function         func() error
}

var (
	cmdInfo map[string]CommandObject
	cmdMap  map[string]func() error
)

func cmdObject(minArgsNumber int, needDimensionArg bool, function func() error) CommandObject {
	return CommandObject{
		MinArgsNumber:    minArgsNumber,
		NeedDimensionArg: needDimensionArg,
		Function:         function,
	}
}

func main() {
	err := initProgram()
	if err != nil {
		record.Error(err)
	}

	function, ok := cmdMap[cmd]
	if !ok {
		record.Error("unknown command:", cmd)
	}

	if err := function(); err != nil {
		record.Error(err)
	}
}

func initProgram() error {
	cmdInfo = map[string]CommandObject{
		"run":        cmdObject(1, false, run),
		"gencfg":     cmdObject(1, false, gencfg),
		"repl":       cmdObject(0, false, repl),
		"help":       cmdObject(0, false, help),
		"about":      cmdObject(0, false, about),
		"ls":         cmdObject(1, false, list),
		"ls-config":  cmdObject(2, true, listConfig),
		"ls-dms":     cmdObject(1, false, listDms),
		"add-dms":    cmdObject(2, true, addDms),
		"rm-dms":     cmdObject(2, true, delDms),
		"set-dms":    cmdObject(3, true, modDms),
		"ls-range":   cmdObject(2, true, listRange),
		"add-range":  cmdObject(6, true, addRange),
		"rm-range":   cmdObject(3, true, delRange),
		"set-range":  cmdObject(7, true, modRange),
		"ls-simple":  cmdObject(2, true, listSimple),
		"add-simple": cmdObject(4, true, addSimple),
		"rm-simple":  cmdObject(3, true, delSimple),
		"set-simple": cmdObject(5, true, modSimple),
		"ls-file":    cmdObject(1, false, listFile),
		"add-file":   cmdObject(2, false, addFile),
		"rm-file":    cmdObject(2, false, delFile),
		"set-file":   cmdObject(3, false, modFile),
	}

	cmdMap = make(map[string]func() error, len(cmdInfo))
	for name, command := range cmdInfo {
		cmdMap[name] = command.Function
	}

	opt.BoolSwitch(&useLegacyMode, "l", "legacy")
	opt.BoolSwitch(&record.EnableColor, "c", "color")
	opt.BoolSwitch(&record.EnableTimestamp, "t", "time")
	opt.BoolFunc("", "help", func() {
		help()
		os.Exit(0)
	})

	err := opt.Parse()
	if err != nil {
		return err
	}

	args := opt.Args()
	if len(args) < 1 {
		args = []string{"repl"}
	}

	if isPath(args[0]) {
		args = []string{"repl", args[0]}
	}

	cmd = args[0]
	subCmdArgs = args[1:]

	return nil
}
