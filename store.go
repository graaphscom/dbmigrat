package dbmigrat

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"time"
)

// CreateLogTable creates table in db where applied migrations will be saved.
// This should be called before use of other functions from dbmigrat lib.
func (s PostgresStore) CreateLogTable() error {
	_, err := s.getDbAccessor().Exec(`
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

func (s PostgresStore) fetchAllMigrationLogs() ([]migrationLog, error) {
	var migrationLogs []migrationLog
	err := s.getDbAccessor().Select(&migrationLogs, `select * from dbmigrat_log`)
	return migrationLogs, err
}

func (s PostgresStore) fetchLastMigrationSerial() (int, error) {
	var result sql.NullInt32
	err := s.getDbAccessor().Get(&result, `select max(migration_serial) from dbmigrat_log`)
	if err != nil {
		return -1, err
	}
	if !result.Valid {
		return -1, nil
	}
	return int(result.Int32), nil
}

func (s PostgresStore) insertLogs(logs []migrationLog) error {
	_, err := s.getDbAccessor().NamedExec(`
			insert into dbmigrat_log (idx, repo, migration_serial, checksum, applied_at, description)
			values (:idx, :repo, :migration_serial, :checksum, default, :description)
			`,
		logs,
	)

	return err
}

func (s PostgresStore) fetchLastMigrationIndexes() (map[Repo]int, error) {
	var dest []struct {
		Idx  int
		Repo Repo
	}
	err := s.getDbAccessor().Select(&dest, `select max(idx) as idx, repo from dbmigrat_log group by repo`)
	if err != nil {
		return nil, err
	}

	repoToMaxIdx := map[Repo]int{}
	for _, res := range dest {
		repoToMaxIdx[res.Repo] = res.Idx
	}

	return repoToMaxIdx, nil
}

func (s PostgresStore) fetchReverseMigrationIndexesAfterSerial(serial int) (map[Repo][]int, error) {
	var dest []struct {
		Idx  int
		Repo Repo
	}
	err := s.getDbAccessor().Select(&dest, `select idx, repo from dbmigrat_log where migration_serial > $1 order by idx desc`, serial)
	if err != nil {
		return nil, err
	}

	repoToReverseMigrationIndexes := map[Repo][]int{}
	for _, res := range dest {
		repoToReverseMigrationIndexes[res.Repo] = append(repoToReverseMigrationIndexes[res.Repo], res.Idx)
	}

	return repoToReverseMigrationIndexes, nil
}

func (s PostgresStore) deleteLogs(logs []migrationLog) error {
	for _, log := range logs {
		_, err := s.getDbAccessor().Exec(`delete from dbmigrat_log where idx = $1 and repo = $2`, log.Idx, log.Repo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) begin() error {
	tx, err := s.DB.Beginx()
	s.tx = tx
	return err
}

func (s *PostgresStore) rollback() error {
	err := s.tx.Rollback()
	s.tx = nil
	return err
}

func (s *PostgresStore) commit() error {
	err := s.tx.Commit()
	s.tx = nil
	return err
}

func (s PostgresStore) exec(query string) error {
	_, err := s.getDbAccessor().Exec(query)
	return err
}

func (s PostgresStore) getDbAccessor() dbAccessor {
	if s.tx != nil {
		return s.tx
	}
	return s.DB
}

type dbAccessor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
}

type PostgresStore struct {
	DB *sqlx.DB
	tx *sqlx.Tx
}

type store interface {
	CreateLogTable() error
	fetchAllMigrationLogs() ([]migrationLog, error)
	fetchLastMigrationSerial() (int, error)
	insertLogs(logs []migrationLog) error
	fetchLastMigrationIndexes() (map[Repo]int, error)
	fetchReverseMigrationIndexesAfterSerial(serial int) (map[Repo][]int, error)
	deleteLogs(logs []migrationLog) error
	begin() error
	rollback() error
	commit() error
	exec(query string) error
}

type migrationLog struct {
	Idx             int
	Repo            Repo
	MigrationSerial int `db:"migration_serial"`
	Checksum        string
	AppliedAt       time.Time `db:"applied_at"`
	Description     string
}
