package dbmigrat

import (
	"embed"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"testing/fstest"
)

//go:embed fixture
var fixture embed.FS

func TestReadDir(t *testing.T) {
	_, err := ReadDir(fixture, "fixture")
	assert.NoError(t, err)
}

func TestParseFileNames(t *testing.T) {
	fs := fstest.MapFS{
		"contains_dir/dir":                     {Mode: os.ModeDir},
		"invalid_file_name/0.description.up":   {},
		"invalid_file_name/foo.description.up": {},
		"valid/0.description.up":               {},
		"valid/1.description.down.sql":         {},
	}
	t.Run("empty dir", func(t *testing.T) {
		res, err := parseFileNames([]os.DirEntry{})
		assert.NoError(t, err)
		assert.Empty(t, res)
	})
	t.Run("contains directory", func(t *testing.T) {
		dirEntries, err := fs.ReadDir("contains_dir")
		assert.NoError(t, err)
		_, err = parseFileNames(dirEntries)
		assert.EqualError(t, err, errContainsDirectory.Error())
	})
	t.Run("invalid file name", func(t *testing.T) {
		dirEntries, err := fs.ReadDir("invalid_file_name")
		assert.NoError(t, err)
		_, err = parseFileNames(dirEntries)
		assert.EqualError(t, err, errFileNameIdx.Error())
	})
	t.Run("sorts valid file names", func(t *testing.T) {
		dirEntries, err := fs.ReadDir("valid")
		assert.NoError(t, err)
		// if read dirEntries are sorted - mix them up
		if string(dirEntries[0].Name()[0]) == "0" {
			dirEntries[0], dirEntries[1] = dirEntries[1], dirEntries[0]
		}
		assert.Equal(t, "1.description.down.sql", dirEntries[0].Name())
		assert.Equal(t, "0.description.up", dirEntries[1].Name())
		res, err := parseFileNames(dirEntries)
		assert.NoError(t, err)
		assert.Equal(t, "0.description.up", res[0].fileName)
		assert.Equal(t, "1.description.down.sql", res[1].fileName)
	})
}

func TestParseFileName(t *testing.T) {
	t.Run("valid file name", func(t *testing.T) {
		res, err := parseFileName("0.description.up")
		assert.NoError(t, err)
		assert.Equal(t, &parsedFileName{fileName: "0.description.up", idx: 0, description: "description", direction: up}, res)
	})
	t.Run("less than three parts", func(t *testing.T) {
		res, err := parseFileName("0.description")
		assert.Nil(t, res)
		assert.EqualError(t, err, errFileNameParts.Error())
	})
	t.Run("index not convertable to int", func(t *testing.T) {
		res, err := parseFileName("a.description.up")
		assert.Nil(t, res)
		assert.EqualError(t, err, errFileNameIdx.Error())
	})
	t.Run("invalid direction", func(t *testing.T) {
		res, err := parseFileName("0.description.UP")
		assert.Nil(t, res)
		assert.EqualError(t, err, errFileNameDirection.Error())
	})
}
