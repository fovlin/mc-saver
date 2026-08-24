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

func SaveOldAllFile(root *os.Root, configFile string, zipWriter *zip.Writer, addFile addFile) (err error) {

	if err := SaveOldDimensionFile(root, configFile, zipWriter, addFile); err != nil {
		return err
	}

	if err := SaveRootDataFile(root, configFile, zipWriter, addFile); err != nil {
		return err
	}

	return nil
}

func SaveOldDimensionFile(root *os.Root, configFile string, zipWriter *zip.Writer, addFile addFile) error {

	rootSaveRule, err := getRootSaveRule(configFile)
	if err != nil {
		return err
	}

	dimensionSaveRule, ok := rootSaveRule["dimension"].(map[string]any)
	if !ok {
		return errors.New("(parse \"dimension\" rule) dimension is not a valid JSON object")
	}

	for namespaceID, saveRule := range dimensionSaveRule {

		namespaceAndID := strings.FieldsFunc(namespaceID, isKeyWord)
		if len(namespaceAndID) != 2 {
			return errors.New("(parse namespaceID) invalid namespace ID \"" + namespaceID + "\"")
		}

		namespace, dimensionID := namespaceAndID[0], namespaceAndID[1]

		var dimensionRootDirPath string

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
			return fmt.Errorf("(open dimension root directory) %w", err)
		}

		saveRule, ok := saveRule.(map[string]any)
		if !ok {
			return errors.New("(parse \"dimension\" rule) rule of \"" + namespaceID + "\" is not a valid JSON object")
		}

		if saveRule["range"] != nil {

			rangeRuleList, ok := saveRule["range"].([]any)
			if !ok {
				return errors.New("(parse \"range\" rule) \"range\" is not a valid JSON array")
			}

			for rangeRuleIndex, rangeRule := range rangeRuleList {
				rangeRule, ok := rangeRule.(map[string]any)
				if !ok {
					return errors.New("(parse \"range\" rule) \"range\" entry at index " + strconv.Itoa(rangeRuleIndex) + " is not a valid JSON object")
				}

				jsonFrom, ok := rangeRule["from"].([]any)
				if !ok {
					return errors.New("(parse \"from\" rule) \"from\" is not a valid JSON array in range entry at index " + strconv.Itoa(rangeRuleIndex))
				}

				jsonTo, ok := rangeRule["to"].([]any)
				if !ok {
					return errors.New("(parse \"to\" rule) \"to\" is not a valid JSON array in range entry at index " + strconv.Itoa(rangeRuleIndex))
				}

				// 将 JSON 中的坐标从 float64 转换为 int64 并校验长度，为后续坐标遍历做准备
				var from []int64
				var to []int64

				// 校验和更新数组类型
				for _, number := range jsonFrom {
					jsonFromValue, ok := number.(float64)
					if !ok {
						return errors.New("(parse \"from\" rule) \"from\" contains a value that is not a number")
					}
					from = append(from, int64(jsonFromValue))
				}
				if len(from) != 2 {
					return errors.New("(parse \"from\" rule) \"from\" must be an array of length 2")
				}

				for _, number := range jsonTo {
					jsonToValue, ok := number.(float64)
					if !ok {
						return errors.New("(parse \"to\" rule) \"to\" contains a value that is not a number")
					}
					to = append(to, int64(jsonToValue))
				}
				if len(to) != 2 {
					return errors.New("(parse \"to\" rule) \"to\" must be an array of length 2")
				}

				for _, regionDataDir := range rootFile {
					for x := from[0]; x <= to[0]; x += 1 {
						for y := from[1]; y <= to[1]; y += 1 {
							mcaFileName := "r." + strconv.FormatInt(x, 10) + "." + strconv.FormatInt(y, 10) + ".mca"
							regionFileName := path.Join(dimensionRootDirPath, regionDataDir, mcaFileName)

							err := addFile(root, regionFileName, zipWriter)
							if err != nil {
								return err
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
				return errors.New("(parse \"simple\" rule) \"simple\" is not a valid JSON array")
			}

			// simple 规则：按给定坐标逐一写入三个区域目录中的文件
			for _, regionDataDir := range rootFile {
				for simpleRuleIndex, simpleRule := range simpleRuleList {

					// 对规则进行断言，取出 x，y 坐标，判断是否为数组并且长度为2
					simpleRule, ok := simpleRule.([]any)
					if !ok {
						return errors.New("(verify \"simple\" rule) \"simple\" entry is not a valid JSON array")
					}
					if len(simpleRule) != 2 {
						return errors.New("(verify \"simple\" rule) \"simple\" entry at index " + strconv.Itoa(simpleRuleIndex) + " must be an array of length 2")
					}

					jsonX, Index1ok := simpleRule[0].(float64)
					x := int64(jsonX)
					jsonY, Index2ok := simpleRule[1].(float64)
					y := int64(jsonY)
					if !Index1ok || !Index2ok {
						return errors.New("(verify \"simple\" rule) \"simple\" contains a value that is not a number")
					}

					mcaFileName := "r." + strconv.FormatInt(x, 10) + "." + strconv.FormatInt(y, 10) + ".mca"
					regionFileName := path.Join(dimensionRootDirPath, regionDataDir, mcaFileName)

					err := addFile(root, regionFileName, zipWriter)
					if err != nil {
						return fmt.Errorf("(add file) %w", err)
					}
				}
			}
		}

		// 备份维度 data 目录：相对路径用于读取，压缩包内路径由 addFile 保留存档名层级，便于直接还原
		dimensionDataDirName := path.Join(dimensionRootDirPath, "data")

		_, err = root.Stat(dimensionDataDirName)
		if !os.IsNotExist(err) {
			dimensionDataRootDir, err := root.OpenRoot(dimensionDataDirName)
			if err != nil {
				return fmt.Errorf("(open dimension data directory) %w", err)
			}
			defer dimensionDataRootDir.Close()

			err = fs.WalkDir(dimensionDataRootDir.FS(), ".", func(subFilePath string, d fs.DirEntry, err error) error {
				if err != nil {
					return fmt.Errorf("(open dimension directory) %w", err)
				}

				dataFileName := path.Join(dimensionDataDirName, subFilePath)

				// 目录由 WalkDir 递归遍历，这里只写入普通文件
				if !d.IsDir() {
					err := addFile(root, dataFileName, zipWriter)
					if err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("(open dimension directory) %w", err)
			}
		}
	}
	return nil
}
