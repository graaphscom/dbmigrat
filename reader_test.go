package dbmigrat

import (
	"embed"
	"errors"
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
		migrations, err := ReadDir(fixture, "fixture/auth")
		assert.NoError(t, err)
		assert.Equal(t, expected, migrations)
	})
	t.Run("properly reads current directory (path not relative to source file)", func(t *testing.T) {
		subFs, err := fs.Sub(fixture, "fixture/auth")
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
	t.Run("returns error when dir contains dir", func(t *testing.T) {
		fileSys := fstest.MapFS{
			"contains_dir/dir": {Mode: os.ModeDir},
		}
		migrations, err := ReadDir(fileSys, "contains_dir")
		assert.EqualError(t, err, errContainsDirectory.Error())
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
		assert.EqualError(t, err, errWithFileName{inner: errNotSequential, fileName: "2.description.up"}.Error())
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
		assert.EqualError(t, err, errWithFileName{inner: errSameDirections, fileName: "1.description.up.sql"}.Error())
		assert.Equal(t, []Migration(nil), migrations)
	})
	t.Run("returns error for invalid path", func(t *testing.T) {
		fileSys := fstest.MapFS{}
		migrations, err := ReadDir(fileSys, "non_existing_dir")
		assert.EqualError(t, err, "open non_existing_dir: file does not exist")
		assert.Equal(t, []Migration(nil), migrations)
	})
	t.Run("returns error for invalid file name", func(t *testing.T) {
		fileSys := fstest.MapFS{
			"0.description.foo": {},
		}
		migrations, err := ReadDir(fileSys, ".")
		assert.EqualError(t, err, errFileNameDirection.Error())
		assert.Equal(t, []Migration(nil), migrations)
	})
	t.Run("returns error when file open filed", func(t *testing.T) {
		fileSys := fstest.MapFS{
			"0.description.up":   {},
			"0.description.down": {},
		}
		migrations, err := ReadDir(fileSysMock{wrapped: fileSys, brokenFileName: "0.description.up"}, ".")
		assert.EqualError(t, err, errOpenFiled.Error())
		assert.Equal(t, []Migration(nil), migrations)

		migrations, err = ReadDir(fileSysMock{wrapped: fileSys, brokenFileName: "0.description.down"}, ".")
		assert.EqualError(t, err, errOpenFiled.Error())
		assert.Equal(t, []Migration(nil), migrations)
	})

}

func TestParseFileNames(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		res, err := parseFileNames([]string{})
		assert.NoError(t, err)
		assert.Empty(t, res)
	})
	t.Run("invalid file name", func(t *testing.T) {
		_, err := parseFileNames([]string{"0.description.up", "foo.description.up"})
		assert.EqualError(t, err, errFileNameIdx.Error())
	})
	t.Run("sorts valid file names", func(t *testing.T) {
		res, err := parseFileNames([]string{
			"1.description.up.sql",
			"0.description.up",
			"0.description.down",
			"1.description.down.sql",
		})
		assert.NoError(t, err)
		assert.Equal(t, "0.description.up", res[0].fileName)
		assert.Equal(t, "0.description.down", res[1].fileName)
		assert.Equal(t, "1.description.up.sql", res[2].fileName)
		assert.Equal(t, "1.description.down.sql", res[3].fileName)
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

type fileSysMock struct {
	wrapped        fs.FS
	brokenFileName string
}

func (f fileSysMock) Open(name string) (fs.File, error) {
	if name == f.brokenFileName {
		return nil, errOpenFiled
	}
	return f.wrapped.Open(name)
}

var errOpenFiled = errors.New("mocked file sys: file open failed")
