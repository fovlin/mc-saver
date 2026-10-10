package main

import (
	"fmt"
	"path"
)

func list() error {
	if len(subCmdArgs) < 1 {
		return usageError("mc-saver list <world>", "list")
	}

	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
		return err
	}

	if len(config.Dimension) == 0 {
		fmt.Println("no dimension config at all")
	}

	for id, rule := range config.Dimension {
		if len(rule.Range) != 0 {
			fmt.Printf("range config for %v:\n", id)
			for i, v := range rule.Range {
				fmt.Printf("	- %v: from: (%v, %v) to: (%v, %v)\n", i, v.From.X, v.From.Y, v.To.X, v.To.Y)
			}
		}

		if len(rule.Simple) != 0 {
			fmt.Printf("simple config for %v:\n", id)
			for i, v := range rule.Simple {
				fmt.Printf("	- %v: (%v, %v)\n", i, v.X, v.Y)
			}
		}
	}

	printFileRules()
	return nil
}

func listConfig() error {
	if len(subCmdArgs) < 2 {
		return usageError("mc-saver list-config <world> <dimension>...", "list-config <dimension>...")
	}
	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
		return err
	}

	if err := printRangeRules(subCmdArgs[1:]); err != nil {
		return err
	}

	return printSimpleRules(subCmdArgs[1:])
}
