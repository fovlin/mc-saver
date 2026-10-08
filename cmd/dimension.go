package main

import (
	"fmt"

	"acovia.net/record"
)

func addDms() (err error) {
	if len(subCmdArgs) < 2 {
		return formatError("syntax error, usage: mc-saver add-dms <world> <dimension>...")
	}

	err = addDmsFunc()
	if err != nil {
		return err
	}

	return nil
}

func replAddDms() (err error) {
	if len(subCmdArgs) < 2 {
		return formatError("syntax error, usage: add-dms <dimension>...")
	}

	err = addDmsFunc()
	if err != nil {
		return err
	}

	return nil
}

func addDmsFunc() (err error) {
	worldDirPath = subCmdArgs[0]
	err = initWorldConfig()
	if err != nil {
		return err
	}
	for _, v := range subCmdArgs[1:] {
		_, ok := config.Dimension[v]
		if ok {
			record.Warn(v+":", "dimension existed, skip")
			continue
		}
		config.Dimension[v] = defaultDimensionConfig
	}

	err = saveConfig()
	if err != nil {
		return formatError("save config:", err)
	}
	return nil
}

func delDms() (err error) {
	if len(subCmdArgs) < 2 {
		return formatError("syntax error, usage: mc-saver del-dms <world> <dimension>...")
	}

	err = delDmsFunc()
	if err != nil {
		return err
	}

	return nil
}

func replDelDms() (err error) {
	if len(subCmdArgs) < 2 {
		return formatError("syntax error, usage: del-dms <dimension>...")
	}

	err = delDmsFunc()
	if err != nil {
		return err
	}

	return nil
}

func delDmsFunc() (err error) {
	worldDirPath = subCmdArgs[0]
	err = initWorldConfig()
	if err != nil {
		return err
	}
	for _, v := range subCmdArgs[1:] {
		_, ok := config.Dimension[v]
		if !ok {
			record.Warn(v+":", "dimension not found, skip")
			continue
		}
		delete(config.Dimension, v)
	}

	err = saveConfig()
	if err != nil {
		return formatError("save config:", err)
	}
	return nil
}

func modDms() (err error) {
	if len(subCmdArgs) < 3 {
		return formatError("syntax error, usage: mc-saver mod-dms <world> <old_dimension> <dimension>")
	}

	err = modDmsFunc()
	if err != nil {
		return err
	}

	return nil
}

func replModDms() (err error) {
	if len(subCmdArgs) < 3 {
		return formatError("syntax error, usage: mod-dms <old_dimension> <dimension>")
	}

	err = modDmsFunc()
	if err != nil {
		return err
	}

	return nil
}

func modDmsFunc() (err error) {
	worldDirPath = subCmdArgs[0]
	err = initWorldConfig()
	if err != nil {
		return err
	}
	if _, ok := config.Dimension[subCmdArgs[1]]; !ok {
		return formatError(subCmdArgs[1]+":", "dimension not found")
	}

	if subCmdArgs[1] == subCmdArgs[2] {
		return formatError("dimension no change")
	}

	config.Dimension[subCmdArgs[2]] = config.Dimension[subCmdArgs[1]]
	delete(config.Dimension, subCmdArgs[1])

	err = saveConfig()
	if err != nil {
		return formatError("save config:", err)
	}
	return nil
}

func listDms() (err error) {
	if len(subCmdArgs) < 1 {
		return formatError("syntax error, usage: mc-saver list-dms <world>")
	}

	err = listDmsFunc()
	if err != nil {
		return err
	}

	return nil
}

func replListDms() (err error) {
	if len(subCmdArgs) < 1 {
		return formatError("syntax error, usage: list-dms")
	}

	err = listDmsFunc()
	if err != nil {
		return err
	}

	return nil
}

func listDmsFunc() (err error) {
	worldDirPath = subCmdArgs[0]
	err = initWorldConfig()
	if err != nil {
		return err
	}
	for id := range config.Dimension {
		fmt.Printf("- %v\n", id)
	}
	return nil
}
