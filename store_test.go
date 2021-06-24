package dbmigrat

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateLogTable(t *testing.T) {
	assert.NoError(t, resetDB())
	// # Create table when it not exists
	assert.NoError(t, pgStore.CreateLogTable())
	// # Try to create table when it exists
	assert.NoError(t, pgStore.CreateLogTable())
}

func TestFetchLastMigrationSerial(t *testing.T) {
	// # Create empty migrations log
	assert.NoError(t, resetDB())
	assert.NoError(t, pgStore.CreateLogTable())

	t.Run("Empty migrations log returns serial -1, no errors", func(t *testing.T) {
		serial, err := pgStore.fetchLastMigrationSerial()
		assert.NoError(t, err)
		assert.Equal(t, -1, serial)
	})

	t.Run("Migrations log with one migration returns serial 0, no errors", func(t *testing.T) {
		assert.NoError(t, pgStore.insertLogs([]migrationLog{{
			Idx:             0,
			Repo:            "foo",
			MigrationSerial: 0,
			Checksum:        "",
			Description:     "",
		}}))
		serial, err := pgStore.fetchLastMigrationSerial()
		assert.NoError(t, err)
		assert.Equal(t, 0, serial)
	})

	t.Run("Migrations log with two migrations returns serial 1, no errors", func(t *testing.T) {
		assert.NoError(t, pgStore.insertLogs([]migrationLog{{
			Idx:             1,
			Repo:            "foo",
			MigrationSerial: 1,
			Checksum:        "",
			Description:     "",
		}}))
		serial, err := pgStore.fetchLastMigrationSerial()
		assert.NoError(t, err)
		assert.Equal(t, 1, serial)
	})
}

func TestFetchLastMigrationIndexes(t *testing.T) {
	assert.NoError(t, resetDB())
	assert.NoError(t, pgStore.CreateLogTable())
	assert.NoError(t, pgStore.insertLogs(complexMigrationLog))

	res, err := pgStore.fetchLastMigrationIndexes()
	assert.NoError(t, err)
	assert.Equal(t, map[Repo]int{"foo": 2, "bar": 1}, res)
}

func TestFetchReverseMigrationIndexesAfterSerial(t *testing.T) {
	assert.NoError(t, resetDB())
	assert.NoError(t, pgStore.CreateLogTable())

	t.Run("Empty migrations log returns empty map, no error", func(t *testing.T) {
		res, err := pgStore.fetchReverseMigrationIndexesAfterSerial(-1)
		assert.NoError(t, err)
		assert.Equal(t, map[Repo][]int{}, res)
	})

	t.Run("Several repos, serials and migrations in log returns proper map, no error", func(t *testing.T) {
		assert.NoError(t, pgStore.insertLogs(complexMigrationLog))

		res, err := pgStore.fetchReverseMigrationIndexesAfterSerial(0)
		assert.NoError(t, err)
		assert.Equal(t, map[Repo][]int{
			"foo": {2, 1},
			"bar": {1},
		}, res)
	})
}

var complexMigrationLog = []migrationLog{
	{
		Idx:             0,
		Repo:            "foo",
		MigrationSerial: 0,
		Checksum:        "",
		Description:     "",
	},
	{
		Idx:             0,
		Repo:            "bar",
		MigrationSerial: 0,
		Checksum:        "",
		Description:     "",
	},
	{
		Idx:             1,
		Repo:            "foo",
		MigrationSerial: 1,
		Checksum:        "",
		Description:     "",
	},
	{
		Idx:             2,
		Repo:            "foo",
		MigrationSerial: 1,
		Checksum:        "",
		Description:     "",
	},
	{
		Idx:             1,
		Repo:            "bar",
		MigrationSerial: 1,
		Checksum:        "",
		Description:     "",
	},
}

func TestNoDbLog(t *testing.T) {
	assert.NoError(t, resetDB())
	expectedErr := `pq: relation "dbmigrat_log" does not exist`

	t.Run("fetchReverseMigrationIndexesAfterSerial", func(t *testing.T) {
		_, err := pgStore.fetchReverseMigrationIndexesAfterSerial(-100)
		assert.EqualError(t, err, expectedErr)
	})

	t.Run("deleteLogs", func(t *testing.T) {
		assert.EqualError(t, pgStore.deleteLogs([]migrationLog{{Idx: 0, Repo: "bar"}}), expectedErr)
	})

	t.Run("fetchLastMigrationIndexes", func(t *testing.T) {
		_, err := pgStore.fetchLastMigrationIndexes()
		assert.EqualError(t, err, expectedErr)
	})

	t.Run("fetchLastMigrationSerial", func(t *testing.T) {
		serial, err := pgStore.fetchLastMigrationSerial()
		assert.EqualError(t, err, expectedErr)
		assert.Equal(t, -1, serial)
	})
}

func TestDeleteLogs(t *testing.T) {
	assert.NoError(t, resetDB())
	assert.NoError(t, pgStore.CreateLogTable())

	assert.NoError(t, pgStore.insertLogs([]migrationLog{
		{
			Idx:             0,
			Repo:            "foo",
			MigrationSerial: 0,
			Checksum:        "",
			Description:     "",
		},
		{
			Idx:             0,
			Repo:            "bar",
			MigrationSerial: 0,
			Checksum:        "",
			Description:     "",
		},
	}))
	assert.NoError(t, pgStore.deleteLogs([]migrationLog{{Idx: 0, Repo: "bar"}}))
	var migrationLogs []migrationLog
	assert.NoError(t, db.Select(&migrationLogs, `select * from dbmigrat_log`))
	assert.Len(t, migrationLogs, 1)
	assert.Equal(t, 0, migrationLogs[0].Idx)
	assert.Equal(t, Repo("foo"), migrationLogs[0].Repo)
}

func (s errorStoreMock) CreateLogTable() error                            { return exampleErr }
func (s errorStoreMock) fetchAllMigrationLogs() ([]migrationLog, error)   { return nil, exampleErr }
func (s errorStoreMock) fetchLastMigrationSerial() (int, error)           { return 0, exampleErr }
func (s errorStoreMock) insertLogs(logs []migrationLog) error             { return exampleErr }
func (s errorStoreMock) fetchLastMigrationIndexes() (map[Repo]int, error) { return nil, exampleErr }
func (s errorStoreMock) fetchReverseMigrationIndexesAfterSerial(serial int) (map[Repo][]int, error) {
	return nil, exampleErr
}
func (s errorStoreMock) deleteLogs(logs []migrationLog) error { return exampleErr }
func (s errorStoreMock) begin() error                         { return exampleErr }
func (s errorStoreMock) rollback() error                      { return exampleErr }
func (s errorStoreMock) commit() error                        { return exampleErr }
func (s errorStoreMock) exec(query string) error              { return exampleErr }

var exampleErr = errors.New("example error")

type errorStoreMock struct{}
