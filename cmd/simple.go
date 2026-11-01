package main

import (
	"fmt"

	"acovia.net/minecraft/save"
	"acovia.net/record"
)

func listSimple() error {
	usage := "mc-saver list-simple <world> <dimension>..."
	replUsage := "list-simple <dimension>..."
	_, _, err := loadWorld(2, usage, replUsage)
	if err != nil {
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
	usage := "mc-saver add-simple <world> <dimension> <x> <y>"
	replUsage := "add-simple <dimension> <x> <y>"
	_, configFilePath, err := loadWorld(4, usage, replUsage)
	if err != nil {
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
	usage := "mc-saver del-simple <world> <dimension> <index>..."
	replUsage := "del-simple <dimension> <index>..."
	_, configFilePath, err := loadWorld(3, usage, replUsage)
	if err != nil {
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
	usage := "mc-saver mod-simple <world> <dimension> <index> <x> <y>"
	replUsage := "mod-simple <dimension> <index> <x> <y>"
	_, configFilePath, err := loadWorld(5, usage, replUsage)
	if err != nil {
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
