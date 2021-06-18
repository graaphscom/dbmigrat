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
	tx, _ := db.Beginx()

	t.Run("Empty migrations log returns serial -1, no errors", func(t *testing.T) {
		serial, err := fetchLastMigrationSerial(tx)
		assert.NoError(t, err)
		assert.Equal(t, int32(-1), serial)
	})

	t.Run("Migrations log with one migration returns serial 0, no errors", func(t *testing.T) {
		assert.NoError(t, insertLogs(db, []migrationLog{{
			Idx:             0,
			Repo:            "foo",
			MigrationSerial: 0,
			Checksum:        "",
			Description:     "",
		}}))
		serial, err := fetchLastMigrationSerial(tx)
		assert.NoError(t, err)
		assert.Equal(t, int32(0), serial)
	})

	t.Run("Migrations log with two migrations returns serial 1, no errors", func(t *testing.T) {
		assert.NoError(t, insertLogs(db, []migrationLog{{
			Idx:             1,
			Repo:            "foo",
			MigrationSerial: 1,
			Checksum:        "",
			Description:     "",
		}}))
		serial, err := fetchLastMigrationSerial(tx)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), serial)
	})
}
