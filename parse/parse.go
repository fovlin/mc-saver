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

type RootRule struct {
	Dimension map[string]DimensionRule `json:"dimension"`
	File      []string                 `json:"file"`
}

type DimensionRule struct {
	Range  []RangeRule `json:"range"`
	Simple [][2]int    `json:"simple"`
}

type RangeRule struct {
	From [2]int `json:"from"`
	To   [2]int `json:"to"`
}

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

	rootRule, err := LoadRootSaveRule(configFile)
	if err != nil {
		return err
	}

	for namespaceID, dimensionRule := range rootRule.Dimension {

		namespaceAndID := strings.FieldsFunc(namespaceID, isKeyWord)
		if len(namespaceAndID) != 2 {
			return errors.New("parse namespaceID: invalid namespace ID \"" + namespaceID + "\"")
		}

		namespace, dimensionID := namespaceAndID[0], namespaceAndID[1]
		dimensionRootDirPath := path.Join("dimensions", namespace, dimensionID)

		if dimensionRootDirStat, err := root.Stat(dimensionRootDirPath); err != nil {
			return fmt.Errorf("open dimension root directory: %w", err)
		} else if !dimensionRootDirStat.IsDir() {
			return fmt.Errorf("open dimension root directory: %v: %w", dimensionRootDirPath, errors.New("not a directory"))
		}

		for _, rangeRule := range dimensionRule.Range {
			for _, regionDataDir := range rootFile {
				for x := rangeRule.From[0]; x <= rangeRule.To[0]; x += 1 {
					for y := rangeRule.From[1]; y <= rangeRule.To[1]; y += 1 {
						regionFileName := formatRegionFilePath(dimensionRootDirPath, regionDataDir, x, y)
						err := addFile(root, regionFileName, zipWriter)
						if err != nil {
							return err
						}
					}
				}
			}
		}

		for _, regionDataDir := range rootFile {
			for _, simpleRule := range dimensionRule.Simple {
				regionFileName := formatRegionFilePath(dimensionRootDirPath, regionDataDir, simpleRule[0], simpleRule[1])
				err := addFile(root, regionFileName, zipWriter)
				if err != nil {
					return err
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
					return fmt.Errorf("open dimension directory: %w", err)
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
				return fmt.Errorf("open dimension directory: "+dimensionDataDirName+": %w", err)
			}
		}
	}
	return nil
}

func SaveRootDataFile(root *os.Root, configFile string, zipWriter *zip.Writer, addFile addFile) error {

	rootRule, err := LoadRootSaveRule(configFile)
	if err != nil {
		return err
	}

	for _, file := range rootRule.File {

		fileStat, err := root.Stat(file)
		if err != nil {
			return fmt.Errorf(":open file: %w", err)
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
					return fmt.Errorf("open dimension data directory: %w", err)
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
				return fmt.Errorf("open directory: %w", err)
			}
		}
	}
	return nil
}

func isKeyWord(char rune) bool {
	if char == rune(":"[0]) {
		return true
	} else {
		return false
	}
}

func formatRegionFilePath(dimensionRootDirPath string, regionDataDir string, x int, y int) string {
	regionFileName := "r." + strconv.FormatInt(int64(x), 10) + "." + strconv.FormatInt(int64(y), 10) + ".mca"
	regionFilePath := path.Join(dimensionRootDirPath, regionDataDir, regionFileName)
	return regionFilePath
}

func LoadRootSaveRule(configFile string) (RootRule, error) {

	jsonData, err := os.ReadFile(configFile)
	if err != nil {
		return RootRule{}, err
	}

	var rootRule RootRule
	err = json.Unmarshal(jsonData, &rootRule)
	if err != nil {
		return RootRule{}, err
	}

	return rootRule, nil

}
