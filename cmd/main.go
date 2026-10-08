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

func initWorldConfig() (err error) {
	initConfigFilePath()
	config, err = save.LoadConfig(configFilePath)
	if err != nil {
		return fmtErr("load config:", err)
	}

	return nil
}

func initConfigFilePath() {
	worldDirPath = subCmdArgs[0]
	configFilePath = path.Join(worldDirPath, configFileName)
}

func gencfg() (err error) {
	if len(subCmdArgs) < 1 {
		return fmtErr("syntax error, usage: mc-saver gencfg <world>")
	}

	initConfigFilePath()

	if useLegacyMode {
		defaultConfig.File = []string{
			"level.dat",
			"data",
			"datapacks",
			"advancements",
			"playerdata",
		}
	}

	if _, err := os.Stat(configFilePath); !os.IsNotExist(err) {
		return fmtErr("generate config file:", "'"+configFilePath+"'", "already exists", configFilePath)
	}

	jsonData, err := json.MarshalIndent(defaultConfig, "", "	")
	if err != nil {
		return fmtErr("load default config struct:", err)
	}

	err = os.WriteFile(configFilePath, jsonData, 0644)
	if err != nil {
		return fmtErr(err)
	}
	record.Info("created config file:", configFilePath)
	return nil
}

func run() (err error) {
	if len(subCmdArgs) < 1 {
		return fmtErr("syntax error, usage: mc-saver run <world> [output]")
	}

	if len(subCmdArgs) > 1 {
		outputPath = subCmdArgs[1]
	}

	worldDirPath = strings.ReplaceAll(worldDirPath, "\\", "/")
	outputPath = strings.ReplaceAll(outputPath, "\\", "/")

	err = initWorldConfig()
	if err != nil {
		return err
	}

	root, err = os.OpenRoot(worldDirPath)
	if err != nil {
		return fmtErr("open world directory:", err)
	}

	zipWriter, _, end, err := initZipWriter()
	if err != nil {
		return fmtErr("init zip writer:", err)
	}

	if useLegacyMode {
		if err := save.SaveOldAllFile(root, config, zipWriter, saveFile); err != nil {
			if err := end(err); err != nil {
				return fmtErr(err)
			}
		}
	} else {
		if err := save.SaveAllFile(root, config, zipWriter, saveFile); err != nil {
			if err := end(err); err != nil {
				return fmtErr(err)
			}
		}
	}

	if err := end(nil); err != nil {
		return fmtErr("close file writer:", err)
	}

	return nil
}

func addDms() (err error) {
	if len(subCmdArgs) < 2 {
		return fmtErr("syntax error, usage: mc-saver add-dms <world> <dimension>...")
	}

	err = initWorldConfig()
	if err != nil {
		return err
	}
	for _, v := range subCmdArgs[1:] {
		_, ok := config.Dimension[v]
		if ok {
			record.Warn(v+":", "dimension existed, skip")
			continue
		}
		config.Dimension[v] = defaultDimensionConfig
	}

	err = saveConfig()
	if err != nil {
		return fmtErr("save config:", err)
	}
	return nil
}

func delDms() (err error) {
	if len(subCmdArgs) < 2 {
		return fmtErr("syntax error, usage: mc-saver del-dms <world> <dimension>...")
	}

	err = initWorldConfig()
	if err != nil {
		return err
	}
	for _, v := range subCmdArgs[1:] {
		_, ok := config.Dimension[v]
		if !ok {
			record.Warn(v+":", "dimension not found, skip")
			continue
		}
		delete(config.Dimension, v)
	}

	err = saveConfig()
	if err != nil {
		return fmtErr("save config:", err)
	}
	return nil
}

func modDms() (err error) {
	if len(subCmdArgs) < 3 {
		return fmtErr("syntax error, usage: mc-saver mod-dms <world> <old_dimension> <dimension>")
	}

	err = initWorldConfig()
	if err != nil {
		return err
	}
	if _, ok := config.Dimension[subCmdArgs[1]]; !ok {
		return fmtErr(subCmdArgs[1]+":", "dimension not found")
	}

	if subCmdArgs[1] == subCmdArgs[2] {
		return fmtErr("dimension no change")
	}

	config.Dimension[subCmdArgs[2]] = config.Dimension[subCmdArgs[1]]
	delete(config.Dimension, subCmdArgs[1])

	err = saveConfig()
	if err != nil {
		return fmtErr("save config:", err)
	}
	return nil
}

func listConfig() (err error) {
	if len(subCmdArgs) < 2 {
		return fmtErr("syntax error, usage: mc-saver list-config <world> <dimension>...")
	}
	listRange()
	listSimple()
	return nil
}

func list() (err error) {
	if len(subCmdArgs) < 1 {
		return fmtErr("syntax error, usage: mc-saver list <world>")
	}

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

func listDms() (err error) {
	if len(subCmdArgs) < 1 {
		return fmtErr("syntax error, usage: mc-saver list-dms <world>")
	}

	err = initWorldConfig()
	if err != nil {
		return err
	}
	for id := range config.Dimension {
		fmt.Printf("- %v\n", id)
	}
	return nil
}
