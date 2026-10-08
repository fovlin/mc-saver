package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"acovia.net/minecraft/save"
	"acovia.net/record"
)

func main() {
	err := initProgram()
	if err != nil {
		record.Error(err)
	}

	configFilePath = strings.ReplaceAll(configFilePath, "\\", "/")
	function, ok := cmdMap[cmd]
	if !ok {
		record.Error("unknown command:", cmd)
	}

	subCmdArgs = loadSubCmdArgs()
	err = function()
	if err != nil {
		record.Error(err)
	}
}

func help() (err error) {
	fmt.Printf("%v", helpInfo)
	return nil
}

func initProgram() (err error) {
	flag.BoolFunc("l", "legacy world mode", func(s string) (err error) {
		useLegacyMode = true
		return nil
	})
	flag.BoolFunc("color", "enable color output", func(s string) (err error) {
		record.EnableColor = true
		return nil
	})
	flag.Parse()
	args = flag.Args()

	if len(args) < 1 {
		args = []string{"repl"}
	}

	cmd = args[0]

	return nil
}

func initWorld() error {
	worldDirStat, err := os.Stat(worldDirPath)
	if err != nil {
		return formatError("stat world directory:", err)
	}
	if !worldDirStat.IsDir() {
		return formatError(worldDirPath, "not a directory")
	}
	return nil
}

func initConfig() (err error) {
	if !fileIsExisted(configFilePath) {
		config = save.NullConfig
		err = saveConfig()
		if err != nil {
			formatError("save config:", err)
		}
	}

	config, err = save.LoadConfig(configFilePath)
	if err != nil {
		return formatError("load config:", err)
	}
	return nil
}
func initWorldConfig() (err error) {
	err = initWorld()
	if err != nil {
		return formatError("init world:", err)
	}

	configFilePath = path.Join(worldDirPath, configFileName)
	err = initConfig()
	if err != nil {
		return formatError("init config:", err)
	}

	return nil
}

func gencfg() (err error) {
	if len(subCmdArgs) < 1 {
		return formatError("syntax error, usage: mc-saver gencfg <world>")
	}

	err = gencfgFunc()
	if err != nil {
		return err
	}

	return nil
}

func replGencfg() (err error) {
	if len(subCmdArgs) < 1 {
		return formatError("syntax error, usage: gencfg")
	}

	err = gencfgFunc()
	if err != nil {
		return err
	}

	return nil
}

func gencfgFunc() (err error) {
	if useLegacyMode {
		defaultConfig.File = []string{
			"level.dat",
			"data",
			"datapacks",
			"advancements",
			"playerdata",
		}
	}

	worldDirPath = subCmdArgs[0]
	configFilePath = path.Join(worldDirPath, configFileName)

	_, err = os.Stat(configFilePath)
	if !os.IsNotExist(err) {
		return formatError(configFilePath, "already exists")
	}

	jsonData, err := json.MarshalIndent(defaultConfig, "", "	")
	if err != nil {
		return formatError("load default config struct:", err)
	}

	err = os.WriteFile(configFilePath, jsonData, 0644)
	if err != nil {
		return formatError(err)
	}
	record.Info("created config file:", configFilePath)
	return nil
}

func run() (err error) {
	if len(subCmdArgs) < 1 {
		return formatError("syntax error, usage: mc-saver run <world> [output]")
	}

	err = runFunc()
	if err != nil {
		return err
	}

	return nil
}

func replRun() (err error) {
	if len(subCmdArgs) < 1 {
		return formatError("syntax error, usage: run [output]")
	}

	err = runFunc()
	if err != nil {
		return err
	}

	return nil
}

func runFunc() (err error) {
	if len(subCmdArgs) > 1 {
		outputPath = subCmdArgs[1]
	}

	worldDirPath = strings.ReplaceAll(worldDirPath, "\\", "/")
	outputPath = strings.ReplaceAll(outputPath, "\\", "/")

	worldDirPath = subCmdArgs[0]
	err = initWorldConfig()
	if err != nil {
		return err
	}

	root, err = os.OpenRoot(worldDirPath)
	if err != nil {
		return formatError("open world directory:", err)
	}

	zipWriter, _, end, err := initZipWriter()
	if err != nil {
		return formatError("init zip writer:", err)
	}

	if useLegacyMode {
		if err := save.SaveOldAllFile(root, config, zipWriter, saveFile); err != nil {
			if err := end(err); err != nil {
				return formatError(err)
			}
		}
	} else {
		if err := save.SaveAllFile(root, config, zipWriter, saveFile); err != nil {
			if err := end(err); err != nil {
				return formatError(err)
			}
		}
	}

	if err := end(nil); err != nil {
		return formatError("close file writer:", err)
	}

	return nil
}

func listConfig() (err error) {
	if len(subCmdArgs) < 2 {
		return formatError("syntax error, usage: mc-saver list-config <world> <dimension>...")
	}
	err = listConfigFunc()
	if err != nil {
		return err
	}

	return nil
}

func replListConfig() (err error) {
	if len(subCmdArgs) < 2 {
		return formatError("syntax error, usage: list-config <dimension>...")
	}
	err = listConfigFunc()
	if err != nil {
		return err
	}

	return nil
}

func listConfigFunc() (err error) {
	listRange()
	listSimple()
	return nil
}

func list() (err error) {
	if len(subCmdArgs) < 1 {
		return formatError("syntax error, usage: mc-saver list <world>")
	}

	err = listFunc()
	if err != nil {
		return err
	}

	return nil
}

func replList() (err error) {
	err = listFunc()
	if err != nil {
		return err
	}

	return nil
}

func listFunc() (err error) {
	worldDirPath = subCmdArgs[0]
	err = initWorldConfig()
	if err != nil {
		return err
	}
	if len(config.Dimension) == 0 {
		fmt.Println("no dimension config at all")
	}

	for id, rule := range config.Dimension {
		fmt.Printf("range config for %v:\n", id)
		if len(rule.Range) == 0 {
			fmt.Printf("no range config for %v\n", id)
		}
		for i, v := range rule.Range {
			fmt.Printf("- %v: from: (%v, %v) to: (%v, %v)\n", i, v.From.X, v.From.Y, v.To.X, v.To.Y)
		}

		fmt.Printf("simple config for %v:\n", id)
		if len(rule.Simple) == 0 {
			fmt.Printf("no simple config for %v\n", id)
		}
		for i, v := range rule.Simple {
			fmt.Printf("- %v: (%v, %v)\n", i, v.X, v.Y)
		}
	}

	fmt.Println("file config:")
	listFile()
	return nil
}
