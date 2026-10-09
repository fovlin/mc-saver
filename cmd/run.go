package main

import (
	"archive/zip"
	"os"
	"path"

	"acovia.net/minecraft/save"
	"acovia.net/record"
)

func run() error {
	if len(subCmdArgs) < 1 {
		return usageError("mc-saver run <world> [output]", "run [output]")
	}
	if len(subCmdArgs) > 1 {
		outputPath = subCmdArgs[1]
	}

	worldDirPath := fmtPath(subCmdArgs[0])
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
		return err
	}

	root, err := os.OpenRoot(worldDirPath)
	if err != nil {
		return record.FmtError("open world directory:", err)
	}
	defer root.Close()

	zipWriter, _, end, err := initZipWriter(worldDirPath)
	if err != nil {
		return record.FmtError("init zip writer:", err)
	}

	writeFile := func(fileName string, zipWriter *zip.Writer) error {
		return writeFileToZip(root, fileName, zipWriter)
	}

	var saveErr error
	if useLegacyMode {
		saveErr = save.SaveOldAllFile(root, config, zipWriter, writeFile)
	} else {
		saveErr = save.SaveAllFile(root, config, zipWriter, writeFile)
	}

	if saveErr != nil {
		return record.FmtError(end(saveErr))
	}

	if err := end(nil); err != nil {
		return record.FmtError("close file writer:", err)
	}

	return nil
}
