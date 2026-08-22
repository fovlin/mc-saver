package main

import (
	"archive/zip"
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"acovia.net/mc-saver/parse"
	"acovia.net/record"
)

var (
	UseLegacyMode  bool   = false
	configFilePath string = "save-rule.json"
	worldDirPath   string = "world"
	outputPath     string = "."

	cmdMap map[string]func() = map[string]func(){
		"run":    run,
		"gencfg": gencfg,
		"help":   help,
		"repl":   repl,
		"":       repl,
	}

	defaultConfig string = `{
	"dimension":{
		"minecraft:overworld":{
			"range":[
				{ "from":[-1, -1], "to":[1, 1] }
			]
		},
		"minecraft:the_nether":{
			"range":[
				{ "from":[-1, -1], "to":[1, 1] }
			]
		},
		"minecraft:the_end":{
			"range":[
				{ "from":[-1, -1], "to":[1, 1] }
			]
		}
	},
	"file":[
		"level.dat",
		"data",
		"datapacks",
		"players"
	]
}`
)

func main() {

	flag.StringVar(&configFilePath, "c", configFilePath, "config file path")
	flag.BoolFunc("l", "legacy world mode", func(s string) error {
		UseLegacyMode = true
		return nil
	})
	flag.Parse()

	configFilePath = strings.ReplaceAll(configFilePath, "\\", "/")

	function, ok := cmdMap[flag.Arg(0)]
	if !ok {
		record.Error("%v", errors.New("\""+flag.Arg(0)+"\" command not found"))
		os.Exit(1)
	}

	function()
}

func help() {
	helpOutput :=
		`
	command:

	run [option] [world_path] [output_path]
		start the backup according to the config file
		the first path is world path, default is "world"
		second path is output path, default is "world-$time.zip"

	gencfg [output_file]
		generate a default config file

	options:

	-c <path>
		specify the config file

`
	fmt.Printf("%v", helpOutput)
}

func repl() {
	scanner := bufio.NewScanner(os.Stdin)
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		fmt.Printf("generate a default config file? (y/n): ")
		if scanner.Scan() {
			if input := scanner.Text(); input != "y" && input != "Y" && len(input) != 0 {
				os.Exit(0)
			}
			err := os.WriteFile(configFilePath, []byte(defaultConfig), 0644)
			if err != nil {
				record.Error("%v", err)
				os.Exit(1)
			}
			record.Info("(created default config file): %s", configFilePath)
			fmt.Printf("continue with default config file? (y/n): ")
			if scanner.Scan() {
				input := scanner.Text()
				if input != "y" && input != "Y" && len(input) != 0 {
					os.Exit(0)
				}
			}
		}
	} else if !os.IsNotExist(err) && err != nil {
		record.Error("(verify config file) %v", err)
		os.Exit(1)
	}
	fmt.Printf("please enter world directory path (default is world): ")
	if scanner.Scan() {
		input := scanner.Text()
		input = strings.ReplaceAll(input, "\\", "/")
		if len(input) != 0 {
			worldDirPath = input
		}
	}
	fmt.Printf("please enter output path (default is world-$time.zip): ")
	if scanner.Scan() {
		input := scanner.Text()
		input = strings.ReplaceAll(input, "\\", "/")
		if len(input) != 0 {
			outputPath = input
		}
	}
	run()
}

func gencfg() {
	if flag.Arg(1) != "" {
		configFilePath = flag.Arg(1)
	}
	if _, err := os.Stat(configFilePath); !os.IsNotExist(err) {
		record.Error("(generate default config) %v", errors.New("file \""+configFilePath+"\" existed"))
		os.Exit(1)
	} else if !os.IsNotExist(err) && err != nil {
		record.Error("(verify config file) %v", err)
		os.Exit(1)
	}

	err := os.WriteFile(configFilePath, []byte(defaultConfig), 0644)
	if err != nil {
		record.Error("%v", err)
		os.Exit(1)
	}
	record.Info("created default config file: %s", configFilePath)
	os.Exit(0)
}

