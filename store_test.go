package dbmigrat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateLogTable(t *testing.T) {
	assert.NoError(t, resetDB())
	// # Create table when it not exists
	assert.NoError(t, CreateLogTable(db))
	// # Try to create table when it exists
	assert.NoError(t, CreateLogTable(db))
}

func TestFetchLastMigrationSerial(t *testing.T) {
	// # Create empty migrations log
	assert.NoError(t, resetDB())
	assert.NoError(t, CreateLogTable(db))

	t.Run("Empty migrations log returns serial -1, no errors", func(t *testing.T) {
		serial, err := fetchLastMigrationSerial(db)
		assert.NoError(t, err)
		assert.Equal(t, -1, serial)
	})

	t.Run("Migrations log with one migration returns serial 0, no errors", func(t *testing.T) {
		assert.NoError(t, insertLogs(db, []migrationLog{{
			Idx:             0,
			Repo:            "foo",
			MigrationSerial: 0,
			Checksum:        "",
			Description:     "",
		}}))
		serial, err := fetchLastMigrationSerial(db)
		assert.NoError(t, err)
		assert.Equal(t, 0, serial)
	})

	t.Run("Migrations log with two migrations returns serial 1, no errors", func(t *testing.T) {
		assert.NoError(t, insertLogs(db, []migrationLog{{
			Idx:             1,
			Repo:            "foo",
			MigrationSerial: 1,
			Checksum:        "",
			Description:     "",
		}}))
		serial, err := fetchLastMigrationSerial(db)
		assert.NoError(t, err)
		assert.Equal(t, 1, serial)
	})
}

func TestFetchLastMigrationIndexes(t *testing.T) {
	assert.NoError(t, resetDB())
	assert.NoError(t, CreateLogTable(db))

	_, err := fetchLastMigrationIndexes(db)
	assert.NoError(t, err)
}

func TestFetchReverseMigrationIndexesAfterSerial(t *testing.T) {
	assert.NoError(t, resetDB())
	assert.NoError(t, CreateLogTable(db))

	t.Run("Empty migrations log returns empty map, no error", func(t *testing.T) {
		res, err := fetchReverseMigrationIndexesAfterSerial(db, -1)
		assert.NoError(t, err)
		assert.Equal(t, map[Repo][]int{}, res)
	})

	t.Run("Several repos, serials and migrations in log returns proper map, no error", func(t *testing.T) {
		assert.NoError(t, insertLogs(db, []migrationLog{
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
		}))

		res, err := fetchReverseMigrationIndexesAfterSerial(db, 0)
		assert.NoError(t, err)
		assert.Equal(t, map[Repo][]int{
			"foo": {2, 1},
			"bar": {1},
		}, res)
	})
}

func TestNoDbLog(t *testing.T) {
	assert.NoError(t, resetDB())
	expectedErr := `pq: relation "dbmigrat_log" does not exist`

	t.Run("fetchReverseMigrationIndexesAfterSerial", func(t *testing.T) {
		_, err := fetchReverseMigrationIndexesAfterSerial(db, -100)
		assert.EqualError(t, err, expectedErr)
	})

	t.Run("deleteLogs", func(t *testing.T) {
		tx, _ := db.Beginx()
		assert.EqualError(t, deleteLogs(tx, []migrationLog{{Idx: 0, Repo: "bar"}}), expectedErr)
	})

	t.Run("fetchLastMigrationIndexes", func(t *testing.T) {
		_, err := fetchLastMigrationIndexes(db)
		assert.EqualError(t, err, expectedErr)
	})

	t.Run("fetchLastMigrationSerial", func(t *testing.T) {
		serial, err := fetchLastMigrationSerial(db)
		assert.EqualError(t, err, expectedErr)
		assert.Equal(t, -1, serial)
	})
}

func TestDeleteLogs(t *testing.T) {
	assert.NoError(t, resetDB())
	assert.NoError(t, CreateLogTable(db))
	tx, _ := db.Beginx()
	defer tx.Commit()

	assert.NoError(t, insertLogs(tx, []migrationLog{
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
	assert.NoError(t, deleteLogs(tx, []migrationLog{{Idx: 0, Repo: "bar"}}))
	var migrationLogs []migrationLog
	assert.NoError(t, tx.Select(&migrationLogs, `select * from dbmigrat_log`))
	assert.Len(t, migrationLogs, 1)
	assert.Equal(t, 0, migrationLogs[0].Idx)
	assert.Equal(t, Repo("foo"), migrationLogs[0].Repo)
}
