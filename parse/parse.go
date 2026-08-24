package parse

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	rootFile = []string{
		"region",
		"entities",
		"poi",
	}
)

func SaveAllFile(root *os.Root, configFile string, zipWriter *zip.Writer, addFile addFile) (err error) {

	if err := SaveDimensionFile(root, configFile, zipWriter, addFile); err != nil {
		return err
	}

	if err := SaveRootDataFile(root, configFile, zipWriter, addFile); err != nil {
		return err
	}

	return nil
}

type addFile func(root *os.Root, fileName string, zipWriter *zip.Writer) error

func SaveDimensionFile(root *os.Root, configFile string, zipWriter *zip.Writer, addFile addFile) error {

	rootSaveRule, err := getRootSaveRule(configFile)
	if err != nil {
		return err
	}

	dimensionSaveRuleList, ok := rootSaveRule["dimension"].(map[string]any)
	if !ok {
		return errors.New("(parse \"dimension\" rule) \"dimension\" is not a valid JSON object")
	}

	for namespaceID, dimensionSaveRule := range dimensionSaveRuleList {

		namespaceAndID := strings.FieldsFunc(namespaceID, isKeyWord)
		if len(namespaceAndID) != 2 {
			return errors.New("(parse namespaceID) invalid namespace ID \"" + namespaceID + "\"")
		}

		namespace, dimensionID := namespaceAndID[0], namespaceAndID[1]
		dimensionRootDirPath := path.Join("dimensions", namespace, dimensionID)

		_, err := root.Stat(dimensionRootDirPath)
		if err != nil {
			return fmt.Errorf("(open dimension root directory) %w", err)
		}

		dimensionSaveRule, ok := dimensionSaveRule.(map[string]any)
		if !ok {
			return errors.New("(parse \"dimension\" rule) rule of \"" + namespaceID + "\" is not a valid JSON object")
		}

		if dimensionSaveRule["range"] != nil {

			rangeRuleList, ok := dimensionSaveRule["range"].([]any)
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

				var from []int64
				var to []int64

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
							regionFileName := formatRegionFilePath(dimensionRootDirPath, regionDataDir, x, y)
							err := addFile(root, regionFileName, zipWriter)
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}

		if dimensionSaveRule["simple"] != nil {
			simpleRuleList, ok := dimensionSaveRule["simple"].([]any)
			if !ok {
				return errors.New("(parse \"simple\" rule) \"simple\" is not a valid JSON array")
			}

			for _, regionDataDir := range rootFile {
				for simpleRuleIndex, simpleRule := range simpleRuleList {

					simpleRule, ok := simpleRule.([]any)
					if !ok {
						return errors.New("(parse \"simple\" rule) \"simple\" entry is not a valid JSON array")
					}
					if len(simpleRule) != 2 {
						return errors.New("(parse \"simple\" rule) \"simple\" entry at index " + strconv.Itoa(simpleRuleIndex) + " must be an array of length 2")
					}

					jsonX, Index1ok := simpleRule[0].(float64)
					x := int64(jsonX)
					jsonY, Index2ok := simpleRule[1].(float64)
					y := int64(jsonY)
					if !Index1ok || !Index2ok {
						return errors.New("(parse \"simple\" rule) \"simple\" contains a value that is not a number")
					}

					regionFileName := formatRegionFilePath(dimensionRootDirPath, regionDataDir, x, y)
					err := addFile(root, regionFileName, zipWriter)
					if err != nil {
						return err
					}
				}
			}
		}

		dimensionDataDirName := path.Join("dimensions", namespace, dimensionID, "data")

		_, err = root.Stat(dimensionDataDirName)
		if !os.IsNotExist(err) {
			dimensionDataRootDir, err := root.OpenRoot(dimensionDataDirName)
			if err != nil {
				return err
			}
			defer dimensionDataRootDir.Close()

			err = fs.WalkDir(dimensionDataRootDir.FS(), ".", func(subFilePath string, d fs.DirEntry, err error) error {

				if err != nil {
					return fmt.Errorf("(open dimension directory) %w", err)
				}

				dataFileName := path.Join(dimensionDataDirName, subFilePath)

				if !d.IsDir() {
					err := addFile(root, dataFileName, zipWriter)
					if err != nil {
						return fmt.Errorf("%w", err)
					}
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("(open dimension directory) "+dimensionDataDirName+": %w", err)
			}
		}
	}
	return nil
}

func SaveRootDataFile(root *os.Root, configFile string, zipWriter *zip.Writer, addFile addFile) error {

	rootRule, err := getRootSaveRule(configFile)
	if err != nil {
		return err
	}

	fileRule, ok := rootRule["file"]
	if !ok {
		return nil
	}

	fileList, ok := fileRule.([]any)
	if !ok {
		return errors.New("(parse \"file\" rule) file is not a valid JSON array")
	}

	for _, file := range fileList {

		file, ok := file.(string)
		if !ok {
			return errors.New("(parse \"file\" rule) file contains a value that is not a string")
		}

		fileStat, err := root.Stat(file)
		if err != nil {
			return fmt.Errorf("(open file) %w", err)
		}

		switch fileStat.IsDir() {

		case false:
			err := addFile(root, file, zipWriter)
			if err != nil {
				return err
			}
		case true:
			rootDataRootDir, err := root.OpenRoot(file)
			if err != nil {
				return err
			}
			defer rootDataRootDir.Close()

			err = fs.WalkDir(rootDataRootDir.FS(), ".", func(subFilePath string, d fs.DirEntry, err error) error {
				if err != nil {
					return fmt.Errorf("(open dimension data directory) %w", err)
				}

				fullFilePath := path.Join(file, subFilePath)

				if !d.IsDir() {
					err := addFile(root, fullFilePath, zipWriter)
					if err != nil {
						return err
					}
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("(open directory) %w", err)
			}
		}
	}
	return nil
}

func getRootSaveRule(configFile string) (map[string]any, error) {

	jsonData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("(open config file) %w", err)
	}

	rootRule := make(map[string]any)
	err = json.Unmarshal(jsonData, &rootRule)
	if err != nil {
		return nil, fmt.Errorf("(parse json data) %w", err)
	}

	return rootRule, nil
}

func isKeyWord(char rune) bool {
	if char == rune(":"[0]) {
		return true
	} else {
		return false
	}
}

func formatRegionFilePath(dimensionRootDirPath string, regionDataDir string, x int64, y int64) string {
	regionFileName := "r." + strconv.FormatInt(x, 10) + "." + strconv.FormatInt(y, 10) + ".mca"
	regionFilePath := path.Join(dimensionRootDirPath, regionDataDir, regionFileName)
	return regionFilePath
}
