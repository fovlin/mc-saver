package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"acovia.net/record"
)

func initWriter(archiveOutputPath string) (*Writer, error) {
	tempFile, err := os.CreateTemp(path.Dir(archiveOutputPath), "archive")
	if err != nil {
		return &Writer{}, fmt.Errorf("create temp archive: %v", err)
	}

	err = tempFile.Chmod(0644)
	if err != nil {
		cleanupErr := errors.Join(tempFile.Close(), os.Remove(tempFile.Name()))
		return &Writer{}, errors.Join(fmt.Errorf("init archive: %w", err), cleanupErr)
	}

	zipWriter := zip.NewWriter(tempFile)

	writer := &Writer{
		ZipWriter:   zipWriter,
		TempFile:    tempFile,
		ArchivePath: archiveOutputPath,
	}

	return writer, nil
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
	}, "adding:", headerName)
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

func initOutputPath(outputPath string, archiveName string) (string, error) {
	archiveStat, err := os.Stat(outputPath)
	switch true {

	case err == nil && archiveStat.IsDir():
		archiveName := fmt.Sprint(archiveName, "-", time.Now().Format(time.DateOnly), ".zip")
		newOutputPath := path.Join(outputPath, archiveName)
		return newOutputPath, nil

	case err == nil && !archiveStat.IsDir() || err != nil && os.IsNotExist(err):
		err = os.MkdirAll(path.Dir(outputPath), 0755)
		if err != nil {
			return "", record.FmtError("create directory:", err)
		}
		return outputPath, nil

	default:
		return "", fmt.Errorf("stat output path: %w", err)
	}
}

func (writer *Writer) Package() error {
	archiveTempPath := writer.TempFile.Name()
	err := writer.Close()
	if err != nil {
		return record.FmtError("close writer:", err)
	}

	writer.ArchivePath, err = addSubfixBeforeExt(writer.ArchivePath)
	if err != nil {
		return record.FmtError("add subfix:", err)
	}

	err = os.Rename(archiveTempPath, writer.ArchivePath)
	if err != nil {
		return record.FmtError("rename temp file:", err)
	}

	return nil
}

func (writer *Writer) Close() error {
	if writer.closed {
		return nil
	}
	writer.closed = true

	return errors.Join(writer.ZipWriter.Close(), writer.TempFile.Close())
}

func (writer *Writer) Clear() error {
	closeErr := writer.Close()

	removeErr := os.Remove(writer.TempFile.Name())
	if os.IsNotExist(removeErr) {
		removeErr = nil
	}

	switch true {
	case closeErr != nil && removeErr != nil:
		return record.FmtError("close writer:", closeErr, ">", "remove temp file:", removeErr)
	case closeErr != nil:
		return record.FmtError("close writer:", closeErr)
	case removeErr != nil:
		return record.FmtError("remove temp file:", removeErr)
	}

	return nil
}
