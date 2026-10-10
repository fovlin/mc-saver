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

	var saveErr error
	if useLegacyMode {
		saveErr = save.SaveOldAllFile(root, config, writer.ZipWriter, writeFile)
	} else {
		saveErr = save.SaveAllFile(root, config, writer.ZipWriter, writeFile)
	}

	if saveErr != nil {
		err = writer.Clear()
		if err != nil {
			return record.FmtError("clear damage file:", err)
		}
		return record.FmtError("backup:", saveErr)
	}

	err = writer.Package()
	if err != nil {
		return record.FmtError("package archive file:", err)
	}

	record.Info("zip writer closed")
	record.Info("backup completed successfully!")

	return nil
}
