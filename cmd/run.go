package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"os"
	"path"

	"acovia.net/minecraft/save"
	"acovia.net/record"
)

func run() error {
	usage := "mc-saver run <world> [output]"
	replUsage := "run [output]"

	worldDirPath, _, err := loadWorld(usage, replUsage)
	if err != nil {
		return err
	}

	var outputPath = fmtPath(defaultOutputPath)

	if len(subCmdArgs) > 1 {
		outputPath = fmtPath(subCmdArgs[1])
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

	err = func() error {
		writeFile := func(fileName string, zipWriter *zip.Writer) error {
			return writeFileToZip(root, fileName, zipWriter)
		}

		if useLegacyMode {
			err = save.SaveOldAllFile(root, config, writer.ZipWriter, writeFile)
		} else {
			err = save.SaveAllFile(root, config, writer.ZipWriter, writeFile)
		}

		if err != nil {
			return record.FmtError("backup:", err)
		}

		err = writer.Package()
		if err != nil {
			return record.FmtError("package archive:", err)
		}

		return nil
	}()

	if err != nil {
		if clearErr := writer.Clear(); clearErr != nil {
			return errors.Join(err, fmt.Errorf("remove damage file: %w", clearErr))
		}
		return err
	}

	record.Info("backup completed successfully")
	return nil
}
