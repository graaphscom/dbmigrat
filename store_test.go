package dbmigrat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateLogTable(t *testing.T) {
	assert.NoError(t, th.resetDB())
	// # Create table when it not exists
	assert.NoError(t, th.pgStore.CreateLogTable())
	// # Try to create table when it exists
	assert.NoError(t, th.pgStore.CreateLogTable())
}

func TestFetchLastMigrationSerial(t *testing.T) {
	// # Create empty migrations log
	assert.NoError(t, th.resetDB())
	assert.NoError(t, th.pgStore.CreateLogTable())

	t.Run("Empty migrations log returns serial -1, no errors", func(t *testing.T) {
		serial, err := th.pgStore.fetchLastMigrationSerial()
		assert.NoError(t, err)
		assert.Equal(t, -1, serial)
	})

	t.Run("Migrations log with one migration returns serial 0, no errors", func(t *testing.T) {
		assert.NoError(t, th.pgStore.insertLogs([]migrationLog{{
			Idx:             0,
			Repo:            "foo",
			MigrationSerial: 0,
			Checksum:        "",
			Description:     "",
		}}))
		serial, err := th.pgStore.fetchLastMigrationSerial()
		assert.NoError(t, err)
		assert.Equal(t, 0, serial)
	})

	t.Run("Migrations log with two migrations returns serial 1, no errors", func(t *testing.T) {
		assert.NoError(t, th.pgStore.insertLogs([]migrationLog{{
			Idx:             1,
			Repo:            "foo",
			MigrationSerial: 1,
			Checksum:        "",
			Description:     "",
		}}))
		serial, err := th.pgStore.fetchLastMigrationSerial()
		assert.NoError(t, err)
		assert.Equal(t, 1, serial)
	})
}
func TestIndexesFetch(t *testing.T) {
	complexMigrationLog := []migrationLog{
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
	t.Run("TestFetchLastMigrationIndexes", func(t *testing.T) {
		assert.NoError(t, th.resetDB())
		assert.NoError(t, th.pgStore.CreateLogTable())
		assert.NoError(t, th.pgStore.insertLogs(complexMigrationLog))

		res, err := th.pgStore.fetchLastMigrationIndexes()
		assert.NoError(t, err)
		assert.Equal(t, map[Repo]int{"foo": 2, "bar": 1}, res)
	})

	t.Run("TestFetchReverseMigrationIndexesAfterSerial", func(t *testing.T) {
		assert.NoError(t, th.resetDB())
		assert.NoError(t, th.pgStore.CreateLogTable())

		t.Run("Empty migrations log returns empty map, no error", func(t *testing.T) {
			res, err := th.pgStore.fetchReverseMigrationIndexesAfterSerial(-1)
			assert.NoError(t, err)
			assert.Equal(t, map[Repo][]int{}, res)
		})

		t.Run("Several repos, serials and migrations in log returns proper map, no error", func(t *testing.T) {
			assert.NoError(t, th.pgStore.insertLogs(complexMigrationLog))

			res, err := th.pgStore.fetchReverseMigrationIndexesAfterSerial(0)
			assert.NoError(t, err)
			assert.Equal(t, map[Repo][]int{
				"foo": {2, 1},
				"bar": {1},
			}, res)
		})
	})
}

func TestNoDbLog(t *testing.T) {
	assert.NoError(t, th.resetDB())
	expectedErr := `pq: relation "dbmigrat_log" does not exist`

	t.Run("fetchReverseMigrationIndexesAfterSerial", func(t *testing.T) {
		_, err := th.pgStore.fetchReverseMigrationIndexesAfterSerial(-100)
		assert.EqualError(t, err, expectedErr)
	})

	t.Run("deleteLogs", func(t *testing.T) {
		assert.EqualError(t, th.pgStore.deleteLogs([]migrationLog{{Idx: 0, Repo: "bar"}}), expectedErr)
	})

	t.Run("fetchLastMigrationIndexes", func(t *testing.T) {
		_, err := th.pgStore.fetchLastMigrationIndexes()
		assert.EqualError(t, err, expectedErr)
	})

	t.Run("fetchLastMigrationSerial", func(t *testing.T) {
		serial, err := th.pgStore.fetchLastMigrationSerial()
		assert.EqualError(t, err, expectedErr)
		assert.Equal(t, -1, serial)
	})
}

func TestDeleteLogs(t *testing.T) {
	assert.NoError(t, th.resetDB())
	assert.NoError(t, th.pgStore.CreateLogTable())

	assert.NoError(t, th.pgStore.insertLogs([]migrationLog{
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
	assert.NoError(t, th.pgStore.deleteLogs([]migrationLog{{Idx: 0, Repo: "bar"}}))
	var migrationLogs []migrationLog
	assert.NoError(t, th.db.Select(&migrationLogs, `select * from dbmigrat_log`))
	assert.Len(t, migrationLogs, 1)
	assert.Equal(t, 0, migrationLogs[0].Idx)
	assert.Equal(t, Repo("foo"), migrationLogs[0].Repo)
}
