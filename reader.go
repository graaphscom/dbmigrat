package dbmigrat

import (
	"embed"
	"errors"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

func ReadDir(fs embed.FS, path string) ([]Migration, error) {
	//dirEntries, err := fs.ReadDir(path)
	//if err != nil {
	//	return nil, err
	//}

	return nil, nil
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

var (
	errFileNameParts     = errors.New("migration's file name must contain at least 3 parts (idx.description.direction)")
	errFileNameIdx       = errors.New("first part of migration's file name must be int")
	errFileNameDirection = errors.New(`third part of migration's file name must be "up" or "down" (case sensitive)`)
	errContainsDirectory = errors.New("migrations directory should contain files only")
)

type direction string
