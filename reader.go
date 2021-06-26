package dbmigrat

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ReadDir is helper func which allows for reading migrations from directory.
// Directory under provided path must contain files only.
// Files names must follow convention: a.b.c
// where a is incrementing int (0,1,2,3,..), b is description, c is direction - "up" or "down".
// Every migration must have corresponding up and down file.
// Up and down file for same migration must have same description.
//
// Examples of valid files names:
//	0.create_users_table.up
//	0.create_users_table.down.sql
//	1.add_username_column.up
//	1.add_username_column.down.sql
func ReadDir(fileSys fs.FS, path string) ([]Migration, error) {
	dirEntries, err := fs.ReadDir(fileSys, path)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			return nil, errContainsDirectory
		}
		fileNames = append(fileNames, dirEntry.Name())
	}

	parsedFN, err := parseFileNames(fileNames)
	if err != nil {
		return nil, err
	}
	var result []Migration
	for i := 0; i+1 < len(parsedFN); i += 2 {
		if parsedFN[i].idx != i/2 || parsedFN[i+1].idx != i/2 {
			return nil, errWithFileName{inner: errNotSequential, fileName: parsedFN[i].fileName}
		}
		if parsedFN[i].description != parsedFN[i+1].description {
			return nil, errWithFileName{inner: errDescriptionNotEqual, fileName: parsedFN[i].fileName}
		}
		if parsedFN[i].direction == parsedFN[i+1].direction {
			return nil, errWithFileName{inner: errSameDirections, fileName: parsedFN[i].fileName}
		}
		iData, err := fs.ReadFile(fileSys, filepath.Join(path, parsedFN[i].fileName))
		if err != nil {
			return nil, err
		}
		iPlus1Data, err := fs.ReadFile(fileSys, filepath.Join(path, parsedFN[i+1].fileName))
		if err != nil {
			return nil, err
		}
		result = append(result, Migration{
			Description: parsedFN[i].description,
			Up:          string(iData),
			Down:        string(iPlus1Data),
		})
	}

	return result, nil
}

func parseFileNames(fileNames []string) (parsedFileNames, error) {
	var parsedFN parsedFileNames
	for _, fileName := range fileNames {
		parsed, err := parseFileName(fileName)
		if err != nil {
			return nil, err
		}
		parsedFN = append(parsedFN, parsed)
	}
	sort.Sort(parsedFN)
	return parsedFN, nil
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

func (a parsedFileNames) Len() int      { return len(a) }
func (a parsedFileNames) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a parsedFileNames) Less(i, j int) bool {
	if a[i].idx == a[j].idx {
		return a[i].direction == up
	}
	return a[i].idx < a[j].idx
}

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
	errNotSequential       = errors.New("index in file name is not sequential (every migration has up and down file?)")
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
