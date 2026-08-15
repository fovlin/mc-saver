package parse

import (
	"acovia.net/record"
	"archive/zip"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"
)

// rootFile 表示每个维度下需要按坐标备份的三个区域文件目录
var (
	rootFile = []string{
		"region",
		"entities",
		"poi",
	}
)

// addFile 是由调用方注入的回调，负责将单个文件写入 zip 压缩包
type addFile func(fileName string, filePath string, zipWriter *zip.Writer) error

// SaveDimensionFile 根据配置文件解析各维度的 range/simple 备份规则，
// 将选中的区域文件与各维度 data 目录下的文件逐一写入 zip 压缩包。
func SaveDimensionFile(levelDir string, configFile string, zipWriter *zip.Writer, addFile addFile) error {

	rootSaveRule, err := getRootSaveRule(configFile)
	if err != nil {
		return err
	}

	// dimension 必须是对象：键为命名空间 ID，值为该维度的备份规则
	dimensionSaveRule, ok := rootSaveRule["dimension"].(map[string]any)
	if !ok {
		return errors.New("(Parse dimension rule) isn't a valid json object - dimension")
	}

	for namespaceID, saveRule := range dimensionSaveRule {

		// 使用函数解析维度的命名空间 ID 并拆解为命名空间和 ID
		namespaceAndID := strings.FieldsFunc(namespaceID, isKeyWord)
		if len(namespaceAndID) != 2 {
			return errors.New("(Verify \"namespaceID\") isn't a valid namespaceID - " + "\"" + namespaceID + "\"")
		}

		namespace, dimensionID := namespaceAndID[0], namespaceAndID[1]

		// 对维度规则对象进行断言，它对应 JSON 文件里命名空间 ID 下的配置
		saveRule, ok := saveRule.(map[string]any)
		if !ok {
			return errors.New("(Parse dimension rule) isn't a valid json object - \"range\"")
		}

		// range 规则如果存在，则根据 range 规则进行备份
		if saveRule["range"] != nil {

			rangeRuleList, ok := saveRule["range"].([]any)
			if !ok {
				return errors.New("(Parse range rule) isn't valid json array - \"range\"")
			}

			for rangeRuleIndex, rangeRule := range rangeRuleList {
				rangeRule, ok := rangeRule.(map[string]any)
				if !ok {
					return errors.New("(Parse range rule) isn't valid json object - \"range\" - " + strconv.Itoa(rangeRuleIndex))
				}

				jsonFrom, ok := rangeRule["from"].([]any)
				if !ok {
					return errors.New("(Parse from rule) isn't valid json array - \"from\" - " + strconv.Itoa(rangeRuleIndex))
				}

				jsonTo, ok := rangeRule["to"].([]any)
				if !ok {
					return errors.New("(Parse to rule) isn't valid json array - \"to\" - " + strconv.Itoa(rangeRuleIndex))
				}

				// 将 JSON 中的坐标从 float64 转换为 int64 并校验长度，为后续坐标遍历做准备
				var from []int64
				var to []int64

				// 校验和更新数组类型
				for _, number := range jsonFrom {
					jsonFromValue, ok := number.(float64)
					if !ok {
						return errors.New("(Verify \"from\" rule) \"from\" rule contains data that isn't a number")
					}
					from = append(from, int64(jsonFromValue))
				}
				if len(from) != 2 {
					return errors.New("(Verify from rule) \"from\" isn't a array of length is 2")
				}

				for _, number := range jsonTo {
					jsonToValue, ok := number.(float64)
					if !ok {
						return errors.New("(Verify \"to\" rule) \"to\" rule contains data that isn't a number")
					}
					to = append(to, int64(jsonToValue))
				}
				if len(to) != 2 {
					return errors.New("(Verify to rule) \"to\" isn't a array of length is 2")
				}

				// 遍历区域目录（region/entities/poi），再遍历 from 到 to 之间的 x，y 坐标，逐一保存区域文件
				for _, regionDataDir := range rootFile {
					for x := from[0]; x <= to[0]; x += 1 {
						for y := from[1]; y <= to[1]; y += 1 {
							// 压缩包内以存档名作为顶层目录，保证备份可直接还原为存档
							regionFilePath := formatRegionFilePath(levelDir, namespace, dimensionID, regionDataDir, x, y)
							regionFileName := formatRegionFilePath(path.Base(levelDir), namespace, dimensionID, regionDataDir, x, y)
							// 单个文件出错只记录日志，不中断整个备份（有意为之）
							err := addFile(regionFileName, regionFilePath, zipWriter)
							if err != nil {
								record.Error("%v", err)
							}
						}
					}
				}
			}
		}

		// simple 规则如果存在，则根据 simple 规则进行备份
		if saveRule["simple"] != nil {
			simpleRuleList, ok := saveRule["simple"].([]any)
			if !ok {
				return errors.New("(Parse \"simple\" rule) isn't a json array - \"simple\"")
			}

			// simple 规则：按给定坐标逐一写入三个区域目录中的文件
			for _, regionDataDir := range rootFile {
				for simpleRuleIndex, simpleRule := range simpleRuleList {

					// 对规则进行断言，取出 x，y 坐标，判断是否为数组并且长度为2
					simpleRule, ok := simpleRule.([]any)
					if !ok {
						return errors.New("(Verify \"simple\" rule) isn't a json array - \"simple\"")
					}
					if len(simpleRule) != 2 {
						return errors.New("(Verify \"simple\" rule) isn't a array that length is 2 - \"simple\" - index " + strconv.Itoa(simpleRuleIndex))
					}

					jsonX, Index1ok := simpleRule[0].(float64)
					x := int64(jsonX)
					jsonY, Index2ok := simpleRule[1].(float64)
					y := int64(jsonY)
					if !Index1ok || !Index2ok {
						return errors.New("(Verify \"simple\" rule) \"simple\" rule contains data that isn't number")
					}
					// 压缩包内以存档名作为顶层目录，保证备份可直接还原为存档
					regionFilePath := formatRegionFilePath(levelDir, namespace, dimensionID, regionDataDir, x, y)
					regionFileName := formatRegionFilePath(path.Base(levelDir), namespace, dimensionID, regionDataDir, x, y)
					// 出错只记录日志，不中断备份（与 range 规则一致）
					err := addFile(regionFileName, regionFilePath, zipWriter)
					if err != nil {
						record.Error("%v", err)
					}
				}
			}
		}

		// 备份维度 data 目录：磁盘路径用于读取，压缩包内路径保留存档名层级，便于直接还原
		dimensionDataDirPath := path.Join(levelDir, "dimensions", namespace, dimensionID, "data")
		dimensionDataDirName := path.Join(path.Base(levelDir), "dimensions", namespace, dimensionID, "data")
		dimensionDataFileSystem := os.DirFS(dimensionDataDirPath)
		err := fs.WalkDir(dimensionDataFileSystem, ".", func(subFilePath string, d fs.DirEntry, err error) error {
			// WalkDir 的遍历错误（如该维度没有 data 目录）仅跳过，不中断备份（有意为之）
			if err == nil {
				endFilePath := path.Join(dimensionDataDirPath, subFilePath)
				dataFileName := path.Join(dimensionDataDirName, subFilePath)
				subFileStat, err := os.Stat(endFilePath)
				if err == nil {
					if !subFileStat.IsDir() {
						err := addFile(dataFileName, endFilePath, zipWriter)
						if err != nil {
							record.Error("%v", err)
						}
					}
				}
			}
			return nil
		})
		if err != nil {
			return errors.Join(errors.New("(Read directory) error in: "+dimensionDataDirPath), err)
		}
	}
	return nil
}

