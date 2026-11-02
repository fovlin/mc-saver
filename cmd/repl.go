package main

import (
	"errors"
	"io"
	"opt"
	"os"
	"sort"
	"strings"

	"acovia.net/minecraft/save"
	"acovia.net/record"
	"golang.org/x/term"
)

var (
	replOnlyCmd = map[string]func() error{
		"world":  selectedWorld,
		"select": selectWorld,
		"help":   replHelp,
		"exit":   exit,
	}

	restoreTerminal func()
)

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
	terminal.AutoCompleteCallback = newCompleteCallback(terminal)
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

		clear(subCmdArgs[1:])
		config = save.NewNullConfig()
	}
}

func replExec(line string) error {
	fields := opt.ParseLine(line)
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

		fields := opt.ParseLine(strings.TrimSpace(command))
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
			record.Print(askWorldHelpInfo)

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

func newCompleteCallback(terminal *term.Terminal) func(line string, pos int, key rune) (string, int, bool) {
	return func(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
		if key != '\t' {
			return "", pos, false
		}

		fields := opt.ParseLine(line[:pos])

		switch true {
		case len(fields) == 1:
			newLine, ok = completeFirstArg(fields, terminal)
			return newLine, len(newLine), ok

		case len(fields) == 2:
			newLine, ok := completeSecondArg(fields, terminal)
			return newLine, len(newLine), ok

		}

		return "", pos, false
	}
}

func loadFirstArgCandidate(field string) []string {
	var candidateList []string
	for command := range cmdMap {
		if strings.HasPrefix(command, field) {
			candidateList = append(candidateList, command)
		}
	}
	return candidateList
}

func loadSecondCandidate(field string) ([]string, error) {
	var candidateList []string

	namespaceIDList, err := save.NamespaceIDList(subCmdArgs[0], useLegacyMode)
	if err != nil {
		return nil, nil
	}

	for _, namespaceID := range namespaceIDList {
		if strings.HasPrefix(namespaceID, field) {
			candidateList = append(candidateList, namespaceID)
		}
	}

	return candidateList, nil
}

func searchCandidate(candidateList []string, field string) ([]string, int) {
	var acount int
	var newCandidateList []string
	for _, testField := range candidateList {
		if strings.HasPrefix(testField, field) {
			newCandidateList = append(newCandidateList, testField)
			acount++
		}
	}

	return newCandidateList, acount
}

func completeFirstArg(fields []string, terminal *term.Terminal) (string, bool) {
	candidateList := loadFirstArgCandidate(fields[0])

	if len(candidateList) == 1 {
		command, ok := cmdInfo[candidateList[0]]
		if !ok {
			return "", false
		}

		switch true {
		case command.MinArgsNumber > 1:
			return candidateList[0] + " ", true

		default:
			return candidateList[0], true
		}
	}

	if len(candidateList) > 1 {
		sort.Strings(candidateList)

		for _, char := range candidateList[0][len(fields[0]):] {
			_, acount := searchCandidate(candidateList, fields[0]+string(char))
			if acount != len(candidateList) {
				break
			}
			fields[0] += string(char)
		}

		record.Println()
		for _, candidate := range candidateList {
			terminal.Write([]byte("| " + candidate + "\r\n"))
		}
	}

	return fields[0], true
}

func completeSecondArg(fields []string, terminal *term.Terminal) (string, bool) {
	command, ok := cmdInfo[fields[0]]
	if !ok || !command.NeedDimensionArg {
		return "", false
	}

	candidateList, err := loadSecondCandidate(fields[1])
	if err != nil {
		return "", false
	}

	if len(candidateList) == 1 {
		switch true {
		case command.MinArgsNumber > 2:
			return record.Build(fields[0], candidateList[0]) + " ", true

		default:
			return record.Build(fields[0], candidateList[0]), true
		}
	}

	if len(candidateList) > 1 {
		sort.Strings(candidateList)

		for _, char := range candidateList[0][len(fields[1]):] {
			_, acount := searchCandidate(candidateList, fields[1]+string(char))

			if acount != len(candidateList) {
				break
			}
			fields[1] += string(char)
		}

		record.Println()
		for _, candidate := range candidateList {
			terminal.Write([]byte("| " + candidate + "\r\n"))
		}
	}

	return record.Build(fields[0], fields[1]), true
}
