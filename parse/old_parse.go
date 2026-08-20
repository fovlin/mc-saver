package parse

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"
)

// SaveOldAllFile 按旧版单文件夹布局备份：主世界位于存档根目录，下界/末地分别为 DIM-1/DIM1
func SaveOldAllFile(root *os.Root, configFile string, zipWriter *zip.Writer, addFile addFile) (err error) {

	if err := SaveOldDimensionFile(root, configFile, zipWriter, addFile); err != nil {
		return err
	}

	if err := SaveRootDataFile(root, configFile, zipWriter, addFile); err != nil {
		return err
	}

	return nil
}

// SaveOldDimensionFile 按旧版布局解析各维度的 range/simple 备份规则，
// 将选中的区域文件与各维度 data 目录下的文件逐一写入 zip 压缩包。
func SaveOldDimensionFile(root *os.Root, configFile string, zipWriter *zip.Writer, addFile addFile) error {

	rootSaveRule, err := getRootSaveRule(configFile)
	if err != nil {
		return err
	}

	// dimension 必须是对象：键为命名空间 ID，值为该维度的备份规则
	dimensionSaveRule, ok := rootSaveRule["dimension"].(map[string]any)
	if !ok {
		return errors.New("(parse dimension rule) dimension is not a valid JSON object")
	}

	for namespaceID, saveRule := range dimensionSaveRule {

		// 使用函数解析维度的命名空间 ID 并拆解为命名空间和 ID
		namespaceAndID := strings.FieldsFunc(namespaceID, isKeyWord)
		if len(namespaceAndID) != 2 {
			return errors.New("(verify namespaceID) invalid namespace ID \"" + namespaceID + "\"")
		}

		namespace, dimensionID := namespaceAndID[0], namespaceAndID[1]

		var dimensionRootDirPath string

		// 旧版布局的维度目录映射：主世界为根目录，下界/末地为 DIM-1/DIM1，自定义维度仍用 dimensions/
		switch namespaceID {
		case "minecraft:overworld":
			dimensionRootDirPath = "."
		case "minecraft:the_nether":
			dimensionRootDirPath = "DIM-1"
		case "minecraft:the_end":
			dimensionRootDirPath = "DIM1"
		default:
			dimensionRootDirPath = path.Join("dimensions", namespace, dimensionID)
		}

		_, err := root.Stat(dimensionRootDirPath)
		if err != nil {
			return fmt.Errorf("(read dimension root directory) %w", err)
		}

		// 对维度规则对象进行断言，它对应 JSON 文件里命名空间 ID 下的配置
		saveRule, ok := saveRule.(map[string]any)
		if !ok {
			return errors.New("(parse dimension rule) rule for \"" + namespaceID + "\" is not a valid JSON object")
		}

		// range 规则如果存在，则根据 range 规则进行备份
		if saveRule["range"] != nil {

			rangeRuleList, ok := saveRule["range"].([]any)
			if !ok {
				return errors.New("(parse range rule) range is not a valid JSON array")
			}

			for rangeRuleIndex, rangeRule := range rangeRuleList {
				rangeRule, ok := rangeRule.(map[string]any)
				if !ok {
					return errors.New("(parse range rule) range entry at index " + strconv.Itoa(rangeRuleIndex) + " is not a valid JSON object")
				}

				jsonFrom, ok := rangeRule["from"].([]any)
				if !ok {
					return errors.New("(parse from rule) from is not a valid JSON array in range entry at index " + strconv.Itoa(rangeRuleIndex))
				}

				jsonTo, ok := rangeRule["to"].([]any)
				if !ok {
					return errors.New("(parse to rule) to is not a valid JSON array in range entry at index " + strconv.Itoa(rangeRuleIndex))
				}

				// 将 JSON 中的坐标从 float64 转换为 int64 并校验长度，为后续坐标遍历做准备
				var from []int64
				var to []int64

				// 校验和更新数组类型
				for _, number := range jsonFrom {
					jsonFromValue, ok := number.(float64)
					if !ok {
						return errors.New("(verify from rule) from contains a value that is not a number")
					}
					from = append(from, int64(jsonFromValue))
				}
				if len(from) != 2 {
					return errors.New("(verify from rule) from must be an array of length 2")
				}

				for _, number := range jsonTo {
					jsonToValue, ok := number.(float64)
					if !ok {
						return errors.New("(verify to rule) to contains a value that is not a number")
					}
					to = append(to, int64(jsonToValue))
				}
				if len(to) != 2 {
					return errors.New("(verify to rule) to must be an array of length 2")
				}

				// 遍历区域目录（region/entities/poi），再遍历 from 到 to 之间的 x，y 坐标，逐一保存区域文件
				for _, regionDataDir := range rootFile {
					for x := from[0]; x <= to[0]; x += 1 {
						for y := from[1]; y <= to[1]; y += 1 {
							mcaFileName := "r." + strconv.FormatInt(x, 10) + "." + strconv.FormatInt(y, 10) + ".mca"
							// 压缩包内以存档名作为顶层目录，保证备份可直接还原为存档
							regionFileName := path.Join(dimensionRootDirPath, regionDataDir, mcaFileName)

							// 单个文件出错只记录日志，不中断整个备份（有意为之）
							err := addFile(root, regionFileName, zipWriter)
							if err != nil {
								return fmt.Errorf("(write file into archive) %w", err)
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
				return errors.New("(parse simple rule) simple is not a valid JSON array")
			}

			// simple 规则：按给定坐标逐一写入三个区域目录中的文件
			for _, regionDataDir := range rootFile {
				for simpleRuleIndex, simpleRule := range simpleRuleList {

					// 对规则进行断言，取出 x，y 坐标，判断是否为数组并且长度为2
					simpleRule, ok := simpleRule.([]any)
					if !ok {
						return errors.New("(verify simple rule) simple entry is not a valid JSON array")
					}
					if len(simpleRule) != 2 {
						return errors.New("(verify simple rule) simple entry at index " + strconv.Itoa(simpleRuleIndex) + " must be an array of length 2")
					}

					jsonX, Index1ok := simpleRule[0].(float64)
					x := int64(jsonX)
					jsonY, Index2ok := simpleRule[1].(float64)
					y := int64(jsonY)
					if !Index1ok || !Index2ok {
						return errors.New("(verify simple rule) simple contains a value that is not a number")
					}

					mcaFileName := "r." + strconv.FormatInt(x, 10) + "." + strconv.FormatInt(y, 10) + ".mca"
					// 压缩包内以存档名作为顶层目录，保证备份可直接还原为存档
					regionFileName := path.Join(dimensionRootDirPath, regionDataDir, mcaFileName)

					err := addFile(root, regionFileName, zipWriter)
					if err != nil {
						return fmt.Errorf("(write file into archive) %w", err)
					}
				}
			}
		}

		// 备份维度 data 目录：相对路径用于读取，压缩包内路径由 addFile 保留存档名层级，便于直接还原
		dimensionDataDirName := path.Join(dimensionRootDirPath, "data")

		// 维度数据文件若不存在，跳过，有意为之
		_, err = root.Stat(dimensionDataDirName)
		if err == nil {
			dimensionDataRootDir, err := root.OpenRoot(dimensionDataDirName)
			if err != nil {
				return err
			}
			defer dimensionDataRootDir.Close()

			err = fs.WalkDir(dimensionDataRootDir.FS(), ".", func(subFilePath string, d fs.DirEntry, err error) error {
				if err != nil {
					return fmt.Errorf("(read dimension directory) %w", err)
				}

				dataFileName := path.Join(dimensionDataDirName, subFilePath)

				// 目录由 WalkDir 递归遍历，这里只写入普通文件
				if !d.IsDir() {
					err := addFile(root, dataFileName, zipWriter)
					if err != nil {
						return fmt.Errorf("(write file into archive) %w", err)
					}
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("(read dimension directory) "+dimensionDataDirName+": %w", err)
			}
		}
	}
	return nil
}
