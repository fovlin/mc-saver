package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"acovia.net/minecraft/save"
	"acovia.net/record"
)

var (
	cmd        string
	config     save.Config = save.NullConfig
	subCmdArgs []string

	useLegacyMode  bool
	replMode       bool
	configFileName string = "saver.json"
	outputPath     string = "."

	defaultDimensionConfig = save.DimensionConfig{
		Range: []save.RangeConfig{
			{
				From: save.Coordinate{
					X: -1,
					Y: -1,
				},
				To: save.Coordinate{
					X: 0,
					Y: 0,
				},
			},
		},
	}

	defaultConfig = save.Config{
		Dimension: map[string]save.DimensionConfig{
			"minecraft:overworld":  defaultDimensionConfig,
			"minecraft:the_nether": defaultDimensionConfig,
			"minecraft:the_end":    defaultDimensionConfig,
		},
		File: []string{
			"level.dat",
			"data",
			"datapacks",
			"players",
		},
	}
)

func usageError(argsLine string, replArgsLine string) error {
	var usage string
	switch true {
	case replMode:
		usage = replArgsLine

	case !replMode:
		usage = argsLine
	}
	return record.FmtError("syntax error, usage:", usage)
}

func formatOutputPath(worldDirPath string, outputPath string) (string, error) {
	worldAbsPath, err := filepath.Abs(worldDirPath)
	if err != nil {
		return "", fmt.Errorf("load absolute path: %w", err)
	}
	worldDirName := filepath.Base(worldAbsPath)

	outputFileInfo, err := os.Stat(outputPath)
	switch true {

	case os.IsNotExist(err):
		err = os.MkdirAll(filepath.Dir(outputPath), 0755)
		if err != nil {
			return "", fmt.Errorf("create output directory: %v", err)
		}
		return outputPath, nil

	case err != nil:
		return "", fmt.Errorf("read file info: %v", err)

	case outputFileInfo.IsDir():
		archiveFilePath := filepath.Join(outputPath, worldDirName+"-"+time.Now().Format(time.DateOnly)+".zip")
		archiveFilePath, err = addSubfixBeforeExt(archiveFilePath)
		if err != nil {
			return "", fmt.Errorf("add subfix: %v", err)
		}
		return archiveFilePath, nil

	default:
		archiveFilePath, err := addSubfixBeforeExt(outputPath)
		if err != nil {
			return "", fmt.Errorf("add subfix: %v", err)
		}
		return archiveFilePath, nil
	}
}

func addSubfixBeforeExt(archiveFilePath string) (string, error) {
	dirPath := filepath.Dir(archiveFilePath)
	nameArr := strings.FieldsFunc(filepath.Base(archiveFilePath), func(char rune) bool {
		return char == '.'
	})

	for number := 1; ; number++ {
		basePath := nameArr[0] + "-" + fmt.Sprint(number)
		for _, ext := range nameArr[1:] {
			basePath += "." + ext
		}

		stat, err := os.Stat(filepath.Join(dirPath, basePath))
		if os.IsNotExist(err) {
			return filepath.Join(dirPath, basePath), nil
		}
		if err != nil {
			return "", fmt.Errorf("read file info: %v", err)
		}
		if !stat.IsDir() {
			continue
		}
	}
}

func deleteSliceElements[T comparable](arr []T, index ...int) ([]T, error) {
	for _, v := range index {
		if v < 0 || v >= len(arr) {
			return nil, fmt.Errorf("index out of range: %v", v)
		}
	}

	var newArr []T
	for n, e := range arr {
		canAppend := true
		for _, i := range index {
			if n == i {
				canAppend = false
				break
			}
		}
		if canAppend {
			newArr = append(newArr, e)
		}
	}

	return newArr, nil
}

func convertIntArray(array []string) ([]int, error) {
	var intList []int
	for _, v := range array {
		number, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return nil, err
		}
		intList = append(intList, int(number))
	}
	return intList, nil
}

func fmtPath(input string) (output string) {
	output = strings.ReplaceAll(input, "\\", "/")
	output = path.Clean(output)
	return output
}

func isPath(s string) bool {
	return strings.HasPrefix(s, "./") || strings.HasPrefix(s, "/")
}
