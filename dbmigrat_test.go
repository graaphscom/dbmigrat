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

func TestMigrate(t *testing.T) {
	assert.NoError(t, resetDB())
	assert.NoError(t, CreateLogTable(db))

	assert.NoError(t, Migrate(
		db,
		Migrations{
			"auth": {
				{
					Up: `
						create table users
						(
							id       serial primary key,
							username varchar(255) not null
						)
						`,
					Description: "create user table",
				},
			},
			"billing": {
				{
					Up: `
						create table orders
						(
						    id          serial primary key,
						    user_id     integer references users (id) not null,
						    total_gross decimal(12, 2)                not null
						)
						`,
					Description: "create orders table",
				},
			},
		},
		RepoOrder{"auth", "billing"},
	))
}

func resetDB() error {
	_, err := db.Exec(`drop schema if exists public cascade;create schema public`)
	return err
}

func truncateLogTable() error {
	_, err := db.Exec(`truncate dbmigrat_log`)
	return err
}
