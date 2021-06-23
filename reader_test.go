package dbmigrat

import (
	"embed"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"
)

//go:embed fixture
var fixture embed.FS

func TestReadDir(t *testing.T) {
	expected := []Migration{
		{
			Description: "create_user_table",
			Up:          "create table users (id serial primary key);",
			Down:        "drop table users;",
		},
		{
			Description: "add_username_column",
			Up:          "alter table users add column username varchar(32);",
			Down:        "alter table users drop column username;",
		},
	}
	t.Run("properly reads subdirectory", func(t *testing.T) {
		migrations, err := ReadDir(fixture, "fixture/migrations")
		assert.NoError(t, err)
		assert.Equal(t, expected, migrations)
	})
	t.Run("properly reads current directory (path not relative to source file)", func(t *testing.T) {
		subFs, err := fs.Sub(fixture, "fixture/migrations")
		migrations, err := ReadDir(subFs, ".")
		assert.NoError(t, err)
		assert.Equal(t, expected, migrations)
	})
	t.Run("returns empty array when dir contains zero files", func(t *testing.T) {
		fileSys := fstest.MapFS{
			"no_files": {Mode: os.ModeDir},
		}
		migrations, err := ReadDir(fileSys, "no_files")
		assert.NoError(t, err)
		assert.Equal(t, []Migration(nil), migrations)
	})
	t.Run("returns empty array when dir contains one file", func(t *testing.T) {
		fileSys := fstest.MapFS{
			"one_file/0.description.up": {},
		}
		migrations, err := ReadDir(fileSys, "one_file")
		assert.NoError(t, err)
		assert.Equal(t, []Migration(nil), migrations)
	})
	t.Run("skips last migration with missing direction", func(t *testing.T) {
		fileSys := fstest.MapFS{
			"0.description.up":   {},
			"0.description.down": {},
			"1.description.down": {},
		}
		migrations, err := ReadDir(fileSys, ".")
		assert.NoError(t, err)
		assert.Len(t, migrations, 1)
	})
	t.Run("returns error when migration's direction file is missing", func(t *testing.T) {
		fileSys := fstest.MapFS{
			"0.description.up":   {},
			"0.description.down": {},
			"1.description.up":   {},
			"2.description.up":   {},
			"2.description.down": {},
		}
		migrations, err := ReadDir(fileSys, ".")
		assert.EqualError(t, err, errWithFileName{inner: errNotSequential, fileName: "1.description.up"}.Error())
		assert.Equal(t, []Migration(nil), migrations)
	})
	t.Run("returns error when files' indexes are not incrementing by one sequence", func(t *testing.T) {
		fileSys := fstest.MapFS{
			"0.description.up":   {},
			"0.description.down": {},
			"2.description.up":   {},
			"2.description.down": {},
		}
		migrations, err := ReadDir(fileSys, ".")
		assert.EqualError(t, err, errWithFileName{inner: errNotSequential, fileName: "2.description.down"}.Error())
		assert.Equal(t, []Migration(nil), migrations)
	})
	t.Run("returns error when files' descriptions are not equal", func(t *testing.T) {
		fileSys := fstest.MapFS{
			"0.description_equal.up":       {},
			"0.description_equal.down":     {},
			"1.description.up":             {},
			"1.description_not_equal.down": {},
		}
		migrations, err := ReadDir(fileSys, ".")
		assert.EqualError(t, err, errWithFileName{inner: errDescriptionNotEqual, fileName: "1.description.up"}.Error())
		assert.Equal(t, []Migration(nil), migrations)
	})
	t.Run("returns error when migration files have same direction", func(t *testing.T) {
		fileSys := fstest.MapFS{
			"0.description.up":     {},
			"0.description.down":   {},
			"1.description.up":     {},
			"1.description.up.sql": {},
			"2.description.up":     {},
			"2.description.down":   {},
		}
		migrations, err := ReadDir(fileSys, ".")
		assert.EqualError(t, err, errWithFileName{inner: errSameDirections, fileName: "1.description.up"}.Error())
		assert.Equal(t, []Migration(nil), migrations)
	})
}

func TestParseFileNames(t *testing.T) {
	fileSys := fstest.MapFS{
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
		dirEntries, err := fileSys.ReadDir("contains_dir")
		assert.NoError(t, err)
		_, err = parseFileNames(dirEntries)
		assert.EqualError(t, err, errContainsDirectory.Error())
	})
	t.Run("invalid file name", func(t *testing.T) {
		dirEntries, err := fileSys.ReadDir("invalid_file_name")
		assert.NoError(t, err)
		_, err = parseFileNames(dirEntries)
		assert.EqualError(t, err, errFileNameIdx.Error())
	})
	t.Run("sorts valid file names", func(t *testing.T) {
		dirEntries, err := fileSys.ReadDir("valid")
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
