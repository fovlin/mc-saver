package main

import (
	"acovia.net/record"
)

func listDms() error {
	usage := "mc-saver ls-dms <world>"
	replUsage := "ls-dms"
	_, _, err := loadWorld(1, usage, replUsage)
	if err != nil {
		return err
	}

	if len(config.Dimension) != 0 {
		record.Printf("dimension config:\n")
		for id := range config.Dimension {
			record.Printf("	- %v\n", id)
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
	usage := "mc-saver rm-dms <world> <dimension>..."
	replUsage := "rm-dms <dimension>..."
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
	usage := "mc-saver set-dms <world> <old_dimension> <new_dimension>"
	replUsage := "set-dms <old_dimension> <new_dimension>"
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
