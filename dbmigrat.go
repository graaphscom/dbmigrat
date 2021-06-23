// Package dbmigrat allows organizing database migrations across multiple locations
// (eg. across multiple repositories in monorepo project)
package dbmigrat

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/jmoiron/sqlx"
)

func Migrate(db *sqlx.DB, migrations Migrations, repoOrder RepoOrder) (int, error) {
	var err error

	tx, err := db.Beginx()
	if err != nil {
		return 0, err
	}

	logCount, err := migrate(tx, migrations, repoOrder)
	if err != nil {
		return 0, multierror.Append(err, tx.Rollback())
	}

	return logCount, tx.Commit()
}

func migrate(tx *sqlx.Tx, migrations Migrations, repoOrder RepoOrder) (int, error) {
	lastMigrationSerial, err := fetchLastMigrationSerial(tx)
	if err != nil {
		return 0, err
	}
	migrationSerial := lastMigrationSerial + 1

	lastMigrationIndexes, err := fetchLastMigrationIndexes(tx)
	if err != nil {
		return 0, err
	}

	var insertedLogsCount int
	for _, orderedRepo := range repoOrder {
		repoMigrations, ok := migrations[orderedRepo]
		if !ok {
			continue
		}
		lastMigrationIdx, ok := lastMigrationIndexes[orderedRepo]
		if !ok {
			lastMigrationIdx = -1
		}
		if len(repoMigrations) <= lastMigrationIdx+1 {
			continue
		}

		var logs []migrationLog
		for i, migrationToRun := range repoMigrations[lastMigrationIdx+1:] {
			_, err = tx.Exec(migrationToRun.Up)
			if err != nil {
				return 0, err
			}
			logs = append(logs, migrationLog{
				Idx:             lastMigrationIdx + 1 + i,
				Repo:            orderedRepo,
				MigrationSerial: migrationSerial,
				Checksum:        sha1Checksum(migrationToRun.Up),
				Description:     migrationToRun.Description,
			})
		}
		err = insertLogs(tx, logs)
		if err != nil {
			return 0, err
		}
		insertedLogsCount += len(logs)
	}

	return insertedLogsCount, nil
}

func Rollback(db *sqlx.DB, migrations Migrations, repoOrder RepoOrder, toMigrationSerial int) (int, error) {
	tx, err := db.Beginx()
	if err != nil {
		return 0, err
	}
	deletedLogs, err := rollback(tx, migrations, repoOrder, toMigrationSerial)
	if err != nil {
		return 0, multierror.Append(err, tx.Rollback())
	}
	return deletedLogs, tx.Commit()
}

func rollback(tx *sqlx.Tx, migrations Migrations, repoOrder RepoOrder, toMigrationSerial int) (int, error) {
	repoToReverseIndexes, err := fetchReverseMigrationIndexesAfterSerial(tx, toMigrationSerial)
	if err != nil {
		return 0, err
	}
	var logsToDelete []migrationLog
	for _, orderedRepo := range repoOrder {
		reverseIndexes, ok := repoToReverseIndexes[orderedRepo]
		if !ok {
			continue
		}
		for _, migrationIdx := range reverseIndexes {
			if len(migrations[orderedRepo]) < migrationIdx {
				return 0, errors.New("migrations passed to Rollback func are not in sync with migrations log. You might want to run CheckLogTableIntegrity func")
			}
			_, err := tx.Exec(migrations[orderedRepo][migrationIdx].Down)
			if err != nil {
				return 0, err
			}
			logsToDelete = append(logsToDelete, migrationLog{Idx: migrationIdx, Repo: orderedRepo})
		}
	}
	err = deleteLogs(tx, logsToDelete)
	if err != nil {
		return 0, err
	}

	return len(logsToDelete), nil
}

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

func sha1Checksum(data string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(data)))
}
