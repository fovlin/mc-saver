package main

import (
	"acovia.net/mc-saver/parse"
	"acovia.net/record"
	"archive/zip"
	"errors"
	"flag"
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
	flag.StringVar(&outputDirPath, "o", ".", "Out put directory path.")
	flag.Parse()

	// gencfg 子命令：生成默认配置文件后直接退出
	if flag.Arg(0) == "gencfg" {
		// 将默认配置写入规则文件后直接退出，方便用户快速生成配置
		err := os.WriteFile(configFilePath, []byte(defaultConfig), 0666)
		if err != nil {
			record.Error("%v", err)
			os.Exit(1)
		}
		record.Info("Create default config - \"%v\"", configFilePath)
		return
	}

	// 考虑 windows 用户路径分割符是 \，提前转换为 / 路径。
	levelDirPath = strings.ReplaceAll(levelDirPath, "\\", "/")
	configFilePath = strings.ReplaceAll(configFilePath, "\\", "/")
	outputDirPath = strings.ReplaceAll(outputDirPath, "\\", "/")

	// 校验存档目录是否有效，无效则直接退出
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

	err = parse.SaveDimensionFile(levelDirPath, configFilePath, zipWriter, addFile)
	if err != nil {
		record.Error("%v", err)
		os.Exit(1)
	}

	err = parse.SaveCoreDataFile(levelDirPath, configFilePath, zipWriter, addFile)
	if err != nil {
		record.Error("%v", err)
		os.Exit(1)
	}

	// 关闭 zip 写入器，完成压缩包内容的写入
	err = zipWriter.Close()
	if err != nil {
		record.Error("%v", err)
		os.Exit(1)
	}

	// 关闭底层输出文件，确保数据完整落盘
	err = fileWriter.Close()
	if err != nil {
		record.Error("%v", err)
		os.Exit(1)
	}

	record.Info("Save completed!")
}

// 校验存档目录是否有效，需要是存在的目录且包含 level.dat 文件
func levelFileIsValid(levelDir string) error {

	// 获取存档目录的信息，判断目录是否存在
	archiveFileInfo, err := os.Stat(levelDir)
	if err != nil {
		return errors.Join(errors.New("(Get level file info) "), err)
	}

	// 若路径存在但不是目录，则视为无效存档
	if !archiveFileInfo.IsDir() {
		return errors.New("(Check world directory) level file isn't a directory")
	}

	// 检查存档目录下是否有 level.dat 文件，这是世界存档的标识文件
	_, err = os.Stat(path.Join(levelDir, "level.dat"))
	if err != nil {
		return errors.Join(errors.New("(Check whether \"level.dat\" file exists)"), err)
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
			return nil, nil, errors.Join(errors.New("(Create directory) "+outputDirPath+")"), err)
		}
	}

	// 创建输出文件，文件名格式为 存档名-日期.zip
	fileWriter, err := os.Create(path.Join(outputDirPath, archiveFileName))
	if err != nil {
		return nil, nil, errors.Join(errors.New("(Create archive file)"), err)
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
		record.Warn("%v", err)
		return nil
	}

	defer fileReader.Close()

	// 在压缩包中创建同名条目，保持与存档一致的目录结构
	file, err := zipWriter.Create(fileName)
	if err != nil {
		return errors.Join(errors.New("(Create new filename in archive)"), err)
	}

	// 将区域文件内容写入压缩包
	_, err = io.Copy(file, fileReader)
	if err != nil {
		return errors.Join(errors.New("(Write compress)"), err)
	}

	record.Info("(Added) %v", fileName)

	return nil

}
