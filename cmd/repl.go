package main

import (
	"errors"
	"io"
	"opt"
	"os"
	"path"
	"sort"
	"strings"

	"acovia.net/minecraft/save"
	"acovia.net/record"
	"golang.org/x/term"
)

var replOnlyCmd map[string]func() error
var terminal *term.Terminal
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

	terminal = term.NewTerminal(os.Stdin, prompt)

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
	terminal.AutoCompleteCallback = completeCallback
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

func completeCallback(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
	if key != '\t' {
		return "", pos, false
	}

	fields := opt.ParseLine(line[:pos])

	switch true {
	case len(fields) == 1:
		fields[0], newPos, ok = completeFirstArg(fields[0], pos)
		return fields[0], newPos, ok

	case len(fields) == 2:
		fields[1], newPos, ok = completeSecondArg(fields[0], fields[1], pos)
		newLine = fields[0] + " " + fields[1]
		return newLine, len(newLine), ok

	}

	return "", pos, false
}

func searchFiestCandidate(field string) ([]string, int) {
	var candidateList []string
	var acount int
	for command := range cmdMap {
		if strings.HasPrefix(command, field) {
			candidateList = append(candidateList, command)
			acount++
		}
	}
	return candidateList, acount
}

func completeFirstArg(field string, pos int) (newField string, newPos int, ok bool) {
	candidateList, acount := searchFiestCandidate(field)

	if len(candidateList) == 1 {
		return candidateList[0] + " ", len(candidateList[0] + " "), true
	}

	if len(candidateList) > 1 {
		sort.Strings(candidateList)

		var newField string
		for _, char := range candidateList[0] {
			newField += string(char)
			_, newAcount := searchFiestCandidate(newField)
			if newAcount != acount {
				break
			}

			field = newField
		}

		record.Println()
		for _, candidate := range candidateList {
			terminal.Write([]byte("| " + candidate + "\r\n"))
		}
	}

	return field, len(field), true
}

func completeSecondArg(command string, field string, pos int) (newField string, newPos int, ok bool) {
	candidateList, acount, err := searchSecondCandidate(command, field)
	if err != nil {
		return field, pos, false
	}

	if len(candidateList) == 1 {
		return candidateList[0] + " ", len(candidateList[0] + " "), true
	}

	if len(candidateList) > 1 {
		sort.Strings(candidateList)

		var newField string = field
		for _, char := range strings.TrimPrefix(candidateList[0], field) {
			newField += string(char)
			_, newAcount, err := searchSecondCandidate(command, newField)
			if err != nil {
				break
			}

			if newAcount != acount {
				break
			}

			field = newField
		}

		record.Println()
		for _, candidate := range candidateList {
			terminal.Write([]byte("| " + candidate + "\r\n"))
		}
	}

	return field, len(field), true
}

func searchSecondCandidate(command string, field string) ([]string, int, error) {
	var candidateList []string
	var acount int

	if !isDimensionArgCommand(command) {
		return candidateList, acount, nil
	}

	idList, err := dimensionIDList()
	if err != nil {
		return candidateList, acount, err
	}

	for _, id := range idList {
		if strings.HasPrefix(id, field) {
			candidateList = append(candidateList, id)
			acount++
		}
	}

	return candidateList, acount, nil
}

func isDimensionArgCommand(command string) bool {
	switch true {
	case command == "add-dms":
		return true
	case command == "del-dms":
		return true
	case command == "mod-dms":
		return true
	case command == "list-config":
		return true
	case command == "list-range":
		return true
	case command == "add-range":
		return true
	case command == "del-range":
		return true
	case command == "mod-range":
		return true
	case command == "list-simple":
		return true
	case command == "add-simple":
		return true
	case command == "del-simple":
		return true
	case command == "mod-simple":
		return true
	}

	return false
}

func dimensionIDList() (idList []string, err error) {
	var idSet map[string]bool = map[string]bool{}

	for id := range defaultConfig.Dimension {
		idSet[id] = true
	}

	if len(subCmdArgs) > 0 {
		err = addWorldDimensionID(idSet, subCmdArgs[0])
		if err != nil {
			return idList, err
		}
	}

	for id := range idSet {
		idList = append(idList, id)
	}
	sort.Strings(idList)

	return idList, nil
}

func addWorldDimensionID(idSet map[string]bool, worldDirPath string) error {
	dimensionsDirPath := path.Join(worldDirPath, "dimensions")

	namespaceList, err := os.ReadDir(dimensionsDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, namespace := range namespaceList {
		if !namespace.IsDir() {
			continue
		}

		namespaceName := namespace.Name()
		subIDList, err := os.ReadDir(path.Join(dimensionsDirPath, namespaceName))
		if err != nil {
			return err
		}

		for _, subID := range subIDList {
			if !subID.IsDir() {
				continue
			}

			idSet[namespaceName+":"+subID.Name()] = true
		}
	}

	return nil
}
