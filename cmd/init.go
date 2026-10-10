package main

import (
	"os"

	"acovia.net/minecraft/save"
	"acovia.net/record"
)

func initWorldConfig(worldDirPath string, configFilePath string) error {
	if err := initWorld(worldDirPath); err != nil {
		return record.FmtError("init world:", err)
	}

	if err := initConfig(configFilePath); err != nil {
		return record.FmtError("init config:", err)
	}

	return nil
}

func initWorld(worldDirPath string) error {
	worldDirStat, err := os.Stat(worldDirPath)
	if err != nil {
		return record.FmtError("stat world directory:", err)
	}
	if !worldDirStat.IsDir() {
		return record.FmtError(worldDirPath, "not a directory")
	}
	return nil
}

func initConfig(configFilePath string) error {
	_, err := os.Stat(configFilePath)

	switch true {
	case os.IsNotExist(err):
		config = save.NewNullConfig()

	case err != nil:
		return record.FmtError("stat config:", err)

	default:
		loaded, err := save.LoadConfig(configFilePath)
		if err != nil {
			return record.FmtError("load config:", err)
		}
		config = loaded
	}

	return nil
}
