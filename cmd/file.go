package main

import (
	"strconv"

	"acovia.net/record"
)

func listFile() error {
	usage := "mc-saver ls-file <world>"
	replUsage := "ls-file"
	_, _, err := loadWorld(usage, replUsage)
	if err != nil {
		return err
	}

	printFileRules()

	return nil
}

func printFileRules() {
	if len(config.File) != 0 {
		record.Printf("file config:\n")
		for i, v := range config.File {
			record.Printf("	- %v: %q\n", i, v)
		}
	}
}

func addFile() error {
	usage := "mc-saver add-file <world> <file_name>..."
	replUsage := "add-file <file_name>..."
	_, configFilePath, err := loadWorld(usage, replUsage)
	if err != nil {
		return err
	}

	config.File = append(config.File, subCmdArgs[1:]...)

	return commitConfig(configFilePath)
}

func delFile() error {
	usage := "mc-saver rm-file <world> <index>..."
	replUsage := "rm-file <index>..."
	_, configFilePath, err := loadWorld(usage, replUsage)
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
	usage := "mc-saver set-file <world> <index> <file_name>"
	replUsage := "set-file <index> <file_name>"
	_, configFilePath, err := loadWorld(usage, replUsage)
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
