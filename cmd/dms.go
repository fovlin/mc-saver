package main

import (
	"fmt"
	"path"

	"acovia.net/record"
)

func listDms() error {
	if len(subCmdArgs) < 1 {
		return usageError("mc-saver list-dms <world>", "list-dms")
	}
	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
		return err
	}

	if len(config.Dimension) != 0 {
		fmt.Printf("dimension config:\n")
		for id := range config.Dimension {
			fmt.Printf("	- %v\n", id)
		}
	}

	return nil
}

func addDms() error {
	if len(subCmdArgs) < 2 {
		return usageError("mc-saver add-dms <world> <dimension>...", "add-dms <dimension>...")
	}
	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
		return err
	}

	for _, id := range subCmdArgs[1:] {
		if _, ok := config.Dimension[id]; ok {
			record.Warn(id+":", "dimension existed, skip")
			continue
		}
		config.Dimension[id] = defaultDimensionConfig
	}

	return commitConfig(configFilePath)
}

func delDms() error {
	if len(subCmdArgs) < 2 {
		return usageError("mc-saver del-dms <world> <dimension>...", "del-dms <dimension>...")
	}
	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
		return err
	}

	for _, id := range subCmdArgs[1:] {
		if _, ok := config.Dimension[id]; !ok {
			record.Warn(id+":", "dimension not found, skip")
			continue
		}
		delete(config.Dimension, id)
	}

	return commitConfig(configFilePath)
}

func modDms() error {
	if len(subCmdArgs) < 3 {
		return usageError("mc-saver mod-dms <world> <old_dimension> <new_dimension>", "mod-dms <old_dimension> <new_dimension>")
	}
	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
		return err
	}

	oldDimension, newDimension := subCmdArgs[1], subCmdArgs[2]
	if _, ok := config.Dimension[oldDimension]; !ok {
		return record.FmtError(oldDimension+":", "dimension not found")
	}
	if oldDimension == newDimension {
		return record.FmtError("dimension no change")
	}

	config.Dimension[newDimension] = config.Dimension[oldDimension]
	delete(config.Dimension, oldDimension)

	return commitConfig(configFilePath)
}
