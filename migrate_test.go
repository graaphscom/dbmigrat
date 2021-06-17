package dbmigrat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFetchLastMigrationSerial(t *testing.T) {
	// # Create empty migrations log
	assert.NoError(t, resetDB())
	assert.NoError(t, CreateLogTable(db.DB))
	tx, _ := db.Begin()

	t.Run("Empty migrations log returns serial -1, no errors", func(t *testing.T) {
		serial, err := fetchLastMigrationSerial(tx)
		assert.NoError(t, err)
		assert.Equal(t, int32(-1), serial)
	})

	t.Run("Migrations log with one migration returns serial 0, no errors", func(t *testing.T) {
		_, err := tx.Exec(`insert into dbmigrat_log values (0, 'foo', 0, '', default, '')`)
		assert.NoError(t, err)
		serial, err := fetchLastMigrationSerial(tx)
		assert.NoError(t, err)
		assert.Equal(t, int32(0), serial)
	})

	t.Run("Migrations log with two migrations returns serial 1, no errors", func(t *testing.T) {
		_, err := tx.Exec(`insert into dbmigrat_log values (1, 'foo', 1, '', default, '')`)
		assert.NoError(t, err)
		serial, err := fetchLastMigrationSerial(tx)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), serial)
	})
}
