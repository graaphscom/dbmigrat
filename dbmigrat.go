package dbmigrat

import (
	"database/sql"
	"time"
)

type Migrations map[Repo][]Migration

type Migration struct {
	Description string
	Up          string
	Down        string
}

type RepoOrder []Repo

// Repo is set of migrations. It allows for storing migrations in several locations.
// Example:
// e-commerce app might store authentication related migrations in repo "auth"
// while warehouse migrations in repo "warehouse".
type Repo string

type migrationLog struct {
	Idx             int
	Repo            Repo
	MigrationSerial int `db:"migration_serial"`
	Checksum        string
	AppliedAt       time.Time `db:"applied_at"`
	Description     string
}

func CreateLogTable(db *sql.DB) error {
	_, err := db.Exec(`
		create table if not exists dbmigrat_log
		(
		    idx              integer      not null,
		    repo             varchar(255) not null,
		    migration_serial integer      not null,
		    checksum         bytea        not null,
		    applied_at       timestamp    not null default current_timestamp,
		    description      text         not null,
		    primary key (idx, repo)
		)
	`)

	return err
}