func run() {

	if len(flag.Arg(1)) != 0 {
		worldDirPath = flag.Arg(1)
	}

	if len(flag.Arg(2)) != 0 {
		outputPath = flag.Arg(2)
	}

	worldDirPath = strings.ReplaceAll(worldDirPath, "\\", "/")
	outputPath = strings.ReplaceAll(outputPath, "\\", "/")

	root, err := os.OpenRoot(worldDirPath)
	if err != nil {
		record.Error("(open level as root directory) %v", err)
		os.Exit(1)
	}
	defer root.Close()

	zipWriter, fileWriter, err := createZipWriter()
	if err != nil {
		record.Error("(create zip writer) %v", err)
		os.Exit(1)
	}
	
	defer fileWriter.Close()
	defer zipWriter.Close()

	if UseLegacyMode {
		if err := parse.SaveOldAllFile(root, configFilePath, zipWriter, addFile); err != nil {
			record.Error("(write file into archive) %v", err)
			zipWriter.Close()
			fileWriter.Close()
			os.Remove(fileWriter.Name())
			os.Exit(1)
		}
	} else {
		if err := parse.SaveAllFile(root, configFilePath, zipWriter, addFile); err != nil {
			record.Error("(write file into archive) %v", err)
			zipWriter.Close()
			fileWriter.Close()
			os.Remove(fileWriter.Name())
			os.Exit(1)
		}
	}

	record.Info("Backup completed successfully!")

}

func createZipWriter() (*zip.Writer, *os.File, error) {
	var archiveFilePath string
	archiveFileName := path.Base(worldDirPath) + "-" + time.Now().Format(time.DateOnly) + ".zip"
	outputFileInfo, err := os.Stat(outputPath)
	switch true {

	case os.IsNotExist(err):
		err = os.MkdirAll(path.Dir(outputPath), 0755)
		if err != nil {
			return nil, nil, fmt.Errorf("(create output directory) %w", err)
		}
		archiveFilePath = outputPath

	case !os.IsNotExist(err) && err != nil:
		return nil, nil, fmt.Errorf("(verify path) %w", err)

	case outputFileInfo.IsDir():
		archiveFilePath = path.Join(outputPath, archiveFileName)

	default:
		archiveFilePath = outputPath

	}

	if _, err := os.Stat(archiveFilePath); !os.IsNotExist(err) && err == nil {
		archiveFileExt := path.Ext(archiveFilePath)
		archiveFileNoExt, _ := strings.CutSuffix(archiveFilePath, archiveFileExt)
		for n := 1; ; n++ {
			suffix := "-" + strconv.Itoa(n)
			if _, err := os.Stat(archiveFileNoExt + suffix + archiveFileExt); os.IsNotExist(err) {
				archiveFilePath = archiveFileNoExt + suffix + archiveFileExt
				break
			}
		}
	} else if !os.IsNotExist(err) && err != nil {
		return nil, nil, err
	}

	fileWriter, err := os.Create(archiveFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("(create archive) %w", err)
	}

	w := zip.NewWriter(fileWriter)

	return w, fileWriter, nil

}

func addFile(root *os.Root, filePath string, zipWriter *zip.Writer) error {

	fileReader, err := root.Open(filePath)
	if err != nil {
		record.Warn("(skip file) %v", err)
		return nil
	}
	defer fileReader.Close()

	fileInfo, err := root.Stat(filePath)
	if err != nil {
		return err
	}

	zipFileHeader, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	headerName := path.Join(path.Base(worldDirPath), filePath)

	zipFileHeader.Name = headerName
	zipFileHeader.Method = zip.Deflate

	file, err := zipWriter.CreateHeader(zipFileHeader)

	if err != nil {
		return fmt.Errorf("(create new file in archive) %w", err)
	}

	_, err = io.Copy(file, fileReader)
	if err != nil {
		return fmt.Errorf("(write compressed file) %w", err)
	}

	record.Info("(added) %s", headerName)

	return nil

}
