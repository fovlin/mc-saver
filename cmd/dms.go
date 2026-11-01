package main

import (
	"fmt"

	"acovia.net/record"
)

func listDms() error {
	usage := "mc-saver list-dms <world>"
	replUsage := "list-dms"
	_, _, err := loadWorld(1, usage, replUsage)
	if err != nil {
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
	usage := "mc-saver add-dms <world> <dimension>..."
	replUsage := "add-dms <dimension>..."
	_, configFilePath, err := loadWorld(2, usage, replUsage)
	if err != nil {
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
	usage := "mc-saver del-dms <world> <dimension>..."
	replUsage := "del-dms <dimension>..."
	_, configFilePath, err := loadWorld(2, usage, replUsage)
	if err != nil {
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
	usage := "mc-saver mod-dms <world> <old_dimension> <new_dimension>"
	replUsage := "mod-dms <old_dimension> <new_dimension>"
	_, configFilePath, err := loadWorld(3, usage, replUsage)
	if err != nil {
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
