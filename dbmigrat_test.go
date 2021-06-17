package dbmigrat

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var db *sqlx.DB

func TestMain(m *testing.M) {
	_db, err := sqlx.Open("postgres", "postgres://dbmigrat:dbmigrat@localhost:5432/dbmigrat?sslmode=disable")
	db = _db
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(m.Run())
}

func TestCreateLogTable(t *testing.T) {
	assert.NoError(t, resetDB())
	// # Create table when it not exists
	assert.NoError(t, CreateLogTable(db.DB))
	// # Try to create table when it exists
	assert.NoError(t, CreateLogTable(db.DB))
}

func resetDB() error {
	_, err := db.Exec(`drop schema if exists public cascade;create schema public`)
	return err
}

func truncateLogTable() error {
	_, err := db.Exec(`truncate dbmigrat_log`)
	return err
}
