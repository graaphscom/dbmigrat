package dbmigrat

import (
	"database/sql"
	"fmt"
)

func Migrate(db *sql.DB, m Migrations, RO RepoOrder) error {
	var err error

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	lastMigrationSerial, err := fetchLastMigrationSerial(tx)
	if err != nil {
		return err
	}
	migrationSerial := lastMigrationSerial + 1
	fmt.Println(migrationSerial)
	tx.Query(`select max(idx) as idx, repo from dbmigrat_log group by repo`)

	return nil
}

func fetchLastMigrationSerial(tx *sql.Tx) (int32, error) {
	row := tx.QueryRow(`select max(migration_serial) from dbmigrat_log`)
	var result sql.NullInt32
	err := row.Scan(&result)
	if err != nil {
		return 0, err
	}
	if !result.Valid {
		return -1, nil
	}
	return result.Int32, nil
}
