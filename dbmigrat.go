// Package dbmigrat allows for organizing database migrations across multiple locations
// (eg. across multiple repositories in monorepo project)
package dbmigrat

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/hashicorp/go-multierror"
)

// Migrate applies migrations to the store in given repoOrder.
//
// store is implemented by PostgresStore
//
// migrations parameter is a map where keys are repositories names (string),
// values are arrays of properly ordered Migration.
//
// repoOrder parameter is an array of repositories names (string). It
// determines order in which values from migrations map will be applied.
// eg. if migrations in repo "A" have foreign keys to repo "B" - then repoOrder should be {"B", "A"}
func Migrate(s store, migrations Migrations, repoOrder RepoOrder) (int, error) {
	err := s.begin()
	if err != nil {
		return 0, err
	}

	logCount, err := migrate(s, migrations, repoOrder)
	if err != nil {
		return 0, multierror.Append(err, s.rollback())
	}

	return logCount, s.commit()
}

func migrate(s store, migrations Migrations, repoOrder RepoOrder) (int, error) {
	lastMigrationSerial, err := s.fetchLastMigrationSerial()
	if err != nil {
		return 0, err
	}
	migrationSerial := lastMigrationSerial + 1

	lastMigrationIndexes, err := s.fetchLastMigrationIndexes()
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
			err = s.exec(migrationToRun.Up)
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
		err = s.insertLogs(logs)
		if err != nil {
			return 0, err
		}
		insertedLogsCount += len(logs)
	}

	return insertedLogsCount, nil
}

// Rollback rolls back migrations applied by Migrate func
//
// repoOrder should be reversed one passed to Migrate func
//
// migration serial represents applied migrations (from different repos) in single run of Migrate func.
// When toMigrationSerial == -1, then all applied migrations will be rolled back.
func Rollback(s store, migrations Migrations, repoOrder RepoOrder, toMigrationSerial int) (int, error) {
	err := s.begin()
	if err != nil {
		return 0, err
	}
	deletedLogs, err := rollback(s, migrations, repoOrder, toMigrationSerial)
	if err != nil {
		return 0, multierror.Append(err, s.rollback())
	}
	return deletedLogs, s.commit()
}

func rollback(s store, migrations Migrations, repoOrder RepoOrder, toMigrationSerial int) (int, error) {
	repoToReverseIndexes, err := s.fetchReverseMigrationIndexesAfterSerial(toMigrationSerial)
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
			if len(migrations[orderedRepo]) <= migrationIdx {
				return 0, errMigrationsOutSync
			}
			err := s.exec(migrations[orderedRepo][migrationIdx].Down)
			if err != nil {
				return 0, err
			}
			logsToDelete = append(logsToDelete, migrationLog{Idx: migrationIdx, Repo: orderedRepo})
		}
	}
	err = s.deleteLogs(logsToDelete)
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
// while billing migrations in repo "billing".
type Repo string

func sha1Checksum(data string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(data)))
}

var errMigrationsOutSync = errors.New("migrations passed to Rollback func are not in sync with migrations log. You might want to run CheckLogTableIntegrity func")
