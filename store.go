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

func fetchAllMigrationLogs(selector selector) ([]migrationLog, error) {
	var migrationLogs []migrationLog
	err := selector.Select(&migrationLogs, `select * from dbmigrat_log`)
	return migrationLogs, err
}

func fetchLastMigrationSerial(dbGetter dbGetter) (int, error) {
	var result sql.NullInt32
	err := dbGetter.Get(&result, `select max(migration_serial) from dbmigrat_log`)
	if err != nil {
		return -1, err
	}
	if !result.Valid {
		return -1, nil
	}
	return int(result.Int32), nil
}

type dbGetter interface {
	Get(dest interface{}, query string, args ...interface{}) error
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

func fetchLastMigrationIndexes(selector selector) (map[Repo]int, error) {
	var dest []struct {
		Idx  int
		Repo Repo
	}
	err := selector.Select(&dest, `select max(idx) as idx, repo from dbmigrat_log group by repo`)
	if err != nil {
		return nil, err
	}

	repoToMaxIdx := map[Repo]int{}
	for _, res := range dest {
		repoToMaxIdx[res.Repo] = res.Idx
	}

	return repoToMaxIdx, nil
}

func fetchReverseMigrationIndexesAfterSerial(selector selector, serial int) (map[Repo][]int, error) {
	var dest []struct {
		Idx  int
		Repo Repo
	}
	err := selector.Select(&dest, `select idx, repo from dbmigrat_log where migration_serial > $1 order by idx desc`, serial)
	if err != nil {
		return nil, err
	}

	repoToReverseMigrationIndexes := map[Repo][]int{}
	for _, res := range dest {
		repoToReverseMigrationIndexes[res.Repo] = append(repoToReverseMigrationIndexes[res.Repo], res.Idx)
	}

	return repoToReverseMigrationIndexes, nil
}

type selector interface {
	Select(dest interface{}, query string, args ...interface{}) error
}

func deleteLogs(tx *sqlx.Tx, logs []migrationLog) error {
	for _, log := range logs {
		_, err := tx.Exec(`delete from dbmigrat_log where idx = $1 and repo = $2`, log.Idx, log.Repo)
		if err != nil {
			return err
		}
	}
	return nil
}

type migrationLog struct {
	Idx             int
	Repo            Repo
	MigrationSerial int `db:"migration_serial"`
	Checksum        string
	AppliedAt       time.Time `db:"applied_at"`
	Description     string
}
