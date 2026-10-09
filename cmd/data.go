package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"acovia.net/minecraft/save"
	"acovia.net/record"
)

var (
	cmd        string
	config     save.Config = save.NullConfig
	subCmdArgs []string

	useLegacyMode     bool
	replMode          bool
	configFileName    string = "saver.json"
	defaultOutputPath string = "."

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

func initOutputDir(inputPath string) (string, error) {
	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		return "", fmt.Errorf("load absolute path: %w", err)
	}

	outputPathStat, err := os.Stat(absPath)
	
	if outputPathStat.IsDir() {
		return absPath, nil
	}

	return path.Dir(absPath), nil
}

func addSubfixBeforeExt(outputPath string) (string, error) {
	resultPath := outputPath

	for number := 0; number >= 0; number++ {
		_, err := os.Stat(resultPath)
		if os.IsNotExist(err) {
			return filepath.Join(resultPath), nil
		}

		resultPath = outputPath

		dir := path.Dir(resultPath)
		name := path.Base(resultPath)
		ext := path.Ext(name)
		pureName, _ := strings.CutSuffix(name, ext)
		pureName += fmt.Sprint("-", number)

		resultPath = path.Join(dir, fmt.Sprint(pureName, ext))
	}

	return resultPath, nil
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
