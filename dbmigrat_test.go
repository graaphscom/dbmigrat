package dbmigrat

import (
	"errors"
	"github.com/hashicorp/go-multierror"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var th *testHelper

func TestMain(m *testing.M) {
	db, err := sqlx.Open("postgres", "postgres://dbmigrat:dbmigrat@localhost:5432/dbmigrat?sslmode=disable")
	th = newTestHelper(db)
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(m.Run())
}

func TestMigrate(t *testing.T) {
	assert.NoError(t, th.resetDB())
	assert.NoError(t, th.pgStore.CreateLogTable())

	logCount, err := Migrate(th.pgStore, th.migrations1, RepoOrder{"auth", "billing"})
	assert.NoError(t, err)
	assert.Equal(t, 3, logCount)

	logCount, err = Migrate(th.pgStore, th.migrations2, RepoOrder{"auth", "billing", "delivery"})
	assert.NoError(t, err)
	assert.Equal(t, 2, logCount)

	// # Check if replying migrate not runs already applied migrations
	logCount, err = Migrate(th.pgStore, th.migrations2, RepoOrder{"auth", "billing", "delivery"})
	assert.NoError(t, err)
	assert.Equal(t, 0, logCount)
}

func TestMigrateError(t *testing.T) {
	assert.NoError(t, th.resetDB())
	assert.NoError(t, th.pgStore.CreateLogTable())
	caseTable := caseTable{
		{name: "tx begin fail", storeMock: errorStoreMock{wrapped: th.pgStore, errBegin: true}, errExpected: exampleErr},
		{name: "fetchLastMigrationSerial fail", storeMock: errorStoreMock{wrapped: th.pgStore, errFetchLastMigrationSerial: true}, errExpected: exampleMultiErr},
		{name: "fetchLastMigrationIndexes fail", storeMock: errorStoreMock{wrapped: th.pgStore, errFetchLastMigrationIndexes: true}, errExpected: exampleMultiErr},
		{name: "exec fail", storeMock: errorStoreMock{wrapped: th.pgStore, errExec: true}, errExpected: exampleMultiErr},
		{name: "insertLogs fail", storeMock: errorStoreMock{wrapped: th.pgStore, errInsertLogs: true}, errExpected: exampleMultiErr},
	}

	for _, testCase := range caseTable {
		t.Run(testCase.name, func(t *testing.T) {
			res, err := Migrate(testCase.storeMock, th.migrations1, RepoOrder{"auth", "billing"})
			assert.EqualError(t, err, testCase.errExpected.Error())
			assert.Equal(t, 0, res)
		})
	}
}

func TestRollback(t *testing.T) {
	before := func(t *testing.T) {
		assert.NoError(t, th.resetDB())
		assert.NoError(t, th.pgStore.CreateLogTable())
		_, err := Migrate(th.pgStore, th.migrations1, RepoOrder{"auth", "billing", "delivery"})
		assert.NoError(t, err)
		_, err = Migrate(th.pgStore, th.migrations2, RepoOrder{"auth", "billing", "delivery"})
		assert.NoError(t, err)
	}

	t.Run("rolled back properly", func(t *testing.T) {
		before(t)

		logCount, err := Rollback(th.pgStore, th.migrations2, RepoOrder{"delivery", "billing", "auth"}, 0)
		assert.NoError(t, err)
		assert.Equal(t, 2, logCount)
	})

	t.Run("error", func(t *testing.T) {
		before(t)

		caseTable := caseTable{
			{name: "tx begin fail", storeMock: errorStoreMock{wrapped: th.pgStore, errBegin: true}, errExpected: exampleErr},
			{name: "fetchReverseMigrationIndexesAfterSerial fail", storeMock: errorStoreMock{wrapped: th.pgStore, errFetchReverseMigrationIndexesAfterSerial: true}, errExpected: exampleMultiErr},
			{name: "exec fail", storeMock: errorStoreMock{wrapped: th.pgStore, errExec: true}, errExpected: exampleMultiErr},
			{name: "deleteLogs fail", storeMock: errorStoreMock{wrapped: th.pgStore, errDeleteLogs: true}, errExpected: exampleMultiErr},
		}

		for _, testCase := range caseTable {
			t.Run(testCase.name, func(t *testing.T) {
				logCount, err := Rollback(testCase.storeMock, th.migrations2, RepoOrder{"delivery", "billing", "auth"}, 0)
				assert.EqualError(t, err, testCase.errExpected.Error())
				assert.Equal(t, 0, logCount)
			})
		}

		t.Run("too less migrations provided", func(t *testing.T) {
			logCount, err := Rollback(th.pgStore, th.migrations1, RepoOrder{"delivery", "billing", "auth"}, 0)
			assert.EqualError(t, err, multierror.Append(errMigrationsOutSync).Error())
			assert.Equal(t, 0, logCount)
		})
	})
}

type caseTable []struct {
	name        string
	storeMock   store
	errExpected error
}

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

var (
	exampleErr      = errors.New("example error")
	exampleMultiErr = multierror.Append(errors.New("example error"))
)

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
