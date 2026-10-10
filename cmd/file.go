package main

import (
	"fmt"
	"path"
	"strconv"

	"acovia.net/record"
)

func listFile() error {
	if len(subCmdArgs) < 1 {
		return usageError("mc-saver list-file <world>", "list-file")
	}
	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
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
	if len(subCmdArgs) < 2 {
		return usageError("mc-saver add-file <world> <file_name>...", "add-file <file_name>...")
	}
	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
		return err
	}

	config.File = append(config.File, subCmdArgs[1:]...)

	return commitConfig(configFilePath)
}

func delFile() error {
	if len(subCmdArgs) < 2 {
		return usageError("mc-saver del-file <world> <index>...", "del-file <index>...")
	}
	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
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
	if len(subCmdArgs) < 3 {
		return usageError("mc-saver mod-file <world> <index> <file_name>", "mod-file <index> <file_name>")
	}
	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
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
