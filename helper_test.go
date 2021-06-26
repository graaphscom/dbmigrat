package dbmigrat

import (
	"github.com/jmoiron/sqlx"
)

func newTestHelper(db *sqlx.DB) *testHelper {
	migrations1 := Migrations{
		"auth": {
			{Up: `create table users (id serial primary key)`, Down: `drop table users`, Description: "create user table"},
			{Up: `alter table users add column username varchar(32)`, Down: `alter table users drop column username`, Description: "add username column"},
		},
		"billing": {
			{Up: `create table orders (id serial primary key, user_id integer references users (id) not null)`, Down: `drop table orders`, Description: `create orders table`},
		},
	}
	tH := testHelper{
		migrations1: migrations1,
		migrations2: Migrations{
			"auth": migrations1["auth"],
			"billing": append(migrations1["billing"],
				Migration{Up: `alter table orders add column value_gross decimal(12,2)`, Down: `alter table orders drop column value_gross`, Description: "add value gross column"},
			),
			"delivery": {
				{Up: `create table delivery_status (status integer, order_id integer references orders(id) primary key)`, Down: `drop table delivery_status`, Description: `create delivery status table`},
			},
		},
		pgStore: &PostgresStore{DB: db},
		db:      db,
	}
	return &tH
}

func (tH testHelper) resetDB() error {
	_, err := tH.db.Exec(`drop schema if exists public cascade;create schema public`)
	return err
}

type testHelper struct {
	migrations1 Migrations
	migrations2 Migrations
	pgStore     *PostgresStore
	db          *sqlx.DB
}
