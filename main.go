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
	"strings"
	"time"

	"acovia.net/mc-saver/parse"
	"acovia.net/record"
)

// 默认配置内容，gencfg 子命令会将其写入指定的规则文件
var (
	UseLegacyMode  bool   = false
	configFilePath string = "save-rule.json"
	levelDirPath   string = "world"
	outputPath     string = "."

	cmdMap map[string]func() = map[string]func(){
		"run":    run,
		"gencfg": gencfg,
		"": empty,
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

// main 是程序入口：解析命令行参数、校验存档目录，并按规则将选中文件写入 zip 压缩包
func main() {

	// 定义命令行参数，-c 指定规则文件
	flag.StringVar(&configFilePath, "c", configFilePath, "config file path")
	flag.BoolFunc("l", "legacy world mode", func(s string) error {
		UseLegacyMode = true
		return nil
	})
	flag.Parse()

	// 考虑 windows 路径分割符是 \，提前转换为 / 路径。
	levelDirPath = strings.ReplaceAll(levelDirPath, "\\", "/")
	configFilePath = strings.ReplaceAll(configFilePath, "\\", "/")
	outputPath = strings.ReplaceAll(outputPath, "\\", "/")

	// 校验存档目录是否有效，无效则直接退出

	function, ok := cmdMap[flag.Arg(0)]
	if !ok {
		record.Error("%v", errors.New("\""+flag.Arg(0)+"\" command not found"))
		os.Exit(1)
	}

	function()

	record.Info("Backup completed successfully!")
}

func empty() {
	scanner := bufio.NewScanner(os.Stdin)
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		fmt.Printf("generate a default config file (y/n): ")
		if scanner.Scan() {
			if input := scanner.Text(); input == "y" || input == "Y" {
				if _, err := os.Stat(configFilePath); !os.IsNotExist(err) {
					record.Error("(generate default config) %v",errors.New("file \"" + configFilePath + "\" existed"))
				}
				err := os.WriteFile(configFilePath, []byte(defaultConfig), 0644)
				if err != nil {
					record.Error("%v", err)
					os.Exit(1)
				}
				record.Info("created default config file: %s", configFilePath)
			} else {
				os.Exit(0)
			}
		}
	}
	fmt.Printf("please enter level directory path (world): ")
	if scanner.Scan() {
		input := scanner.Text()
		if len(input) != 0 {
			levelDirPath = input
		}
	}
	fmt.Printf("please enter output path (world-xxxx-xx-xx.zip): ")
	if scanner.Scan() {
		input := scanner.Text()
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
		record.Error("(generate default config) %v",errors.New("file \"" + configFilePath + "\" existed"))
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
		levelDirPath = flag.Arg(1)
	}

	if len(flag.Arg(2)) != 0 {
		outputPath = flag.Arg(2)
	}

	err := levelFileIsValid(levelDirPath)
	if err != nil {
		record.Error("%v", err)
		os.Exit(1)
	}

	// 创建 zip 压缩包写入器
	zipWriter, fileWriter, err := createZipWriter()
	if err != nil {
		record.Error("%v", err)
		os.Exit(1)
	}

	if UseLegacyMode {
		if err := parse.SaveOldAllFile(levelDirPath, configFilePath, zipWriter, addFile); err != nil {
			record.Error("%v", err)
			os.Remove(fileWriter.Name())
			os.Exit(1)
		}
	} else {
		if err := parse.SaveAllFile(levelDirPath, configFilePath, zipWriter, addFile); err != nil {
			record.Error("%v", err)
			os.Remove(fileWriter.Name())
			os.Exit(1)
		}
	}

	// 关闭 zip 写入器，完成压缩包内容的写入
	if err := zipWriter.Close(); err != nil {
		record.Error("%v", err)
		os.Remove(fileWriter.Name())
		os.Exit(1)
	}

	if err := fileWriter.Close(); err != nil {
		record.Error("%v", err)
		os.Remove(fileWriter.Name())
		os.Exit(1)
	}

}

func levelFileIsValid(levelDir string) error {

	// 获取存档目录的信息，判断目录是否存在
	archiveFileInfo, err := os.Stat(levelDir)
	if err != nil {
		return fmt.Errorf("(read level directory) %w", err)
	}

	// 若路径存在但不是目录，则视为无效存档
	if !archiveFileInfo.IsDir() {
		return errors.New("(check world directory) level path is not a directory")
	}

	return nil
}

// 创建 zip 压缩包写入器，输出文件名为存档名加日期
func createZipWriter() (*zip.Writer, *os.File, error) {
	var archiveFilePath string
	archiveFileName := path.Base(levelDirPath) + "-" + time.Now().Format(time.DateOnly) + ".zip"
	outputFileInfo, err := os.Stat(outputPath)
	switch true {

	case os.IsNotExist(err):
		err = os.MkdirAll(path.Dir(outputPath), 0755)
		if err != nil {
			return nil, nil, fmt.Errorf("(create archive) %w", err)
		}
		archiveFilePath = path.Join(path.Dir(outputPath), path.Base(outputPath))
		
	case !os.IsNotExist(err) && err != nil:
		return nil, nil, fmt.Errorf("(create archive) %w", err)

	case outputFileInfo.IsDir():
		archiveFilePath = path.Join(outputPath, archiveFileName)

	case !outputFileInfo.IsDir():
		archiveFilePath = outputPath

	default:
		return nil, nil, fmt.Errorf("(create archive) %w", errors.New("invalid output path"))

	}

	fileWriter, err := os.Create(path.Join(archiveFilePath))
	if err != nil {
		return nil, nil, fmt.Errorf("(create archive) %w", err)
	}

	// 基于输出文件创建 zip 写入器
	w := zip.NewWriter(fileWriter)

	return w, fileWriter, nil

}

// addFile 将单个文件写入 zip 压缩包。
// 源文件不存在时仅记录警告并跳过，而不是终止整个备份流程（有意为之）。
func addFile(fileName string, filePath string, zipWriter *zip.Writer) error {

	// 打开源文件；若文件不存在则输出警告并跳过（有意为之：用户目录缺少文件不应中断备份）
	fileReader, err := os.Open(filePath)
	if err != nil {
		record.Warn("(skip file) %v", err)
		return nil
	}

	defer fileReader.Close()

	// 在压缩包中创建同名条目，保持与存档一致的目录结构
	file, err := zipWriter.Create(fileName)
	if err != nil {
		return fmt.Errorf("(create new file in archive) %w", err)
	}

	// 将区域文件内容写入压缩包
	_, err = io.Copy(file, fileReader)
	if err != nil {
		return fmt.Errorf("(write compressed file) %w", err)
	}

	record.Info("(added) %s", fileName)

	return nil

}
