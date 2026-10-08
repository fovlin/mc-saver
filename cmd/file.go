package main

import (
	"fmt"
	"strconv"
)

func listFile() (err error) {
	if len(subCmdArgs) < 1 {
		return formatError("syntax error, usage: mc-saver list-file <world>")
	}

	err = listFileFunc()
	if err != nil {
		return err
	}

	return nil
}

func replListFile() (err error) {
	if len(subCmdArgs) < 1 {
		return formatError("syntax error, usage: list-file")
	}

	err = listFileFunc()
	if err != nil {
		return err
	}

	return nil
}

func listFileFunc() (err error) {
	worldDirPath = subCmdArgs[0]
	err = initWorldConfig()
	if err != nil {
		return err
	}

	if len(config.File) == 0 {
		fmt.Println("no file config")
	}

	for i, v := range config.File {
		fmt.Printf("- %v: %q\n", i, v)
	}

	return nil
}

func addFile() (err error) {
	if len(subCmdArgs) < 2 {
		return formatError("syntax error, usage: mc-saver add-file <world> <file_name>...")
	}

	err = addFileFunc()
	if err != nil {
		return err
	}

	return nil
}

func replAddFile() (err error) {
	if len(subCmdArgs) < 2 {
		return formatError("syntax error, usage: add-file <file_name>...")
	}

	err = addFileFunc()
	if err != nil {
		return err
	}

	return nil
}

func addFileFunc() (err error) {
	worldDirPath = subCmdArgs[0]
	err = initWorldConfig()
	if err != nil {
		return err
	}

	config.File = append(config.File, subCmdArgs[1:]...)

	err = saveConfig()
	if err != nil {
		return formatError("save config:", err)
	}

	return nil
}

func delFile() (err error) {
	if len(subCmdArgs) < 2 {
		return formatError("syntax error, usage: mc-saver del-file <world> <number>...")
	}

	err = delFileFunc()
	if err != nil {
		return err
	}

	return nil
}

func replDelFile() (err error) {
	if len(subCmdArgs) < 2 {
		return formatError("syntax error, usage: del-file <number>...")
	}

	err = delFileFunc()
	if err != nil {
		return err
	}

	return nil
}

func delFileFunc() (err error) {
	worldDirPath = subCmdArgs[0]
	err = initWorldConfig()
	if err != nil {
		return err
	}

	indexSet, err := convertIntArray(subCmdArgs[1:])
	if err != nil {
		return formatError("parse command line args:", err)
	}

	config.File, err = deleteSliceElements(config.File, indexSet...)
	if err != nil {
		return formatError("delete element:", err)
	}

	err = saveConfig()
	if err != nil {
		return formatError("save config:", err)
	}

	return nil
}

func modFile() (err error) {
	if len(subCmdArgs) < 3 {
		return formatError("syntax error, usage: mc-saver mod-file <world> <number> <file_name>")
	}

	err = modFileFunc()
	if err != nil {
		return err
	}

	return nil
}

func replModFile() (err error) {
	if len(subCmdArgs) < 3 {
		return formatError("syntax error, usage: mod-file <number> <file_name>")
	}

	err = modFileFunc()
	if err != nil {
		return err
	}

	return nil
}

func modFileFunc() (err error) {
	worldDirPath = subCmdArgs[0]
	err = initWorldConfig()
	if err != nil {
		return err
	}

	index, err := strconv.ParseInt(subCmdArgs[1], 10, 32)
	if err != nil {
		return formatError("parse command line args:", err)
	}

	if index < 0 || int(index) >= len(config.File) {
		return formatError("index out of range:", index)
	}

	config.File[index] = subCmdArgs[2]

	err = saveConfig()
	if err != nil {
		return formatError("save config:", err)
	}

	return nil
}
