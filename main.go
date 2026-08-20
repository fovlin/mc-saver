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

// 默认配置内容，gencfg 子命令会将其写入指定的规则文件
var (
	UseLegacyMode  bool   = false
	configFilePath string = "save-rule.json"
	levelDirPath   string = "world"
	outputPath     string = "."

	cmdMap map[string]func() = map[string]func(){
		"run":    run,
		"gencfg": gencfg,
		"":       empty,
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

	function, ok := cmdMap[flag.Arg(0)]
	if !ok {
		record.Error("%v", errors.New("\""+flag.Arg(0)+"\" command not found"))
		os.Exit(1)
	}

	function()

	record.Info("Backup completed successfully!")
}

// empty 是交互向导：无子命令时依次询问是否生成默认配置、存档目录和输出路径，然后开始备份
func empty() {
	scanner := bufio.NewScanner(os.Stdin)
	// 配置文件缺失时先询问是否生成默认配置
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		fmt.Printf("generate a default config file (y/n): ")
		if scanner.Scan() {
			if input := scanner.Text(); input == "y" || input == "Y" {
				err := os.WriteFile(configFilePath, []byte(defaultConfig), 0644)
				if err != nil {
					record.Error("%v", err)
					os.Exit(1)
				}
				record.Info("(created default config file): %s", configFilePath)
			} else {
				os.Exit(0)
			}
		}
	} else if !os.IsNotExist(err) && err != nil {
		record.Error("(verify config file) %v", err)
		os.Exit(1)
	}
	// 依次询问存档目录与输出路径，直接回车使用默认值
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

// gencfg 将默认配置写入规则文件后退出，方便用户快速生成配置
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

// run 打开存档根目录、创建 zip 压缩包，并按规则将选中文件写入压缩包
func run() {

	// 子命令的位置参数：第一个为存档目录，第二个为输出路径
	if len(flag.Arg(1)) != 0 {
		levelDirPath = flag.Arg(1)
	}

	if len(flag.Arg(2)) != 0 {
		outputPath = flag.Arg(2)
	}

	// 以存档根目录打开 root，之后所有文件访问都会被限制在存档内（防止 .. 逃逸）
	root, err := os.OpenRoot(levelDirPath)
	if err != nil {
		record.Error("(open level as root directory) %v", err)
		os.Exit(1)
	}
	defer root.Close()

	// 创建 zip 压缩包写入器
	zipWriter, fileWriter, err := createZipWriter()
	if err != nil {
		record.Error("(create zip writer) %v", err)
		os.Exit(1)
	}

	if UseLegacyMode {
		if err := parse.SaveOldAllFile(root, configFilePath, zipWriter, addFile); err != nil {
			record.Error("(write file into archive) %v", err)
			os.Remove(fileWriter.Name())
			os.Exit(1)
		}
	} else {
		if err := parse.SaveAllFile(root, configFilePath, zipWriter, addFile); err != nil {
			record.Error("(write file into archive) %v", err)
			os.Remove(fileWriter.Name())
			os.Exit(1)
		}
	}

	// 关闭 zip 写入器，完成压缩包内容的写入
	if err := zipWriter.Close(); err != nil {
		record.Error("(close zip writer) %v", err)
		os.Remove(fileWriter.Name())
		os.Exit(1)
	}

	if err := fileWriter.Close(); err != nil {
		record.Error("(close file writer) %v", err)
		os.Remove(fileWriter.Name())
		os.Exit(1)
	}

}

func createZipWriter() (*zip.Writer, *os.File, error) {
	var archiveFilePath string
	archiveFileName := path.Base(levelDirPath) + "-" + time.Now().Format(time.DateOnly) + ".zip"
	outputFileInfo, err := os.Stat(outputPath)
	// 输出路径存在且为目录时按 存档名-日期.zip 命名，否则将输出路径视为完整文件路径
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

	// 输出文件已存在时自动追加 -1、-2 序号，避免覆盖同一天的旧备份
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

	// 基于输出文件创建 zip 写入器
	w := zip.NewWriter(fileWriter)

	return w, fileWriter, nil

}

// addFile 将单个文件写入 zip 压缩包。
// 源文件不存在时仅记录警告并跳过，而不是终止整个备份流程（有意为之）。
func addFile(root *os.Root, fileName string, zipWriter *zip.Writer) error {

	fileReader, err := root.Open(fileName)
	if err != nil {
		record.Warn("(skip file) %v", err)
		return nil
	}
	defer fileReader.Close()

	fileInfo, err := root.Stat(fileName)
	if err != nil {
		return err
	}

	zipFileHeader, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	headerName := path.Join(path.Base(levelDirPath), fileName)

	// 压缩包内以存档名作为顶层目录，保证备份可直接还原为存档
	zipFileHeader.Name = headerName
	// 统一使用 Deflate 压缩，兼顾体积与兼容性
	zipFileHeader.Method = zip.Deflate

	file, err := zipWriter.CreateHeader(zipFileHeader)

	if err != nil {
		return fmt.Errorf("(create new file in archive) %w", err)
	}

	// 将区域文件内容写入压缩包
	_, err = io.Copy(file, fileReader)
	if err != nil {
		return fmt.Errorf("(write compressed file) %w", err)
	}

	record.Info("(added) %s", headerName)

	return nil

}
