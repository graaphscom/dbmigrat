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

func resetDB() {
	_, err := db.Exec(`drop schema if exists public cascade;create schema public`)
	if err != nil {
		log.Fatalln(err)
	}
}
