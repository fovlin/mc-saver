package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"acovia.net/record"
)

func initZipWriter(outputPath string) (*zip.Writer, func(error) error, error) {
	archiveDirPath, err := initOutputDir(outputPath)
	if err != nil {
		return nil, nil, fmt.Errorf("format output path: %v", err)
	}

	file, err := os.CreateTemp(path.Dir(archiveDirPath), "archive")
	if err != nil {
		return nil, nil, fmt.Errorf("create temp archive: %v", err)
	}

	err = file.Chmod(0644)
	if err != nil {
		return nil, nil, record.FmtError("init archive:", err)
	}



	zipWriter := zip.NewWriter(file)

	return zipWriter, closeFunc(zipWriter, file, path.Base(outputPath)), nil
}

func writeFileToZip(root *os.Root, fileName string, zipWriter *zip.Writer) error {
	fileReader, err := root.Open(fileName)
	if os.IsNotExist(err) {
		record.Warn("skip file:", err)
		return nil
	} else if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer fileReader.Close()

	fileInfo, err := root.Stat(fileName)
	if err != nil {
		return fmt.Errorf("read file info: %v", err)
	}

	zipFileHeader, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	headerName := path.Join(path.Base(root.Name()), fileName)
	zipFileHeader.Name = headerName
	zipFileHeader.Method = zip.Deflate

	file, err := zipWriter.CreateHeader(zipFileHeader)
	if err != nil {
		return fmt.Errorf("create file: %v", err)
	}

	err = record.RunningInfo(func() error {
		if _, err := io.Copy(file, fileReader); err != nil {
			return fmt.Errorf("write file: %v", err)
		}
		return nil
	}, "adding: ", headerName)
	if err != nil {
		return err
	}

	return nil
}

func saveConfig(configFilePath string) error {
	jsonData, err := json.MarshalIndent(config, "", "	")
	if err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	err = os.WriteFile(configFilePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func commitConfig(configFilePath string) error {
	if err := saveConfig(configFilePath); err != nil {
		return record.FmtError("save config:", err)
	}

	return nil
}

func initOutputPath(archivePath string) (string, error) {
	archiveStat, err := os.Stat(archivePath)
	switch true {

	case archiveStat.IsDir():
		archiveName := fmt.Sprint(path.Base(archivePath), "-", time.Now().Format(time.DateOnly))
		outputPath, err := addSubfixBeforeExt(path.Join(archivePath, archiveName))
		if err != nil {
			return "", err
		}
		return outputPath, nil

	case err != nil:
		return "", err

	default:
		outputPath, err := addSubfixBeforeExt(archivePath)
		if err != nil {
			return "", err
		}
		return outputPath, nil
	}
}

func closeFunc(zipWriter *zip.Writer, file *os.File, outputPath string) func(error) error {
	close := func(inputErr error) error {
		archivePath := file.Name()
		err := zipWriter.Close()
		if err != nil {
			return err
		}
		err = file.Close()
		if err != nil {
			return err
		}

		if inputErr != nil {
			err := os.Remove(archivePath)
			if err != nil {
				return record.FmtError("remove damage archive:", err)
			}
		}

		outputPath, err = initOutputPath(outputPath)
		if err != nil {
			return record.FmtError("init output path:", err)
		}

		err = os.Rename(archivePath, outputPath)
		if err != nil {
			return err
		}

		return nil
	}
	return close
}