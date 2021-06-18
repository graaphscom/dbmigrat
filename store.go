package dbmigrat

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"time"
)

func CreateLogTable(db *sqlx.DB) error {
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

func fetchAllMigrationLogs(db *sqlx.DB) ([]migrationLog, error) {
	var migrationLogs []migrationLog
	err := db.Select(&migrationLogs, `select * from dbmigrat_log`)
	return migrationLogs, err
}

func fetchLastMigrationSerial(tx *sqlx.Tx) (int, error) {
	row := tx.QueryRow(`select max(migration_serial) from dbmigrat_log`)
	var result sql.NullInt32
	err := row.Scan(&result)
	if err != nil {
		return -1, err
	}
	if !result.Valid {
		return -1, nil
	}
	return int(result.Int32), nil
}

func insertLogs(execer namedExecer, logs []migrationLog) error {
	_, err := execer.NamedExec(`
			insert into dbmigrat_log (idx, repo, migration_serial, checksum, applied_at, description)
			values (:idx, :repo, :migration_serial, :checksum, default, :description)
			`,
		logs,
	)

	return err
}

type namedExecer interface {
	NamedExec(query string, arg interface{}) (sql.Result, error)
}

func fetchLastMigrationIndexes(tx *sqlx.Tx) (map[Repo]int, error) {
	var dest []struct {
		Idx  int
		Repo Repo
	}
	err := tx.Select(&dest, `select max(idx) as idx, repo from dbmigrat_log group by repo`)
	if err != nil {
		return nil, err
	}

	repoToMaxIdx := map[Repo]int{}
	for _, res := range dest {
		repoToMaxIdx[res.Repo] = res.Idx
	}

	return repoToMaxIdx, nil
}

type migrationLog struct {
	Idx             int
	Repo            Repo
	MigrationSerial int `db:"migration_serial"`
	Checksum        string
	AppliedAt       time.Time `db:"applied_at"`
	Description     string
}
