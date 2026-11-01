package main

import (
	"errors"
	"io"
	"os"
	"strings"

	"acovia.net/minecraft/save"
	"acovia.net/record"
	"golang.org/x/term"
)

var replOnlyCmd map[string]func() error

var restoreTerminal func()

func init() {
	replOnlyCmd = map[string]func() error{
		"world":  selectedWorld,
		"select": selectWorld,
		"help":   replHelp,
		"exit":   exit,
	}
}

func repl() error {
	replMode = true
	record.Terminal = true

	terminal := term.NewTerminal(os.Stdin, prompt)

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return record.FmtError("repl needs a terminal:", err)
	}

	restoreTerminal = func() {
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
	}
	defer func() {
		if restoreTerminal != nil {
			restoreTerminal()
			restoreTerminal = nil
		}
	}()

	if len(subCmdArgs) < 1 {
		err := askWorld(terminal)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}

	if len(subCmdArgs) == 0 {
		return nil
	}

	if err := initWorld(subCmdArgs[0]); err != nil {
		return err
	}

	record.Info("selected:", "'"+subCmdArgs[0]+"',", "type 'help' for commands, 'exit' to quit repl mode")

	return replLoop(terminal)
}

func replLoop(terminal *term.Terminal) error {
	for {
		command, err := terminal.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		line := strings.TrimSpace(command)
		if line == "" {
			continue
		}

		if err := replExec(line); err != nil {
			record.ErrorNoExit(err)
		}

		// reset var
		clear(subCmdArgs[1:])
		config = save.NewNullConfig()
	}
}

func replExec(line string) error {
	fields := parseCmdArgs(line)
	if len(fields) == 0 {
		return nil
	}

	cmd = fields[0]
	subCmdArgs = append(subCmdArgs[:1], fields[1:]...)

	function, ok := replOnlyCmd[cmd]
	if ok {
		return function()
	}

	function, ok = cmdMap[cmd]
	if ok {
		return function()
	}

	return record.FmtError("unknown command:", cmd)
}

func askWorld(terminal *term.Terminal) (err error) {
	record.Info("no world selected, please run: 'select <world>'")

	for {
		command, err := terminal.ReadLine()
		if err != nil {
			return err
		}

		fields := parseCmdArgs(strings.TrimSpace(command))
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {

		case "select":
			if len(fields) < 2 {
				record.ErrorNoExit(record.FmtError("syntax error, usage: select <world>"))
				continue
			}

			err = initWorld(fields[1])
			if err != nil {
				record.ErrorNoExit(err)
				continue
			}

			subCmdArgs = []string{fields[1]}
			return nil

		case "help":
			record.Println(askWorldHelpInfo)

		case "exit":
			exit()

		default:
			record.ErrorNoExit("unknown command:", fields[0])
		}
	}
}

func selectWorld() error {
	if len(subCmdArgs) < 2 {
		return record.FmtError("syntax error, usage: select <world>")
	}
	err := initWorld(subCmdArgs[1])
	if err != nil {
		return err
	}

	subCmdArgs = []string{subCmdArgs[1]}
	record.Info("selected:", subCmdArgs[0])
	return nil
}

func exit() error {
	if restoreTerminal != nil {
		restoreTerminal()
	}
	os.Exit(0)
	return nil
}

func selectedWorld() error {
	record.Info("selected:", "'"+subCmdArgs[0]+"'")
	return nil
}
