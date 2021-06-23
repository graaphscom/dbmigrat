package dbmigrat

func CheckLogTableIntegrity(selector selector, migrations Migrations) (*IntegrityCheckResult, error) {
	migrationLogs, err := fetchAllMigrationLogs(selector)

	if err != nil {
		return nil, err
	}

	result := newIntegrityCheckResult()

	for _, log := range migrationLogs {
		repoMigrations, ok := migrations[log.Repo]
		if !ok {
			result.IsCorrupted = true
			result.RedundantRepos[log.Repo] = true
			continue
		}
		if log.Idx >= len(repoMigrations) {
			result.IsCorrupted = true
			result.RedundantMigrations[log.Repo] = append(result.RedundantMigrations[log.Repo], log)
			continue
		}

		if log.Checksum != sha1Checksum(repoMigrations[log.Idx].Up) {
			result.IsCorrupted = true
			result.InvalidChecksums[log.Repo] = append(result.RedundantMigrations[log.Repo], log)
		}
	}

	return result, nil
}

func newIntegrityCheckResult() *IntegrityCheckResult {
	return &IntegrityCheckResult{
		IsCorrupted:         false,
		RedundantRepos:      map[Repo]bool{},
		RedundantMigrations: map[Repo][]migrationLog{},
		InvalidChecksums:    map[Repo][]migrationLog{},
	}
}

// IntegrityCheckResult
// RedundantRepos and RedundantMigrations represent objects which exist in DB log but not in passed migrations
type IntegrityCheckResult struct {
	IsCorrupted         bool
	RedundantRepos      map[Repo]bool
	RedundantMigrations map[Repo][]migrationLog
	InvalidChecksums    map[Repo][]migrationLog
}
