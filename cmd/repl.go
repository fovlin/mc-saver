package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"acovia.net/record"
)

var (
	replSpecCmd = map[string]func(){
		"exit": func() {
			os.Exit(0)
		},
		"help": printHelper,
	}
)

func repl() (err error) {
	// At least one paramter to execute this stage.
	if len(subCmdArgs) < 1 {
		err := replNoArgs()
		if err != nil {
			return err
		}
	}

	worldDirPath = subCmdArgs[0] // get the world Directory

	err = initWorldConfig()
	if err != nil {
		return err
	}

	record.Info("==== MC-SAVER Repl Mode ====")
	record.Info("enter repl mode for", worldDirPath)
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

		replSpecFunc, ok := replSpecCmd[args[0]]
		if ok {
			replSpecFunc()
		}

		cmd = args[0]                                            // get the command header
		subCmdArgs = append([]string{worldDirPath}, args[1:]...) // get the command arguments

		err := executeReplCommand()
		if err != nil {
			record.ErrorNoExit(err)
		}

	}
	// Error Check

	if err := scanner.Err(); err != nil {
		return fmtErr("scanner error:", err)
	}

	return nil
}

func replNoArgs() (err error) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("mc-saver: select world > ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text()) // remove the space char.
		args = strings.Fields(line)               // Get text slices by Fields()
		if len(args) == 0 {
			continue
		}

		subCmdArgs = []string{args[0]}
		break
	}

	err = scanner.Err()
	if err != nil {
		return err
	}

	return nil
}

func printHelper() {
	fmt.Println(replHelpInfo)

}
func executeReplCommand() (err error) {
	function, ok := cmdMap[cmd]
	if !ok {
		return fmtErr("unknown command: " + cmd)
	}

	err = function()
	if err != nil {
		record.ErrorNoExit(err)
	}

	return nil
}
