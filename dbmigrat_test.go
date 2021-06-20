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

	assert.NoError(t, Migrate(db, migrations1, RepoOrder{"auth", "billing"}))
	assert.NoError(t, Migrate(db, migrations2, RepoOrder{"auth", "billing", "delivery"}))

	// # Check if replying migrate not runs already applied migrations
	assert.NoError(t, Migrate(db, migrations2, RepoOrder{"auth", "billing", "delivery"}))
}

func TestRollback(t *testing.T) {
	assert.NoError(t, resetDB())
	assert.NoError(t, CreateLogTable(db))
	assert.NoError(t, Migrate(db, migrations1, RepoOrder{"auth", "billing", "delivery"}))
	assert.NoError(t, Migrate(db, migrations2, RepoOrder{"auth", "billing", "delivery"}))

	assert.NoError(t, Rollback(db, migrations2, RepoOrder{"delivery", "billing", "auth"}, 0))
}

var migrations1 = Migrations{
	"auth": {
		{Up: `create table users (id serial primary key)`, Down: `drop table users`, Description: "create user table"},
		{Up: `alter table users add column username varchar(32)`, Down: `alter table users drop column username`, Description: "add username column"},
	},
	"billing": {
		{Up: `create table orders (id serial primary key, user_id integer references users (id) not null)`, Down: `drop table orders`, Description: `create orders table`},
	},
}

var migrations2 = Migrations{
	"auth": migrations1["auth"],
	"billing": append(migrations1["billing"],
		Migration{Up: `alter table orders add column value_gross decimal(12,2)`, Down: `alter table orders drop column value_gross`, Description: "add value gross column"},
	),
	"delivery": {
		{Up: `create table delivery_status (status integer, order_id integer references orders(id) primary key)`, Down: `drop table delivery_status`, Description: `create delivery status table`},
	},
}

func resetDB() error {
	_, err := db.Exec(`drop schema if exists public cascade;create schema public`)
	return err
}

func truncateLogTable() error {
	_, err := db.Exec(`truncate dbmigrat_log`)
	return err
}
