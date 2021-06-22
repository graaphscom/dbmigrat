package dbmigrat

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func ReadDir(embedFs embed.FS, path string) ([]Migration, error) {
	dirEntries, err := embedFs.ReadDir(path)
	if err != nil {
		return nil, err
	}
	parsedFileNames, err := parseFileNames(dirEntries)
	if err != nil {
		return nil, err
	}
	var result []Migration
	for i := 0; i+1 < len(parsedFileNames); i += 2 {
		if parsedFileNames[i].idx != i || parsedFileNames[i+1].idx != i {
			return nil, errWithFileName{inner: errNotSequential, fileName: parsedFileNames[i].fileName}
		}
		if parsedFileNames[i].description != parsedFileNames[i+1].description {
			return nil, errWithFileName{inner: errDescriptionNotEqual, fileName: parsedFileNames[i].fileName}
		}
		if parsedFileNames[i].direction == parsedFileNames[i+1].direction {
			return nil, errWithFileName{inner: errSameDirections, fileName: parsedFileNames[i].fileName}
		}
		iData, err := os.ReadFile(filepath.Join(path, parsedFileNames[i].fileName))
		if err != nil {
			return nil, err
		}
		iPlus1Data, err := os.ReadFile(filepath.Join(path, parsedFileNames[i+1].fileName))
		if err != nil {
			return nil, err
		}
		var upSql, downSql []byte
		if parsedFileNames[i].direction == up {
			upSql, downSql = iData, iPlus1Data
		} else {
			upSql, downSql = iPlus1Data, iData
		}
		result = append(result, Migration{
			Description: parsedFileNames[i].description,
			Up:          string(upSql),
			Down:        string(downSql),
		})
	}

	return result, nil
}

func parseFileNames(dirEntries []fs.DirEntry) (parsedFileNames, error) {
	var parsedFileNames parsedFileNames
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			return nil, errContainsDirectory
		}
		parsed, err := parseFileName(dirEntry.Name())
		if err != nil {
			return nil, err
		}
		parsedFileNames = append(parsedFileNames, parsed)
	}
	sort.Sort(parsedFileNames)
	return parsedFileNames, nil
}

func parseFileName(fileName string) (*parsedFileName, error) {
	divided := strings.Split(fileName, ".")
	if len(divided) < 3 {
		return nil, errFileNameParts
	}
	idx, err := strconv.Atoi(divided[0])
	if err != nil {
		return nil, errFileNameIdx
	}
	if divided[2] != string(up) && divided[2] != string(down) {
		return nil, errFileNameDirection
	}
	return &parsedFileName{
		fileName:    fileName,
		idx:         idx,
		description: divided[1],
		direction:   direction(divided[2]),
	}, nil
}

func (a parsedFileNames) Len() int           { return len(a) }
func (a parsedFileNames) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a parsedFileNames) Less(i, j int) bool { return a[i].idx < a[j].idx }

type parsedFileNames []*parsedFileName

type parsedFileName struct {
	fileName    string
	idx         int
	description string
	direction   direction
}

const (
	up   direction = "up"
	down direction = "down"
)

type direction string

var (
	errFileNameParts       = errors.New("migration's file name must contain at least 3 parts (idx.description.direction)")
	errFileNameIdx         = errors.New("first part of migration's file name must be int")
	errFileNameDirection   = errors.New(`third part of migration's file name must be "up" or "down" (case sensitive)`)
	errContainsDirectory   = errors.New("migrations directory should contain files only")
	errNotSequential       = errors.New("index in file name is not sequential")
	errDescriptionNotEqual = errors.New("descriptions for migration differs")
	errSameDirections      = errors.New("migration must have up and down files")
)

func (e errWithFileName) Error() string {
	return fmt.Sprintf("%s (%s)", e.inner.Error(), e.fileName)
}

type errWithFileName struct {
	inner    error
	fileName string
}
