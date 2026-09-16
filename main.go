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
	"path/filepath"
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
		record.Error(errors.New("\"" + flag.Arg(0) + "\" command not found"))
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
				record.Error(err)
			}
			record.Info("created default config file:", configFilePath)
			fmt.Printf("continue with default config file? (y/n): ")
			if scanner.Scan() {
				input := scanner.Text()
				if input != "y" && input != "Y" && len(input) != 0 {
					os.Exit(0)
				}
			}
		}
	} else if !os.IsNotExist(err) && err != nil {
		record.Error("verify config file:", err)
	}
	fmt.Printf("world directory path (default is world): ")
	if scanner.Scan() {
		input := scanner.Text()
		input = strings.ReplaceAll(input, "\\", "/")
		if len(input) != 0 {
			worldDirPath = input
		}
		if absPath, err := filepath.Abs(worldDirPath); err != nil {
			record.Error("check world directory path):", err)
		} else {
			worldDirPath = absPath
		}
	}
	fmt.Printf("output path (default is world-$time.zip): ")
	if scanner.Scan() {
		input := scanner.Text()
		input = strings.ReplaceAll(input, "\\", "/")
		if len(input) != 0 {
			outputPath = path.Clean(input)
		}
	}
	run()
}

func gencfg() {
	if flag.Arg(1) != "" {
		configFilePath = flag.Arg(1)
	}

	if _, err := os.Stat(configFilePath); !os.IsNotExist(err) {
		record.Error("generate config file:", "\"%v\"", "already exists", configFilePath)
	}

	err := os.WriteFile(configFilePath, []byte(defaultConfig), 0644)
	if err != nil {
		record.Error(err)
	}
	record.Info("created config file:", configFilePath)
	os.Exit(0)
}

func run() {

	if len(flag.Arg(1)) != 0 {
		if absPath, err := filepath.Abs(flag.Arg(1)); err != nil {
			record.Error("check world directory path:", err)
		} else {
			worldDirPath = absPath
		}
	}

	if len(flag.Arg(2)) != 0 {
		outputPath = path.Clean(flag.Arg(2))
	}

	worldDirPath = strings.ReplaceAll(worldDirPath, "\\", "/")
	outputPath = strings.ReplaceAll(outputPath, "\\", "/")

	root, err := os.OpenRoot(worldDirPath)
	if err != nil {
		record.Error("open world directory:", err)
	}

	zipWriter, _, end, err := createZipWriter()
	if err != nil {
		record.Error("create zip writer:", err)
	}

	defer end(err)

	if UseLegacyMode {
		if err := parse.SaveOldAllFile(root, configFilePath, zipWriter, addFile); err != nil {
			if err := end(err); err != nil {
				record.Error(err)
			}
		}
	} else {
		if err := parse.SaveAllFile(root, configFilePath, zipWriter, addFile); err != nil {
			if err := end(err); err != nil {
				record.Error(err)
			}
		}
	}
}

func createZipWriter() (*zip.Writer, *os.File, func(error) error, error) {

	var archiveFilePath string
	outputFileInfo, err := os.Stat(outputPath)
	switch true {

	case os.IsNotExist(err):
		err = os.MkdirAll(path.Dir(outputPath), 0755)
		if err != nil {
			return nil, nil, nil, err
		}
		archiveFilePath = outputPath

	case !os.IsNotExist(err) && err != nil:
		return nil, nil, nil, err

	case outputFileInfo.IsDir():
		archiveFileName := path.Base(worldDirPath) + "-" + time.Now().Format(time.DateOnly) + ".zip"
		archiveFilePath = path.Join(outputPath, archiveFileName)

	default:
		archiveFilePath = outputPath
	}

	file, err := os.CreateTemp(path.Dir(archiveFilePath), "archive")
	if err != nil {
		return nil, nil, nil, err
	}

	if err = file.Chmod(0655); err != nil {
		return nil, nil, nil, err
	}

	zipWriter := zip.NewWriter(file)

	end := func(err error) error {

		if err != nil {
			os.Remove(file.Name())
			return err
		}

		if err := os.Rename(file.Name(), archiveFilePath); err != nil {
			os.Remove(file.Name())
			return err
		}

		if err := zipWriter.Close(); err != nil {
			os.Remove(file.Name())
			return err
		}

		if err := file.Close(); err != nil {
			os.Remove(file.Name())
			return err
		}

		record.Info("backup completed successfully!")

		return nil

	}

	return zipWriter, file, end, nil

}

func addFile(root *os.Root, filePath string, zipWriter *zip.Writer) error {

	fileReader, err := root.Open(filePath)
	if err != nil {
		record.Warn("skip file:", err)
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

	headerName := path.Join(path.Base(root.Name()), filePath)

	zipFileHeader.Name = headerName
	zipFileHeader.Method = zip.Deflate

	file, err := zipWriter.CreateHeader(zipFileHeader)

	if err != nil {
		return fmt.Errorf("(create file) %w", err)
	}

	if err = record.RunningInfo(func() error {
		_, err = io.Copy(file, fileReader)
		if err != nil {
			return fmt.Errorf("(write file) %w", err)
		}
		return nil
	}, "adding", headerName); err != nil {
		return err
	}

	return nil

}