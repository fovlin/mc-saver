package main

import (
	"fmt"
)

func list() error {
	usage := "mc-saver list <world>"
	replUsage := "list"
	_, _, err := loadWorld(1, usage, replUsage)
	if err != nil {
		return err
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
	usage := "mc-saver list-config <world> <dimension>..."
	replUsage := "list-config <dimension>..."
	_, _, err := loadWorld(2, usage, replUsage)
	if err != nil {
		return err
	}

	if err := printRangeRules(subCmdArgs[1:]); err != nil {
		return err
	}

	return printSimpleRules(subCmdArgs[1:])
}
