package main

import (
	"acovia.net/mc-saver/parse"
	"acovia.net/record"
	"archive/zip"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"
)

// 默认配置内容，gencfg 子命令会将其写入指定的规则文件
var (
	configFilePath string
	levelDirPath   string
	outputDirPath  string

	cmdMap map[string]func() = map[string]func(){
		"":       run,
		"old":    old,
		"gencfg": gencfg,
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

	// 定义命令行参数，-c 指定规则文件，-t 指定存档目录
	flag.StringVar(&configFilePath, "c", "save-rule.json", "Json config file path.")
	flag.StringVar(&levelDirPath, "t", "world", "World file path.")
	flag.StringVar(&outputDirPath, "o", ".", "Output directory path.")
	flag.Parse()

	// 考虑 windows 路径分割符是 \，提前转换为 / 路径。
	levelDirPath = strings.ReplaceAll(levelDirPath, "\\", "/")
	configFilePath = strings.ReplaceAll(configFilePath, "\\", "/")
	outputDirPath = strings.ReplaceAll(outputDirPath, "\\", "/")

	// 校验存档目录是否有效，无效则直接退出

	function, ok := cmdMap[flag.Arg(0)]
	if !ok {
		record.Error("%v", errors.New("\""+flag.Arg(0)+"\"command not found"))
		os.Exit(1)
	}

	function()

	record.Info("Backup completed successfully!")
}

func old() {

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

	if err := parse.SaveOldAllFile(levelDirPath, configFilePath, zipWriter, addFile); err != nil {
		record.Error("%v", err)
		os.Exit(1)
	}

	// 关闭 zip 写入器，完成压缩包内容的写入
	if err := zipWriter.Close(); err != nil {
		record.Error("%v", err)
		os.Exit(1)
	}

	if err := fileWriter.Close(); err != nil {
		record.Error("%v", err)
		os.Exit(1)
	}
}

func gencfg() {
	// 将默认配置写入规则文件后直接退出，方便用户快速生成配置
	err := os.WriteFile(configFilePath, []byte(defaultConfig), 0644)
	if err != nil {
		record.Error("%v", err)
		os.Exit(1)
	}
	record.Info("created default config file: %s", configFilePath)
	os.Exit(0)
}

func run() {

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

	if err := parse.SaveAllFile(levelDirPath, configFilePath, zipWriter, addFile); err != nil {
		record.Error("%v", err)
		os.Exit(1)
	}

	// 关闭 zip 写入器，完成压缩包内容的写入
	if err := zipWriter.Close(); err != nil {
		record.Error("%v", err)
		os.Exit(1)
	}

	if err := fileWriter.Close(); err != nil {
		record.Error("%v", err)
		os.Exit(1)
	}

}

// 校验存档目录是否有效，需要是存在的目录且包含 level.dat 文件
func levelFileIsValid(levelDir string) error {

	// 获取存档目录的信息，判断目录是否存在
	archiveFileInfo, err := os.Stat(levelDir)
	if err != nil {
		return fmt.Errorf("(get level directory info) %w", err)
	}

	// 若路径存在但不是目录，则视为无效存档
	if !archiveFileInfo.IsDir() {
		return errors.New("(check world directory) level path is not a directory")
	}

	return nil
}

// 创建 zip 压缩包写入器，输出文件名为存档名加日期
func createZipWriter() (*zip.Writer, *os.File, error) {

	archiveFileName := path.Base(levelDirPath) + "-" + time.Now().Format(time.DateOnly) + ".zip"

	// 输出目录不存在时自动创建（MkdirAll 支持多级路径）
	_, err := os.Stat(outputDirPath)

	if err != nil {
		err := os.MkdirAll(outputDirPath, fs.ModePerm)
		if err != nil {
			return nil, nil, fmt.Errorf("(create directory) "+outputDirPath+": %w", err)
		}
	}

	// 创建输出文件，文件名格式为 存档名-日期.zip
	fileWriter, err := os.Create(path.Join(outputDirPath, archiveFileName))
	if err != nil {
		return nil, nil, fmt.Errorf("(create archive file) %w", err)
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
