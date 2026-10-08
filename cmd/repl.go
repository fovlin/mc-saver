package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"acovia.net/record"
)

// repl is command for world.

var (
	replNoArgsCmd = map[string]func() error{
		"exit":   exit,
		"select": replNoArgsSelectWorld,
		"help":   replHelp,
	}
)

func repl() (err error) {
	if len(subCmdArgs) < 1 {
		err = replNoArgs()
		if err != nil {
			return err
		}
	}

	if len(subCmdArgs) > 0 {
		worldDirPath = subCmdArgs[0]
		err = initWorldConfig()
		if err != nil {
			return err
		}
	}

	record.Info("==== MC-SAVER Repl Mode ====")
	record.Info("selected:", subCmdArgs[0])
	record.Info("type 'help' for commands, 'exit' or 'quit' to quit repl mode ")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("mc-saver > ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text()) // remove the space char.
		args = strings.Fields(line)               // Get text slices by Fields()
		if len(args) == 0 {
			continue
		}

		cmd = args[0]                                    // get the command header
		subCmdArgs = append(subCmdArgs[:1], args[1:]...) // get the command arguments
		replSpecFunc, ok := replCmdMap[args[0]]
		if ok {
			err = replSpecFunc()
			if err != nil {
				record.ErrorNoExit(err)
			}
			continue
		}

		err := runReplCmd()
		if err != nil {
			record.ErrorNoExit(err)
			continue
		}
	}
	// Error Check

	if err := scanner.Err(); err != nil {
		return formatError("scanner error:", err)
	}

	return nil
}

func replNoArgs() (err error) {
	scanner := bufio.NewScanner(os.Stdin)
	for len(subCmdArgs) == 0 {
		record.Info("no world selected, please run: 'select <world>'")
		fmt.Print("mc-saver: > ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text()) // remove the space char.
		args = strings.Fields(line)               // Get text slices by Fields()
		if len(args) == 0 {
			continue
		}
		cmd = args[0]
		replSpecFunc, ok := replNoArgsCmd[cmd]
		if ok {
			err = replSpecFunc()
			if err != nil {
				record.ErrorNoExit(err)
			}
			continue
		} else {
			record.ErrorNoExit("unknow command:", cmd)
		}
	}

	err = scanner.Err()
	if err != nil {
		return err
	}

	return nil
}

func runReplCmd() (err error) {
	cmdFunc, ok := replCmdMap[cmd]
	if !ok {
		return formatError("unknow command:", cmd)
	}
	err = cmdFunc()
	if err != nil {
		return err
	}

	return nil
}

func replHelp() (err error) {
	fmt.Println(replHelpInfo)
	return nil
}

func exit() (err error) {
	os.Exit(0)
	return nil
}

func replSelectWorld() (err error) {
	if len(subCmdArgs) < 2 {
		return formatError("syntax error, usage: select <world>")
	}

	worldDirPath = subCmdArgs[1]
	err = initWorldConfig()
	if err != nil {
		return formatError(err)
	}

	subCmdArgs[0] = subCmdArgs[1]

	return err
}

func replNoArgsSelectWorld() (err error) {
	if len(args) < 2 {
		return formatError("syntax error, usage: select <world>")
	}

	subCmdArgs = []string{args[1]}
	worldDirPath = subCmdArgs[0]

	err = initWorldConfig()
	if err != nil {
		subCmdArgs = make([]string, 0)
		return formatError(err)
	}

	return err
}
