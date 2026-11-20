package postgres

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/DATA-DOG/go-sqlmock"
	libLog "github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateUp_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockFS := fstest.MapFS{
		"000001_create_test.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE test_table (id INT);"),
		},
		"000001_create_test.down.sql": &fstest.MapFile{
			Data: []byte("DROP TABLE test_table;"),
		},
	}

	migrator := NewMigratorWithFS(db, mockFS, &libLog.NoneLogger{})

	// 1. Current Database & Schema
	mock.ExpectQuery("SELECT CURRENT_DATABASE\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("testdb"))

	mock.ExpectQuery("SELECT CURRENT_SCHEMA\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"current_schema"}).AddRow("public"))

	// 2. Initial driver lock
	mock.ExpectExec("SELECT pg_advisory_lock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 3. Check schema_migrations_receivables in information_schema.tables
	mock.ExpectQuery("SELECT COUNT\\(1\\) FROM information_schema\\.tables").
		WithArgs("public", "schema_migrations_receivables").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// 4. Initial driver unlock
	mock.ExpectExec("SELECT pg_advisory_unlock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 5. Migrate lock
	mock.ExpectExec("SELECT pg_advisory_lock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 6. Query current version (empty)
	mock.ExpectQuery("SELECT version, dirty FROM \"public\"\\.\"schema_migrations_receivables\" LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"version", "dirty"}))

	// 7. Set dirty true for version 1
	mock.ExpectBegin()
	mock.ExpectExec("TRUNCATE \"public\"\\.\"schema_migrations_receivables\"").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO \"public\"\\.\"schema_migrations_receivables\"").
		WithArgs(int64(1), true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 8. Execute migration sql
	mock.ExpectExec("CREATE TABLE test_table").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 9. Set dirty false for version 1
	mock.ExpectBegin()
	mock.ExpectExec("TRUNCATE \"public\"\\.\"schema_migrations_receivables\"").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO \"public\"\\.\"schema_migrations_receivables\"").
		WithArgs(int64(1), false).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 10. Unlock
	mock.ExpectExec("SELECT pg_advisory_unlock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = migrator.MigrateUp(context.Background())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMigrateUp_NoChange(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mockFS := fstest.MapFS{
		"000001_create_test.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE test_table (id INT);"),
		},
	}

	migrator := NewMigratorWithFS(db, mockFS, &libLog.NoneLogger{})

	mock.ExpectQuery("SELECT CURRENT_DATABASE\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("testdb"))

	mock.ExpectQuery("SELECT CURRENT_SCHEMA\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"current_schema"}).AddRow("public"))

	mock.ExpectExec("SELECT pg_advisory_lock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery("SELECT COUNT\\(1\\) FROM information_schema\\.tables").
		WithArgs("public", "schema_migrations_receivables").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectExec("SELECT pg_advisory_unlock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("SELECT pg_advisory_lock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Already at version 1, dirty = false
	mock.ExpectQuery("SELECT version, dirty FROM \"public\"\\.\"schema_migrations_receivables\" LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"version", "dirty"}).AddRow(int64(1), false))

	mock.ExpectExec("SELECT pg_advisory_unlock").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = migrator.MigrateUp(context.Background())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
