package parse

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
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

	rootRule, err := LoadRootSaveRule(configFile)
	if err != nil {
		return err
	}

	for namespaceID, dimensionRule := range rootRule.Dimension {

		namespaceAndID := strings.FieldsFunc(namespaceID, isNamespaceKeyWord)
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

		if dimensionRootDirStat, err := root.Stat(dimensionRootDirPath); err != nil {
			return fmt.Errorf("(open dimension root directory) %w", err)
		} else if !dimensionRootDirStat.IsDir() {
			return fmt.Errorf("(open dimension root directory) %v: %w", dimensionRootDirPath, errors.New("not a directory"))
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