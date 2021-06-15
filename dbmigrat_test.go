package dbmigrat

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var db *sql.DB

func TestMain(m *testing.M) {
	_db, err := sql.Open("postgres", "postgres://dbmigrat:dbmigrat@localhost:5432/dbmigrat?sslmode=disable")
	db = _db
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(m.Run())
}

func TestCreateLogTable(t *testing.T) {
	resetDB()
	// # Create table when it not exists
	assert.NoError(t, CreateLogTable(db))
	// # Try to create table when it exists
	assert.NoError(t, CreateLogTable(db))
}

func TestFetchLastMigrationSerial(t *testing.T) {
	// # Create empty migrations log
	resetDB()
	assert.NoError(t, CreateLogTable(db))
	tx, _ := db.Begin()

	t.Run("Empty migrations log returns serial -1, no errors", func(t *testing.T) {
		serial, err := fetchLastMigrationSerial(tx)
		assert.NoError(t, err)
		assert.Equal(t, int32(-1), serial)
	})

	t.Run("Migrations log with one migration returns serial 0, no errors", func(t *testing.T) {
		_, err := tx.Exec(`insert into dbmigrat_log values (0, 'foo', 0, default)`)
		assert.NoError(t, err)
		serial, err := fetchLastMigrationSerial(tx)
		assert.NoError(t, err)
		assert.Equal(t, int32(0), serial)
	})

	t.Run("Migrations log with two migrations returns serial 1, no errors", func(t *testing.T) {
		_, err := tx.Exec(`insert into dbmigrat_log values (1, 'foo', 1, default)`)
		assert.NoError(t, err)
		serial, err := fetchLastMigrationSerial(tx)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), serial)
	})
}

func resetDB() {
	_, err := db.Exec(`drop schema if exists public cascade;create schema public`)
	if err != nil {
		log.Fatalln(err)
	}
}
