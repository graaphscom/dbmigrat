package dbmigrat

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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

func resetDB() error {
	_, err := db.Exec(`drop schema if exists public cascade;create schema public`)
	return err
}

func truncateLogTable() error {
	_, err := db.Exec(`truncate dbmigrat_log`)
	return err
}
