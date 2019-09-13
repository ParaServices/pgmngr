package pgmngr

import (
	"bufio"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/icrowley/fake"
	"github.com/stretchr/testify/require"
)

func TestCreateMigration(t *testing.T) {
	t.Run("migration created", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "migrations_")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := testConfig(t)
		config.Migration.Directory = tempDir

		err = CreateMigration(config, "test", false)
		require.NoError(t, err)

		var files []string
		err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if filepath.Ext(path) == ".sql" {
				files = append(files, path)
			}
			return nil
		})
		require.NoError(t, err)
		// there should only be two files crated
		require.Equal(t, 2, len(files))

		placeholders := []string{
			string(downPlaceHolder),
			string(upPlaceHolder),
		}

		for i := range placeholders {
			file, err := os.Open(files[i])
			require.NoError(t, err)
			defer file.Close()

			reader := bufio.NewReader(file)

			var line string
			for {
				line, err = reader.ReadString('\n')
				require.Equal(t, string(placeholders[i]), line)

				if err != nil {
					break
				}
			}
		}
	})
}

var stmnttableExists = `
SELECT EXISTS (
   SELECT 1
   FROM   information_schema.tables
   WHERE  table_schema = $1
   AND    table_name = $2
   );
`

func TestApplyMigration(t *testing.T) {
	cfg := testConfig(t)
	err := CreateDatabase(*cfg)
	require.NoError(t, err)
	defer func(t *testing.T) {
		err = DropDatabase(*cfg)
		require.NoError(t, err)
	}(t)

	count := 10
	migrationFiles := make([]string, count)
	for i := 0; i < count; i++ {
		f := fake.Word() + fake.Word() + fake.Word()
		err = CreateMigration(cfg, f, false)
		require.NoError(t, err)
		migrationFiles[i] = f
	}

	// count*2 because we both generate an up and down
	migrationData := make(map[string]map[string]string)

	// now that we've created the migration, writes create some migrations
	// for it
	err = filepath.Walk(cfg.Migration.Directory, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".sql" {
			f, err := os.Open(path)
			require.NoError(t, err)
			defer f.Close()

			migrationVersion, err := getVersionFromFileName(filepath.Base(path))
			require.NoError(t, err)

			m, ok := migrationData[migrationVersion]
			if !ok {
				m = make(map[string]string)
			}

			// determine if it's an up or down
			if isDownMigrationRegex.Match([]byte(path)) {
				m["down_filepath"] = path
			}

			if isUpMigrationRegex.Match([]byte(path)) {
				m["up_filepath"] = path
			}
			migrationData[migrationVersion] = m
		}
		return nil
	})
	require.NoError(t, err)

	tables := make([]string, 0)
	schemaName := "public"

	for _, vMap := range migrationData {
		table := strings.ToLower(fake.Word()) + "_" + strings.ToLower(fake.Word())
		func(table, fPath string) {
			f, err := os.OpenFile(fPath, os.O_APPEND|os.O_RDWR, 0644)
			require.NoError(t, err)
			defer f.Close()

			_, err = f.Write([]byte("\nCREATE TABLE " + schemaName + "." + table + "();"))
			require.NoError(t, err)
		}(table, vMap["up_filepath"])

		func(table, fPath string) {
			f, err := os.OpenFile(fPath, os.O_APPEND|os.O_RDWR, 0644)
			require.NoError(t, err)
			defer f.Close()

			_, err = f.Write([]byte("\nCREATE TABLE " + schemaName + "." + table + "();"))
			require.NoError(t, err)
		}(table, vMap["down_filepath"])

		tables = append(tables, table)
	}

	err = ApplyMigration(Forward, cfg)
	require.NoError(t, err)

	for i := range tables {
		cfg.Connection.PingIntervals = 300
		exists, err := tableExists(schemaName, tables[i], cfg)
		require.NoError(t, err)
		require.True(t, exists)
	}

	// check if tables are created
}

func tableExists(schemaName, tableName string, cfg *Config) (bool, error) {
	dbURL, err := cfg.dbURL()
	if err != nil {
		return false, NewError(err)
	}

	db, err := sql.Open(pgDriver, dbURL)
	if err != nil {
		return false, NewError(err)
	}
	defer db.Close()

	err = pingDatabase(db, *cfg)
	if err != nil {
		return false, NewError(err)
	}

	stmnt, err := db.Prepare(stmnttableExists)
	if err != nil {
		return false, NewError(err)
	}

	row := stmnt.QueryRow(
		schemaName,
		tableName,
	)

	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, NewError(
				fmt.Errorf("database: %s already exists", cfg.Connection.Database),
			)
		}
		return false, NewError(err)
	}

	return exists, nil
}
