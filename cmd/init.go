package main

import (
	"archive/zip"
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
	config     save.Config = save.NewNullConfig()
	subCmdArgs []string

	useLegacyMode     bool
	replMode          bool
	configFileName    string = "saver.json"
	defaultOutputPath string = "."
	prompt            string = "mc-saver > "

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

type Writer struct {
	ZipWriter   *zip.Writer
	TempFile    *os.File
	ArchivePath string
	closed      bool
}

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

func addSubfixBeforeExt(outputPath string) (string, error) {
	resultPath := outputPath

	for number := 1; number > 0; number++ {
		_, err := os.Stat(resultPath)

		switch true {
		case os.IsNotExist(err):
			return resultPath, nil

		case err != nil:
			return "", err
		}

		resultPath = outputPath

		dir := path.Dir(resultPath)
		basePath := path.Base(resultPath)
		ext := path.Ext(basePath)
		pureName, _ := strings.CutSuffix(basePath, ext)
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
		if len(v) == 0 {
			continue
		}

		switch true {
		case v[0] == ':':
			if len(v) > 1 {
				v = v[1:]
			}

		case v[0] == '=':
			if len(v) > 1 {
				v = v[1:]
			}

		case v[0] == '@':
			if len(v) > 1 {
				v = v[1:]
			}
		}
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
	s = strings.ReplaceAll(s, "\\", "/")
	return strings.HasPrefix(s, "./") || strings.HasPrefix(s, "/")
}
func loadWorld(usage string, replUsage string) (worldDirPath string, configFilePath string, err error) {
	command, ok := cmdInfo[cmd]
	if !ok {
		return "", "", record.FmtError("unknown command:", cmd)
	}

	if len(subCmdArgs) < command.MinArgsNumber {
		return "", "", usageError(usage, replUsage)
	}

	worldDirPath, err = filepath.Abs(subCmdArgs[0])
	if err != nil {
		return "", "", record.FmtError("load world absolute path:", err)
	}
	worldDirPath = fmtPath(worldDirPath)
	configFilePath = fmtPath(path.Join(worldDirPath, configFileName))

	if err := initWorldConfig(worldDirPath, configFilePath); err != nil {
		return "", "", record.FmtError("init world config:", err)
	}

	return worldDirPath, configFilePath, nil
}

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
