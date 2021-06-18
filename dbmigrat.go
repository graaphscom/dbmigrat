package dbmigrat

import (
	"crypto/sha1"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/jmoiron/sqlx"
)

func Migrate(db *sqlx.DB, migrations Migrations, repoOrder RepoOrder) error {
	var err error

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	err = migrate(tx, migrations, repoOrder)
	if err != nil {
		return multierror.Append(err, tx.Rollback())
	}

	return tx.Commit()
}

func migrate(tx *sqlx.Tx, migrations Migrations, repoOrder RepoOrder) error {
	lastMigrationSerial, err := fetchLastMigrationSerial(tx)
	if err != nil {
		return err
	}
	migrationSerial := lastMigrationSerial + 1
	fmt.Println(migrationSerial)

	var dest []struct {
		idx  int
		repo string
	}
	err = tx.Select(&dest, `select max(idx) as idx, repo from dbmigrat_log group by repo`)
	if err != nil {
		return err
	}
	repoToMaxIdx := map[string]int{}
	for _, res := range dest {
		repoToMaxIdx[res.repo] = res.idx
	}

	for _, repo := range repoOrder {
		repoMigrations, ok := migrations[repo]
		if !ok {
			continue
		}
		maxIdx, ok := repoToMaxIdx[string(repo)]
		if !ok {
			maxIdx = -1
		}
		if len(repoMigrations) > maxIdx {
			continue
		}

		_, err = tx.Exec(repoMigrations[maxIdx+1].Up)
		if err != nil {
			return err
		}
	}

	return nil
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
