package dbmigrat

import (
	"database/sql"
)

func Migrate(db *sql.DB, m Migrations, PO PackageOrder) error {
	var err error

	//tx, err := db.Begin()
	if err != nil {
		return err
	}

	//lastMigrationSerial, err := fetchLastMigrationSerial(tx)
	//if err != nil {
	//	return err
	//}

	return nil
}

func fetchLastMigrationSerial(tx *sql.Tx) (int, error) {
	row := tx.QueryRow(`select max(migration_serial) from dbmigrat_log`)
	var result int
	err := row.Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func Rollback(beforeSerial int) {

}

type Migrations map[Package][]Migration

type Migration struct {
	Package Package
	Up      string
	Down    string
}

type PackageOrder []Package

// Package is package of migrations. It allows for storing migrations in several locations.
// Example:
// e-commerce app might store authentication related migrations in module "auth"
// while warehouse migrations in module "warehouse".
type Package string

func CreateLogTable(db *sql.DB) error {
	_, err := db.Exec(`
		create table if not exists dbmigrat_log
		(
		    idx              integer      not null,
		    package          varchar(255) not null,
		    migration_serial integer      not null,
		    applied_at       timestamp    not null default current_timestamp,
		    primary key (idx, package)
		)
	`)

	return err
}
