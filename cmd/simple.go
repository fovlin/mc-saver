package main

import (
	"fmt"

	"acovia.net/minecraft/save"
)

func listSimple() (err error) {
	if len(subCmdArgs) < 2 {
		return fmtErr("syntax error, usage: mc-saver list-simple <world> <dimension>...")
	}

	err = initWorldConfig()
	if err != nil {
		return err
	}

	for _, id := range subCmdArgs[1:] {
		_, ok := config.Dimension[id]
		if !ok {
			return fmtErr(id+":", "dimension not found")
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
		return fmtErr("syntax error, usage: mc-saver add-simple <world> <dimension> <x> <y>")
	}

	err = initWorldConfig()
	if err != nil {
		return err
	}

	indexSet, err := convertIntArray(subCmdArgs[2:])
	if err != nil {
		return fmtErr("parse command line args:", err)
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
		return fmtErr("save config:", err)
	}
	return nil
}

func delSimple() (err error) {
	if len(subCmdArgs) < 3 {
		return fmtErr("syntax error, usage: mc-saver del-simple <world> <dimension> <number>...")
	}

	err = initWorldConfig()
	if err != nil {
		return err
	}

	dimension, ok := config.Dimension[subCmdArgs[1]]
	if !ok {
		return fmtErr(subCmdArgs[1]+":", "dimension not found")
	}

	var indexSet []int

	indexSet, err = convertIntArray(subCmdArgs[2:])
	if err != nil {
		return fmtErr("parse command line args:", err)
	}

	dimension.Simple, err = deleteSliceElements(dimension.Simple, indexSet...)
	if err != nil {
		return fmtErr("delete element:", err)
	}

	config.Dimension[subCmdArgs[1]] = dimension

	err = saveConfig()
	if err != nil {
		return fmtErr("save config:", err)
	}

	return nil
}

func modSimple() (err error) {
	if len(subCmdArgs) < 5 {
		return fmtErr("syntax error, usage: mc-saver mod-simple <world> <dimension> <number> <x> <y>")
	}

	err = initWorldConfig()
	if err != nil {
		return err
	}

	_, ok := config.Dimension[subCmdArgs[1]]
	if !ok {
		return fmtErr(subCmdArgs[1]+":", "dimension not found")
	}

	indexSet, err := convertIntArray(subCmdArgs[2:])
	if err != nil {
		return fmtErr("parse command line args:", err)
	}

	newCoordinate := save.Coordinate{
		X: indexSet[1],
		Y: indexSet[2],
	}

	if indexSet[0] < 0 || indexSet[0] >= len(config.Dimension[subCmdArgs[1]].Simple) {
		return fmtErr("index out of range:", indexSet[0])
	}

	config.Dimension[subCmdArgs[1]].Simple[indexSet[0]] = newCoordinate

	err = saveConfig()
	if err != nil {
		return fmtErr("save config:", err)
	}
	return nil
}
