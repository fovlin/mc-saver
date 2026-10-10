package main

import (
	"archive/zip"
	"os"
	"path"
	"path/filepath"

	"acovia.net/minecraft/save"
	"acovia.net/record"
)

func run() error {
	if len(subCmdArgs) < 1 {
		return usageError("mc-saver run <world> [output]", "run [output]")
	}

	var outputPath = fmtPath(defaultOutputPath)

	if len(subCmdArgs) > 1 {
		outputPath = fmtPath(subCmdArgs[1])
	}

	worldDirPath, err := filepath.Abs(subCmdArgs[0])
	if err != nil {
		return record.FmtError("load world absolute path:", err)
	}

	worldDirPath = fmtPath(worldDirPath)
	configFilePath := fmtPath(path.Join(worldDirPath, configFileName))
	err = initWorldConfig(worldDirPath, configFilePath)
	if err != nil {
		return record.FmtError("init world config:", err)
	}

	root, err := os.OpenRoot(worldDirPath)
	if err != nil {
		return record.FmtError("open world directory:", err)
	}
	defer root.Close()

	outputPath, err = initOutputPath(outputPath, path.Base(worldDirPath))
	if err != nil {
		return record.FmtError("init output path:", err)
	}

	writer, err := initWriter(outputPath)
	if err != nil {
		return record.FmtError("init zip writer:", err)
	}

	writeFile := func(fileName string, zipWriter *zip.Writer) error {
		return writeFileToZip(root, fileName, zipWriter)
	}

	if useLegacyMode {
		err = save.SaveOldAllFile(root, config, writer.ZipWriter, writeFile)
	} else {
		err = save.SaveAllFile(root, config, writer.ZipWriter, writeFile)
	}

	if err != nil {
		clearErr := writer.Clear()
		if clearErr != nil {
			return record.FmtError(err, ">", "clear damage file:", clearErr)
		}
		return record.FmtError("backup:", err)
	}

	err = writer.Package()
	if err != nil {
		return record.FmtError("package archive:", err)
	}

	record.Info("backup completed successfully")

	return nil
}
