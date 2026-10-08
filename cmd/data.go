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
)

var (
	args           []string
	cmd            string
	config         save.Config = save.NullConfig
	configFilePath string
	subCmdArgs     []string

	useLegacyMode  bool   = false
	worldDirPath   string = "world"
	configFileName string = "saver.json"
	outputPath     string = "."

	// Declare the commandMapping Func : 
	cmdMap map[string] func() 

	root *os.Root

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


// cmdMapping func init : 
func init() {
	cmdMap = map[string]func(){
		"run":         run,
		"gencfg":      gencfg,
		"help":        help,
		"list":        list,
		"list-config": listConfig,
		"list-dms":    listDms,
		"add-dms":     addDms,
		"del-dms":     delDms,
		"mod-dms":     modDms,
		"list-range":  listRange,
		"add-range":   addRange,
		"del-range":   delRange,
		"mod-range":   modRange,
		"list-simple": listSimple,
		"add-simple":  addSimple,
		"del-simple":  delSimple,
		"mod-simple":  modSimple,
		"list-file":   listFile,
		"add-file":    addFile,
		"del-file":    delFile,
		"mod-file":    modFile,
		"repl": repl,
		"about": about,
	}
}


var helpInfo = `mc-saver [-l] [-color] <command> <world> [args...]

every command takes the world directory as its first argument, and the rule
file is read from <world>/saver.json. run "gencfg <world>" first to create it.

backup command:

	run <world> [output]
		back up <world> according to <world>/saver.json.
		output defaults to ".", where a dated zip is created.

	gencfg <world>
		write the default rule file to <world>/saver.json.
		throw error if the file is already existed.

	repl <world>
		enter interactive repl mode for <world>.

	help
		print this help text.
		
	about
		The information and copyright about this kit.

config command:

	indices are 0-based, as shown by the list commands. an index outside the
	list is an error for both the mod and the delete commands.

	list <world>
		list all dimension rules and file rules.

	list-config <world> <dimension>...
		list both range rules and simple rules of a dimension.

	list-dms <world>
		list dimension namespace ids.

	add-dms <world> <dimension>...
		add a dimension with the default range rule.

	del-dms <world> <dimension>...
		delete a dimension.

	mod-dms <world> <old_dimension> <new_dimension>
		rename a dimension, keeping its rules.

	list-range <world> <dimension>...
		list range rules of a dimension.

	add-range <world> <dimension> <from_x> <from_y> <to_x> <to_y>
		add a range rule to a dimension.

	del-range <world> <dimension> <index>...
		delete the range rules of the given indices.

	mod-range <world> <dimension> <index> <from_x> <from_y> <to_x> <to_y>
		replace the range rule at the given index.

	list-simple <world> <dimension>...
		list simple rules of a dimension.

	add-simple <world> <dimension> <x> <y>
		add a simple rule to a dimension.

	del-simple <world> <dimension> <index>...
		delete the simple rules of the given indices.

	mod-simple <world> <dimension> <index> <x> <y>
		replace the simple rule at the given index.

	list-file <world>
		list file rules.

	add-file <world> <name>...
		add one or more file rules.

	del-file <world> <index>...
		delete the file rules of the given indices.

	mod-file <world> <index> <name>
		replace the file rule at the given index.

options:

	-l
		legacy world mode, for worlds from before 1.21.11.

	-color
		enable color output.
`



func formatOutputPath(archiveFilePath string) (string, error) {

	worldAbsPath, err := filepath.Abs(archiveFilePath)
	if err != nil {
		return "", fmt.Errorf("load absolute path: %w", err)
	}

	worldDirName := path.Base(worldAbsPath)

	outputFileInfo, err := os.Stat(outputPath)
	switch true {

	case os.IsNotExist(err):
		err = os.MkdirAll(path.Dir(outputPath), 0755)
		if err != nil {
			return "", fmt.Errorf("create output directory: %v", err)
		}
		archiveFilePath = outputPath

	case !os.IsNotExist(err) && err != nil:
		return "", fmt.Errorf("read file info: %v", err)

	case outputFileInfo.IsDir():
		archiveFileName := worldDirName + "-" + time.Now().Format(time.DateOnly) + ".zip"
		archiveFilePath = path.Join(outputPath, archiveFileName)
		archiveFilePath, err = addSubfixBeforeExt(archiveFilePath)
		if err != nil {
			return "", fmt.Errorf("add subfix: %v", err)
		}

	default:
		archiveFilePath, err = addSubfixBeforeExt(outputPath)
		if err != nil {
			return "", fmt.Errorf("add subfix: %v", err)
		}
	}

	return archiveFilePath, nil
}

func addSubfixBeforeExt(archiveFilePath string) (string, error) {
	basePath := path.Base(archiveFilePath)
	dirPath := path.Dir(archiveFilePath)
	nameArr := strings.FieldsFunc(basePath, isExtKeyWord)
	for number := 1; ; number++ {
		var subfix string = "-" + fmt.Sprint(number)
		basePath = nameArr[0] + subfix
		for _, ext := range nameArr[1:] {
			basePath += "." + ext
		}
		stat, err := os.Stat(path.Join(dirPath, basePath))
		if os.IsNotExist(err) {
			break
		} else if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("read file info: %v", err)
		} else if !stat.IsDir() || !os.IsNotExist(err) {
			continue
		}
	}
	return path.Join(dirPath, basePath), nil
}

func isExtKeyWord(char rune) bool {
	if char == rune("."[0]) {
		return true
	} else {
		return false
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

func loadSubCmdArgs() []string {
	if len(args) > 1 {
		return args[1:]
	}
	return nil
}
