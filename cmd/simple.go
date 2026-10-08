package main

import (
	"fmt"

	"acovia.net/minecraft/save"
)

func listSimple() (err error) {
	if len(subCmdArgs) < 2 {
		return formatError("syntax error, usage: mc-saver list-simple <world> <dimension>...")
	}

	err = listSimpleFunc()
	if err != nil {
		return err
	}

	return nil
}

func replListSimple() (err error) {
	if len(subCmdArgs) < 2 {
		return formatError("syntax error, usage: list-simple <dimension>...")
	}

	err = listSimpleFunc()
	if err != nil {
		return err
	}

	return nil
}

func listSimpleFunc() (err error) {
	worldDirPath = subCmdArgs[0]
	err = initWorldConfig()
	if err != nil {
		return err
	}

	for _, id := range subCmdArgs[1:] {
		_, ok := config.Dimension[id]
		if !ok {
			return formatError(id+":", "dimension not found")
		}

		simpleConfig := config.Dimension[id].Simple
		if len(simpleConfig) == 0 {
			fmt.Printf("no simple config for %v\n", id)
			continue
		}
		fmt.Printf("simple config for %v:\n", id)
		for i, v := range simpleConfig {
			fmt.Printf("- %v: (%v, %v)\n", i, v.X, v.Y)
		}
	}

	return nil
}

func addSimple() (err error) {
	if len(subCmdArgs) < 4 {
		return formatError("syntax error, usage: mc-saver add-simple <world> <dimension> <x> <y>")
	}

	err = addSimpleFunc()
	if err != nil {
		return err
	}

	return nil
}

func replAddSimple() (err error) {
	if len(subCmdArgs) < 4 {
		return formatError("syntax error, usage: add-simple <dimension> <x> <y>")
	}

	err = addSimpleFunc()
	if err != nil {
		return err
	}

	return nil
}

func addSimpleFunc() (err error) {
	worldDirPath = subCmdArgs[0]
	err = initWorldConfig()
	if err != nil {
		return err
	}

	indexSet, err := convertIntArray(subCmdArgs[2:])
	if err != nil {
		return formatError("parse command line args:", err)
	}

	newCoordinate := save.Coordinate{
		X: indexSet[0],
		Y: indexSet[1],
	}

	dimension, _ := config.Dimension[subCmdArgs[1]]
	dimension.Simple = append(dimension.Simple, newCoordinate)

	config.Dimension[subCmdArgs[1]] = dimension

	err = saveConfig()
	if err != nil {
		return formatError("save config:", err)
	}

	return nil
}

func delSimple() (err error) {
	if len(subCmdArgs) < 3 {
		return formatError("syntax error, usage: mc-saver del-simple <world> <dimension> <number>...")
	}

	err = delSimpleFunc()
	if err != nil {
		return err
	}

	return nil
}

func replDelSimple() (err error) {
	if len(subCmdArgs) < 3 {
		return formatError("syntax error, usage: del-simple <dimension> <number>...")
	}

	err = delSimpleFunc()
	if err != nil {
		return err
	}

	return nil
}

func delSimpleFunc() (err error) {
	worldDirPath = subCmdArgs[0]
	err = initWorldConfig()
	if err != nil {
		return err
	}

	dimension, ok := config.Dimension[subCmdArgs[1]]
	if !ok {
		return formatError(subCmdArgs[1]+":", "dimension not found")
	}

	var indexSet []int

	indexSet, err = convertIntArray(subCmdArgs[2:])
	if err != nil {
		return formatError("parse command line args:", err)
	}

	dimension.Simple, err = deleteSliceElements(dimension.Simple, indexSet...)
	if err != nil {
		return formatError("delete element:", err)
	}

	config.Dimension[subCmdArgs[1]] = dimension

	err = saveConfig()
	if err != nil {
		return formatError("save config:", err)
	}

	return nil
}

func modSimple() (err error) {
	if len(subCmdArgs) < 5 {
		return formatError("syntax error, usage: mc-saver mod-simple <world> <dimension> <number> <x> <y>")
	}

	err = modSimpleFunc()
	if err != nil {
		return err
	}

	return nil
}

func replModSimple() (err error) {
	if len(subCmdArgs) < 5 {
		return formatError("syntax error, usage: mod-simple <dimension> <number> <x> <y>")
	}

	err = modSimpleFunc()
	if err != nil {
		return err
	}

	return nil
}

func modSimpleFunc() (err error) {
	worldDirPath = subCmdArgs[0]
	err = initWorldConfig()
	if err != nil {
		return err
	}

	_, ok := config.Dimension[subCmdArgs[1]]
	if !ok {
		return formatError(subCmdArgs[1]+":", "dimension not found")
	}

	indexSet, err := convertIntArray(subCmdArgs[2:])
	if err != nil {
		return formatError("parse command line args:", err)
	}

	newCoordinate := save.Coordinate{
		X: indexSet[1],
		Y: indexSet[2],
	}

	if indexSet[0] < 0 || indexSet[0] >= len(config.Dimension[subCmdArgs[1]].Simple) {
		return formatError("index out of range:", indexSet[0])
	}

	config.Dimension[subCmdArgs[1]].Simple[indexSet[0]] = newCoordinate

	err = saveConfig()
	if err != nil {
		return formatError("save config:", err)
	}
	return nil
}
