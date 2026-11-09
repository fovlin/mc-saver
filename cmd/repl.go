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

type Completer struct {
	Terminal *term.Terminal
	Loaders  []func(field string) []string
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
	var completer = &Completer{
		Loaders: []func(field string) []string{
			loadFirstArgCandidate,
			loadSecondCandidate,
		},
	}

	terminal.AutoCompleteCallback = completer.CompleteArgs
	completer.Terminal = terminal

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

func (completer *Completer) CompleteArgs(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
	if key != '\t' {
		return line, pos, false
	}

	if line == "" {
		return line, pos, false
	}

	fields := opt.ParseLine(line[:pos])
	index := len(fields) - 1

	if index >= len(completer.Loaders) {
		return line, pos, false
	}

	commandField := fields[index]
	candidateList := completer.Loaders[index](fields[0])

	selectedCandidateList, selectedNumber := searchCandidate(candidateList, commandField)

	if selectedNumber == 0 {
		return line, pos, false
	}

	if selectedNumber == 1 {
		if index > 0 {
			return completer.BuildLine(line, pos, selectedCandidateList[0], false)
		}

		command, ok := cmdInfo[selectedCandidateList[0]]
		if !ok {
			return line, pos, false
		}

		switch true {
		case command.MinArgsNumber > 1:
			return completer.BuildLine(line, pos, selectedCandidateList[0], true)

		default:
			return completer.BuildLine(line, pos, selectedCandidateList[0], false)
		}
	}

	sort.Strings(selectedCandidateList)

	var completed bool
	for _, char := range selectedCandidateList[0][len(commandField):] {
		_, acount := searchCandidate(selectedCandidateList, commandField+string(char))
		if acount != selectedNumber {
			break
		}
		commandField += string(char)
		completed = true
	}

	if !completed {
		record.Println()
		for _, candidate := range selectedCandidateList {
			completer.Terminal.Write([]byte("| " + candidate + "\r\n"))
		}
		return line, pos, false
	}

	return completer.BuildLine(line, pos, commandField, false)
}

func (completer Completer) BuildLine(line string, pos int, candidate string, nextIsArgs bool) (string, int, bool) {
	if pos > len(line) {
		pos = len(line)
	}

	prefix := line[:pos]
	cut := strings.LastIndexAny(prefix, " \t") + 1

	newPrefix := prefix[:cut] + candidate
	if nextIsArgs {
		newPrefix += " "
	}

	return newPrefix + line[pos:], len(newPrefix), true
}

func loadFirstArgCandidate(_ string) []string {
	var candidateList []string
	for command := range cmdMap {
		candidateList = append(candidateList, command)
	}
	return candidateList
}

func loadSecondCandidate(field string) []string {
	var candidateList []string

	namespaceIDList, err := save.NamespaceIDList(field, useLegacyMode)
	if err != nil {
		return nil
	}

	for _, namespaceID := range namespaceIDList {
		candidateList = append(candidateList, namespaceID)
	}

	return candidateList
}

func searchCandidate(candidateList []string, field string) ([]string, int) {
	var acount int
	var selectedList []string
	for _, testField := range candidateList {
		if strings.HasPrefix(testField, field) {
			selectedList = append(selectedList, testField)
			acount++
		}
	}

	return selectedList, acount
}
