package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"acovia.net/record"
)

var replOnlyCmd map[string]func() error

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

	scanner := bufio.NewScanner(os.Stdin)

	if len(subCmdArgs) < 1 {
		err := askWorld(scanner)
		if err != nil {
			return err
		}
	}

	if len(subCmdArgs) > 0 {
		if err := initWorld(subCmdArgs[0]); err != nil {
			return err
		}
	}

	record.Info("selected:", subCmdArgs[0])
	record.Info("type 'help' for commands, 'exit' to quit repl mode")

	return replLoop(scanner)
}

func replLoop(scanner *bufio.Scanner) error {
	for {
		fmt.Print("mc-saver > ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if err := replExec(line); err != nil {
			record.ErrorNoExit(err)
		}

		clear(subCmdArgs[1:])
	}

	return scanner.Err()
}

func replExec(line string) error {
	fields := strings.Fields(line)
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

func askWorld(scanner *bufio.Scanner) (err error) {
	record.Info("no world selected, please run: 'select <world>'")

	for {
		fmt.Print("mc-saver: > ")
		if !scanner.Scan() {
			fmt.Println()
			return record.FmtError("input is empty")
		}

		fields := strings.Fields(strings.TrimSpace(scanner.Text()))
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
			fmt.Println(askWorldHelpInfo)

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
	return nil
}

func exit() error {
	os.Exit(0)
	return nil
}

func selectedWorld() error {
	record.Info("selected:", subCmdArgs[0])
	return nil
}
