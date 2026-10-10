package main

import (
	"fmt"
	"path"

	"acovia.net/minecraft/save"
	"acovia.net/record"
)

func listSimple() error {
	if len(subCmdArgs) < 2 {
		return usageError("mc-saver list-simple <world> <dimension>...", "list-simple <dimension>...")
	}
	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
		return err
	}

	return printSimpleRules(subCmdArgs[1:])
}

// printSimpleRules prints the simple rules of every given dimension.
func printSimpleRules(dimensions []string) error {
	for _, id := range dimensions {
		rule, ok := config.Dimension[id]
		if !ok {
			return record.FmtError(id+":", "dimension not found")
		}

		if len(rule.Simple) == 0 {
			continue
		}

		fmt.Printf("simple config for %v:\n", id)
		for i, v := range rule.Simple {
			fmt.Printf("	- %v: (%v, %v)\n", i, v.X, v.Y)
		}
	}

	return nil
}

func addSimple() error {
	if len(subCmdArgs) < 4 {
		return usageError("mc-saver add-simple <world> <dimension> <x> <y>", "add-simple <dimension> <x> <y>")
	}
	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
		return err
	}

	indexSet, err := convertIntArray(subCmdArgs[2:])
	if err != nil {
		return record.FmtError("parse command line args:", err)
	}

	newCoordinate := save.Coordinate{
		X: indexSet[0],
		Y: indexSet[1],
	}

	id := subCmdArgs[1]
	dimension := config.Dimension[id]
	dimension.Simple = append(dimension.Simple, newCoordinate)
	config.Dimension[id] = dimension

	return commitConfig(configFilePath)
}

func delSimple() error {
	if len(subCmdArgs) < 3 {
		return usageError("mc-saver del-simple <world> <dimension> <index>...", "del-simple <dimension> <index>...")
	}
	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
		return err
	}

	id := subCmdArgs[1]
	dimension, ok := config.Dimension[id]
	if !ok {
		return record.FmtError(id+":", "dimension not found")
	}

	indexSet, err := convertIntArray(subCmdArgs[2:])
	if err != nil {
		return record.FmtError("parse command line args:", err)
	}

	dimension.Simple, err = deleteSliceElements(dimension.Simple, indexSet...)
	if err != nil {
		return record.FmtError("delete element:", err)
	}
	config.Dimension[id] = dimension

	return commitConfig(configFilePath)
}

func modSimple() error {
	if len(subCmdArgs) < 5 {
		return usageError("mc-saver mod-simple <world> <dimension> <index> <x> <y>", "mod-simple <dimension> <index> <x> <y>")
	}
	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
		return err
	}

	id := subCmdArgs[1]
	dimension, ok := config.Dimension[id]
	if !ok {
		return record.FmtError(id+":", "dimension not found")
	}

	indexSet, err := convertIntArray(subCmdArgs[2:])
	if err != nil {
		return record.FmtError("parse command line args:", err)
	}

	index := indexSet[0]
	if index < 0 || index >= len(dimension.Simple) {
		return record.FmtError("index out of range:", index)
	}

	dimension.Simple[index] = save.Coordinate{
		X: indexSet[1],
		Y: indexSet[2],
	}
	config.Dimension[id] = dimension

	return commitConfig(configFilePath)
}
