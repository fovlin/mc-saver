package main

import (
	"fmt"
	"path"

	"acovia.net/minecraft/save"
	"acovia.net/record"
)

func listRange() error {
	if len(subCmdArgs) < 2 {
		return usageError("mc-saver list-range <world> <dimension>...", "list-range <dimension>...")
	}
	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
		return err
	}

	return printRangeRules(subCmdArgs[1:])
}

func printRangeRules(dimensions []string) error {
	for _, id := range dimensions {
		rule, ok := config.Dimension[id]
		if !ok {
			return record.FmtError(id+":", "dimension not found")
		}

		if len(rule.Range) != 0 {
			fmt.Printf("range config for %v:\n", id)
			for i, v := range rule.Range {
				fmt.Printf("	- %v: from: (%v, %v) to: (%v, %v)\n", i, v.From.X, v.From.Y, v.To.X, v.To.Y)
			}
		}
	}

	return nil
}

func addRange() error {
	if len(subCmdArgs) < 6 {
		return usageError("mc-saver add-range <world> <dimension> <from_x> <from_y> <to_x> <to_y>", "add-range <dimension> <from_x> <from_y> <to_x> <to_y>")
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

	newRangeConfig := save.RangeConfig{
		From: save.Coordinate{
			X: indexSet[0],
			Y: indexSet[1],
		},
		To: save.Coordinate{
			X: indexSet[2],
			Y: indexSet[3],
		},
	}

	id := subCmdArgs[1]
	dimension := config.Dimension[id]
	dimension.Range = append(dimension.Range, newRangeConfig)
	config.Dimension[id] = dimension

	return commitConfig(configFilePath)
}

func delRange() error {
	if len(subCmdArgs) < 3 {
		return usageError("mc-saver del-range <world> <dimension> <index>...", "del-range <dimension> <index>...")
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

	dimension.Range, err = deleteSliceElements(dimension.Range, indexSet...)
	if err != nil {
		return record.FmtError("delete element:", err)
	}
	config.Dimension[id] = dimension

	return commitConfig(configFilePath)
}

func modRange() error {
	if len(subCmdArgs) < 7 {
		return usageError("mc-saver mod-range <world> <dimension> <index> <from_x> <from_y> <to_x> <to_y>", "mod-range <dimension> <index> <from_x> <from_y> <to_x> <to_y>")
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
	if index < 0 || index >= len(dimension.Range) {
		return record.FmtError("index out of range:", index)
	}

	dimension.Range[index] = save.RangeConfig{
		From: save.Coordinate{
			X: indexSet[1],
			Y: indexSet[2],
		},
		To: save.Coordinate{
			X: indexSet[3],
			Y: indexSet[4],
		},
	}
	config.Dimension[id] = dimension

	return commitConfig(configFilePath)
}
