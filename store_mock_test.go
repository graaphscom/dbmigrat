package dbmigrat

import "errors"

func (s errorStoreMock) CreateLogTable() error {
	if s.errCreateLogTable {
		return exampleErr
	}
	return s.wrapped.CreateLogTable()
}
func (s errorStoreMock) fetchAllMigrationLogs() ([]migrationLog, error) {
	if s.errFetchAllMigrationLogs {
		return nil, exampleErr
	}
	return s.wrapped.fetchAllMigrationLogs()
}
func (s errorStoreMock) fetchLastMigrationSerial() (int, error) {
	if s.errFetchLastMigrationSerial {
		return 0, exampleErr
	}
	return s.wrapped.fetchLastMigrationSerial()
}
func (s errorStoreMock) insertLogs(logs []migrationLog) error {
	if s.errInsertLogs {
		return exampleErr
	}
	return s.wrapped.insertLogs(logs)
}
func (s errorStoreMock) fetchLastMigrationIndexes() (map[Repo]int, error) {
	if s.errFetchLastMigrationIndexes {
		return nil, exampleErr
	}
	return s.wrapped.fetchLastMigrationIndexes()
}
func (s errorStoreMock) fetchReverseMigrationIndexesAfterSerial(serial int) (map[Repo][]int, error) {
	if s.errFetchReverseMigrationIndexesAfterSerial {
		return nil, exampleErr
	}
	return s.wrapped.fetchReverseMigrationIndexesAfterSerial(serial)
}
func (s errorStoreMock) deleteLogs(logs []migrationLog) error {
	if s.errDeleteLogs {
		return exampleErr
	}
	return s.wrapped.deleteLogs(logs)
}
func (s errorStoreMock) begin() error {
	if s.errBegin {
		return exampleErr
	}
	return s.wrapped.begin()
}
func (s errorStoreMock) rollback() error {
	if s.errRollback {
		return exampleErr
	}
	return s.wrapped.rollback()
}
func (s errorStoreMock) commit() error {
	if s.errCommit {
		return exampleErr
	}
	return s.wrapped.commit()
}
func (s errorStoreMock) exec(query string) error {
	if s.errExec {
		return exampleErr
	}
	return s.wrapped.exec(query)
}

var exampleErr = errors.New("example error")

type errorStoreMock struct {
	wrapped                                    store
	errCreateLogTable                          bool
	errFetchAllMigrationLogs                   bool
	errFetchLastMigrationSerial                bool
	errInsertLogs                              bool
	errFetchLastMigrationIndexes               bool
	errFetchReverseMigrationIndexesAfterSerial bool
	errDeleteLogs                              bool
	errBegin                                   bool
	errRollback                                bool
	errCommit                                  bool
	errExec                                    bool
}
