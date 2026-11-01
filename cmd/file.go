package main

import (
	"fmt"
	"strconv"

	"acovia.net/record"
)

func listFile() error {
	usage := "mc-saver list-file <world>"
	replUsage := "list-file"
	_, _, err := loadWorld(1, usage, replUsage)
	if err != nil {
		return err
	}

	printFileRules()

	return nil
}

// printFileRules prints the file rules of the loaded config.
func printFileRules() {
	if len(config.File) != 0 {
		fmt.Printf("file config:\n")
		for i, v := range config.File {
			fmt.Printf("	- %v: %q\n", i, v)
		}
	}
}

func addFile() error {
	usage := "mc-saver add-file <world> <file_name>..."
	replUsage := "add-file <file_name>..."
	_, configFilePath, err := loadWorld(2, usage, replUsage)
	if err != nil {
		return err
	}

	config.File = append(config.File, subCmdArgs[1:]...)

	return commitConfig(configFilePath)
}

func delFile() error {
	usage := "mc-saver del-file <world> <index>..."
	replUsage := "del-file <index>..."
	_, configFilePath, err := loadWorld(2, usage, replUsage)
	if err != nil {
		return err
	}

	indexSet, err := convertIntArray(subCmdArgs[1:])
	if err != nil {
		return record.FmtError("parse command line args:", err)
	}

	config.File, err = deleteSliceElements(config.File, indexSet...)
	if err != nil {
		return record.FmtError("delete element:", err)
	}

	return commitConfig(configFilePath)
}

func modFile() error {
	usage := "mc-saver mod-file <world> <index> <file_name>"
	replUsage := "mod-file <index> <file_name>"
	_, configFilePath, err := loadWorld(3, usage, replUsage)
	if err != nil {
		return err
	}

	index, err := strconv.ParseInt(subCmdArgs[1], 10, 32)
	if err != nil {
		return record.FmtError("parse command line args:", err)
	}

	if index < 0 || int(index) >= len(config.File) {
		return record.FmtError("index out of range:", index)
	}

	config.File[index] = subCmdArgs[2]

	return commitConfig(configFilePath)
}
