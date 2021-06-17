package dbmigrat

import (
	"crypto/sha1"
	"fmt"
	"github.com/jmoiron/sqlx"
)

func CheckLogTableIntegrity(db *sqlx.DB, migrations Migrations) (*IntegrityCheckResult, error) {
	var migrationLogs []migrationLog
	err := db.Select(&migrationLogs, `select * from dbmigrat_log`)

	if err != nil {
		return nil, err
	}

	result := EmptyIntegrityCheckResult()

	for _, migrationLog := range migrationLogs {
		repoMigrations, ok := migrations[migrationLog.Repo]
		if !ok {
			result.IsCorrupted = true
			result.RedundantRepos[migrationLog.Repo] = true
			continue
		}
		if migrationLog.Idx >= len(repoMigrations) {
			result.IsCorrupted = true
			result.RedundantMigrations[migrationLog.Repo] = append(result.RedundantMigrations[migrationLog.Repo], migrationLog)
			continue
		}

		if migrationLog.Checksum != fmt.Sprintf("%x", sha1.Sum([]byte(repoMigrations[migrationLog.Idx].Up))) {
			result.IsCorrupted = true
			result.InvalidChecksums[migrationLog.Repo] = append(result.RedundantMigrations[migrationLog.Repo], migrationLog)
		}
	}

	return result, nil
}

// IntegrityCheckResult
// RedundantRepos and RedundantMigrations represent objects which exist in DB log but not in passed migrations
type IntegrityCheckResult struct {
	IsCorrupted         bool
	RedundantRepos      map[Repo]bool
	RedundantMigrations map[Repo][]migrationLog
	InvalidChecksums    map[Repo][]migrationLog
}

func EmptyIntegrityCheckResult() *IntegrityCheckResult {
	return &IntegrityCheckResult{
		IsCorrupted:         false,
		RedundantRepos:      map[Repo]bool{},
		RedundantMigrations: map[Repo][]migrationLog{},
		InvalidChecksums:    map[Repo][]migrationLog{},
	}
}
