package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	//"text/scanner"

	"acovia.net/record"
	"log"
	//"github.com/docker/docker/libnetwork/drivers/null"
)

func repl() {
	// At least one paramter to execute this stage. 
	if len(subCmdArgs) < 1 { 
		record.Error("Syntax Error, usage: mc-saver repl <world path>") 
	}

	worldDirPath = subCmdArgs[0] // get the world Directory 
	configFilePath = path.Join(worldDirPath,configFileName) // Complete path.

	initWorldConfig() 

	record.Info("==== MC-SAVER Repl Mode ====")
	record.Info("Enter repl mode for", worldDirPath) 
	record.Info("Type 'help' for commands, 'exit' or 'quit' to quit repl mode ")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("mc-saver > ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text()) // remove the space char. 
		if line == "" {
			continue
		}else if line == "exit" || line == "quit" {
			break
		}else if line == "help" {
			printHelper()
			continue
		}
		args := strings.Fields(line) // Get text slices by Fields()
		if len(args) == 0{
			continue
		}

		cmdName := args[0] // get the command header
		cmdArgs := args[1:] // get the command arguments

		executeReplCommand(cmdName, cmdArgs)

		

	}
	// Error Check 

	if err := scanner.Err() ; err != nil {
		log.Printf("scanner error: %v", err)
	}

}	


func printHelper() {
	fmt.Println(`repl commands (world argument is implicit):
  list                 - list all dimension rules and file rules
  list-config <dimension>...  - list range and simple rules of dimension(s)
  list-dms             - list dimension namespace ids
  add-dms <dimension>...      - add dimension(s) with default range rule
  del-dms <dimension>...      - delete dimension(s)
  mod-dms <old> <new>         - rename a dimension
  list-range <dimension>...   - list range rules of dimension(s)
  add-range <dimension> <from_x> <from_y> <to_x> <to_y>  - add range rule
  del-range <dimension> <index>...  - delete range rule(s) by index
  mod-range <dimension> <index> <from_x> <from_y> <to_x> <to_y>  - modify range rule
  list-simple <dimension>...  - list simple rules of dimension(s)
  add-simple <dimension> <x> <y>  - add simple rule
  del-simple <dimension> <index>...  - delete simple rule(s) by index
  mod-simple <dimension> <index> <x> <y>  - modify simple rule
  list-file            - list file rules
  add-file <name>...   - add file rule(s)
  del-file <index>...  - delete file rule(s) by index
  mod-file <index> <name>  - modify file rule
  help                 - show this help
  about				   - show the information and copyright about this kit.
  exit / quit          - leave repl mode`)

}
func executeReplCommand(cmdName string, cmdArgs []string) {
	subCmdArgs = append([]string{worldDirPath}, cmdArgs...) // get the complete command

	function, ok := cmdMap[cmdName]
	if !ok {
		record.Error("Unknown Command: " + cmdName)
		return 
	}
	function()

}



