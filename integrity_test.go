package dbmigrat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckLogTableIntegrity(t *testing.T) {
	assert.NoError(t, th.resetDB())
	assert.NoError(t, th.pgStore.CreateLogTable())

	truncateLogTable := func() error {
		_, err := th.db.Exec(`truncate dbmigrat_log`)
		return err
	}

	t.Run("Empty migrations log is not corrupted", func(t *testing.T) {
		assert.NoError(t, truncateLogTable())

		// # Check for no migrations passed from outside
		result, err := CheckLogTableIntegrity(th.pgStore, Migrations{})
		assert.NoError(t, err)
		assert.Equal(t, newIntegrityCheckResult(), result)

		// # Check for several migrations passed from outside
		result, err = CheckLogTableIntegrity(th.pgStore, Migrations{
			"repo1": {},
			"repo2": {{
				Description: "example migration",
				Up:          "create table foo (id integer primary key)",
			}},
		})
		assert.NoError(t, err)
		assert.Equal(t, newIntegrityCheckResult(), result)
	})

	t.Run("Not corrupted log with one migration and extra migrations passed from outside", func(t *testing.T) {
		assert.NoError(t, truncateLogTable())
		upSql := "create table foo (id integer primary key)"
		assert.NoError(t, th.pgStore.insertLogs([]migrationLog{{
			Idx:             0,
			Repo:            "repo1",
			MigrationSerial: 0,
			Checksum:        sha1Checksum(upSql),
			Description:     "example migration",
		}}))

		result, err := CheckLogTableIntegrity(th.pgStore, Migrations{
			"repo1": {
				{Up: upSql},
				{Up: "example additional"},
			},
			"repo2": {},
			"repo3": {{Up: "example additional"}},
		})
		assert.NoError(t, err)
		assert.Equal(t, newIntegrityCheckResult(), result)
	})

	t.Run("Corrupted log", func(t *testing.T) {
		assert.NoError(t, truncateLogTable())
		invalidChecksum := migrationLog{
			Idx:             0,
			Repo:            "repo1",
			MigrationSerial: 0,
			Checksum:        "",
			Description:     "example migration invalid checksum",
		}
		redundantMigration := migrationLog{
			Idx:             1,
			Repo:            "repo1",
			MigrationSerial: 0,
			Checksum:        sha1Checksum("example"),
			Description:     "example redundant migration",
		}
		redundantRepo := migrationLog{
			Idx:             0,
			Repo:            "repoRedundant",
			MigrationSerial: 0,
			Checksum:        sha1Checksum("example"),
			Description:     "example migration redundant repo",
		}
		assert.NoError(t, th.pgStore.insertLogs([]migrationLog{invalidChecksum, redundantMigration, redundantRepo}))

		result, err := CheckLogTableIntegrity(th.pgStore, Migrations{
			"repo1": {
				{Up: "sql other than stored in log"},
			},
		})
		assert.NoError(t, err)
		// Set AppliedAt to be the same as inserted one
		redundantMigration.AppliedAt = result.RedundantMigrations["repo1"][0].AppliedAt
		invalidChecksum.AppliedAt = result.InvalidChecksums["repo1"][0].AppliedAt
		assert.Equal(t, &IntegrityCheckResult{
			IsCorrupted:         true,
			RedundantRepos:      map[Repo]bool{"repoRedundant": true},
			RedundantMigrations: map[Repo][]migrationLog{"repo1": {redundantMigration}},
			InvalidChecksums:    map[Repo][]migrationLog{"repo1": {invalidChecksum}},
		}, result)
	})

	t.Run("db error", func(t *testing.T) {
		storeMock := errorStoreMock{wrapped: th.pgStore, errFetchAllMigrationLogs: true}
		res, err := CheckLogTableIntegrity(storeMock, Migrations{})
		assert.EqualError(t, err, exampleErr.Error())
		assert.Nil(t, res)
	})
}