// SaveCoreDataFile 遍历配置中 file 列表的条目：普通文件直接写入 zip，
// 目录则递归遍历其中的全部文件一并写入。
func SaveCoreDataFile(levelDir string, configFile string, zipWriter *zip.Writer, addFile addFile) error {

	rootRule, err := getRootSaveRule(configFile)
	if err != nil {
		return err
	}

	fileList, ok := rootRule["file"].([]any)
	if !ok {
		return errors.New("(Parse file rule) is not a valid json object - file")
	}

	for _, file := range fileList {

		file, ok := file.(string)
		if !ok {
			return errors.New("(Parse file rule) is not a valid json object - file")
		}

		filePath := path.Join(levelDir, file)
		// 压缩包内以存档名作为顶层目录，保证备份可直接还原为存档
		fileName := path.Join(path.Base(levelDir), file)
		fileStat, err := os.Stat(filePath)
		// 文件或目录不存在时跳过，不中断备份（有意为之：用户目录可能缺少文件）
		if err != nil {
			continue
		}

		switch fileStat.IsDir() {

		case false:
			// 单个文件出错只记录日志，不中断备份（有意为之）
			err := addFile(fileName, filePath, zipWriter)
			if err != nil {
				record.Error("%v", err)
			}
		case true:
			fileSystem := os.DirFS(filePath)
			err := fs.WalkDir(fileSystem, ".", func(subFilePath string, d fs.DirEntry, err error) error {
				// 遍历错误（如子目录不可读）仅跳过，不中断备份（有意为之）
				if err == nil {
					fullFilePath := path.Join(filePath, subFilePath)
					fullFileName := path.Join(fileName, subFilePath)
					subFileStat, err := os.Stat(fullFilePath)
					if err == nil {
						if !subFileStat.IsDir() {
							err := addFile(fullFileName, fullFilePath, zipWriter)
							if err != nil {
								record.Error("%v", err)
							}
						}
					}
				}
				return nil
			})
			if err != nil {
				return errors.Join(err, errors.New("(Read directory)"+"error in: "+filePath))
			}
		}
	}
	return nil
}

// 解析 JSON 获得各维度保存规则
func getRootSaveRule(configFile string) (map[string]any, error) {

	// 读取规则文件的内容
	jsonData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, errors.Join(errors.New("(Read config file) unable to read json file: "+configFile), err)
	}

	// 将 JSON 内容解析为映射表
	rootRule := make(map[string]any)
	err = json.Unmarshal(jsonData, &rootRule)
	if err != nil {
		return nil, errors.Join(errors.New("(Parse json data) unable to parse json data: "+configFile), err)
	}

	return rootRule, nil
}

// isKeyWord 判断字符是否为 ':'，配合 FieldsFunc 按 ':' 拆分命名空间 ID
func isKeyWord(char rune) bool {
	if char == rune(":"[0]) {
		return true
	} else {
		return false
	}
}

// formatRegionFilePath 拼接区域文件的路径，文件名格式为 r.<x>.<y>.mca。
// 传入 levelDir 时生成磁盘路径；传入 path.Base(levelDir) 时生成压缩包内路径。
func formatRegionFilePath(levelDir string, namespace string, dimensionID string, regionDataDir string, x int64, y int64) string {
	regionFilePathName := "r." + strconv.FormatInt(x, 10) + "." + strconv.FormatInt(y, 10) + ".mca"
	regionFilePath := path.Join(levelDir, "dimensions", namespace, dimensionID, regionDataDir, regionFilePathName)
	return regionFilePath
}
