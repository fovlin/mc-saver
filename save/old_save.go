package save

import (
	"archive/zip"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

func SaveOldAllFile(root *os.Root, config Config, zipWriter *zip.Writer, saveFile saveFile) (err error) {

	if err := SaveOldDimensionFile(root, config, zipWriter, saveFile); err != nil {
		return err
	}

	if err := SaveRootDataFile(root, config, zipWriter, saveFile); err != nil {
		return err
	}

	return nil
}

func SaveOldDimensionFile(root *os.Root, config Config, zipWriter *zip.Writer, saveFile saveFile) error {

	for namespaceID, dimensionRule := range config.Dimension {

		namespaceAndID := strings.FieldsFunc(namespaceID, IsNamespaceKeyWord)
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
			return fmt.Errorf("open dimension root directory: %w", err)
		} else if !dimensionRootDirStat.IsDir() {
			return fmt.Errorf("open dimension root directory: %v: %w", dimensionRootDirPath, errors.New("not a directory"))
		}

		for _, rangeRule := range dimensionRule.Range {
			for _, regionDataDir := range rootFile {
				iterator := NewRangeIterator(rangeRule.From.X, rangeRule.From.Y, rangeRule.To.X, rangeRule.To.Y)
				err := iterator.run(func(x int, y int) error {
					regionFileName := FormatRegionFilePath(dimensionRootDirPath, regionDataDir, x, y)
					err := saveFile(regionFileName, zipWriter)
					if err != nil {
						return err
					}
					return nil
				})
				if err != nil {
					return err
				}
			}
		}

		for _, regionDataDir := range rootFile {
			for _, simpleRule := range dimensionRule.Simple {
				regionFileName := FormatRegionFilePath(dimensionRootDirPath, regionDataDir, simpleRule.X, simpleRule.Y)
				err := saveFile(regionFileName, zipWriter)
				if err != nil {
					return err
				}
			}
		}

		dimensionDataDirName := path.Join(dimensionRootDirPath, "data")
		_, err := root.Stat(dimensionDataDirName)

		if os.IsNotExist(err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("open dimension data directory: %w", err)
		}

		err = saveDirInRoot(zipWriter, root, dimensionDataDirName, saveFile)
		if err != nil {
			return fmt.Errorf("save directory: %w", err)
		}
	}
	return nil
}
