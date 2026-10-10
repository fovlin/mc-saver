package main

import (
	"encoding/json"
	"os"
	"path"

	"acovia.net/minecraft/save"
	"acovia.net/record"
)

func gencfg() error {
	if len(subCmdArgs) < 1 {
		return usageError("mc-saver gencfg <world>", "gencfg")
	}

	if useLegacyMode {
		defaultConfig.File = []string{
			"level.dat",
			"data",
			"datapacks",
			"advancements",
			"playerdata",
		}
	}

	worldDirPath := fmtPath(subCmdArgs[0])
	err := initWorld(worldDirPath)
	if err != nil {
		return record.FmtError("init world config:", err)
	}

	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	_, err = os.Stat(configFilePath)

	switch true {
	case os.IsNotExist(err):
		config = save.NewNullConfig()

	case !os.IsNotExist(err) && err == nil:
		err = initConfig(configFilePath)
		if err != nil {
			return record.FmtError("init config:", err)
		}

	default:
		return record.FmtError("stat config file:", err)
	}

	if len(config.File)+len(config.Dimension) != 0 {
		return record.FmtError("config not empty")
	}

	jsonData, err := json.MarshalIndent(defaultConfig, "", "\t")
	if err != nil {
		return record.FmtError("load default config struct:", err)
	}

	if err := os.WriteFile(configFilePath, jsonData, 0644); err != nil {
		return record.FmtError(err)
	}

	record.Info("created config file:", configFilePath)
	return nil
}
