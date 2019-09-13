package pgmngr

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/color"
)

func generateMigrationVersion(c *Config) string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

var upPlaceHolder = []byte(`-- SQL statement for migration goes here.`)
var downPlaceHolder = []byte(`-- SQL statement for reversing/reverting the migration.\naslkdjfasklfjsaklfjsalkfjslk`)

type migrationType int

const (
	// Forward execute up migrations
	Forward migrationType = iota
	// Rollback execute down migratiopns
	Rollback
)

// CreateMigration generates new, empty migration files.
func CreateMigration(c *Config, name string, noTransaction bool) error {
	version := generateMigrationVersion(c)
	prefix := fmt.Sprint(version, "_", name)

	if noTransaction {
		prefix += ".no_txn"
	}

	upFilepath := filepath.Join(c.Migration.Directory, prefix+".up.sql")
	downFilepath := filepath.Join(c.Migration.Directory, prefix+".down.sql")

	err := ioutil.WriteFile(upFilepath, upPlaceHolder, 0644)
	if err != nil {
		return err
	}
	color.Info.Tips("Created migration file: %s", colorBlue(upFilepath))

	err = ioutil.WriteFile(downFilepath, downPlaceHolder, 0644)
	if err != nil {
		return err
	}
	color.Info.Tips("Created migration file: %s", colorBlue(upFilepath))

	return nil
}

var colorBlue = color.FgBlue.Render

func wrapInTransaction(file string) bool {
	return !strings.Contains(file, ".no_txn.")
}

type execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// ApplyMigration ...
func ApplyMigration(mType migrationType, cfg *Config) error {
	// check if migration table exists
	exists, err := schemaMigrationsTableExists(cfg)
	if err != nil {
		return NewError(err)
	}
	if !exists {
		err = createTableSchemaMigration(cfg)
		if err != nil {
			return NewError(err)
		}
	}

	mFiles, err := getUnAppliedMigrationFiles(mType, cfg)
	if err != nil {
		return NewError(err)
	}
	mFilesKeysSorted := make([]int64, len(mFiles))
	i := 0
	for k := range mFiles {
		mFilesKeysSorted[i] = k
		i++
	}

	sort.Slice(
		mFilesKeysSorted,
		func(i, j int) bool {
			return mFilesKeysSorted[i] < mFilesKeysSorted[j]
		},
	)

	dbURL, err := cfg.dbURL()
	if err != nil {
		return NewError(err)
	}

	db, err := sql.Open(pgDriver, dbURL)
	if err != nil {
		return NewError(err)
	}
	defer db.Close()

	err = pingDatabase(db, cfg.Connection.PingIntervals)
	if err != nil {
		return NewError(err)
	}

	rollback := func(tx *sql.Tx) {
		if tx != nil {
			tx.Rollback()
		}
	}

	var exec execer
	exec = db
	for i := range mFilesKeysSorted {
		var tx *sql.Tx
		filePath := mFiles[mFilesKeysSorted[i]]
		wrapInTxn := wrapInTransaction(filePath)
		if wrapInTxn {
			tx, err = db.Begin()
			if err != nil {
				return NewError(err)
			}
			exec = tx
		}
		color.Note.Tips("Running migration for: %s", colorBlue(filePath))
		f, err := os.Open(filePath)
		if err != nil {
			return NewError(err)
		}
		defer f.Close()

		reader := bufio.NewReader(f)

		var line string
		var builder strings.Builder
		defer builder.Reset()
		for {
			line, err = reader.ReadString('\n')
			builder.WriteString(line)
			if err != nil {
				break
			}
			builder.WriteRune('\n')
		}
		if err != io.EOF {
			return NewError(err)
		}

		_, err = exec.Exec(builder.String())
		if err != nil {
			rollback(tx)
			return NewError(err)
		}

		if wrapInTxn {
			err = tx.Commit()
			if err != nil {
				rollback(tx)
				return NewError(err)
			}
		}
		color.Success.Tips("Migration successful using migration file: %s", colorBlue(filePath))
		continue
	}

	return nil
}

func schemaMigrationsTableExists(cfg *Config) (bool, error) {
	dbURL, err := cfg.dbURL()
	if err != nil {
		return false, NewError(err)
	}

	db, err := sql.Open(pgDriver, dbURL)
	if err != nil {
		return false, NewError(err)
	}
	defer db.Close()

	err = pingDatabase(db, cfg.Connection.PingIntervals)
	if err != nil {
		return false, NewError(err)
	}

	row := db.QueryRow(
		stmntSchemaMigrationTableExists,
		cfg.Migration.Table.Schema,
		cfg.Migration.Table.Name,
	)
	if err != nil {
		return false, NewError(err)
	}

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

func createTableSchemaMigration(cfg *Config) error {
	exists, err := schemaMigrationsTableExists(cfg)
	if err != nil {
		return NewError(err)
	}

	if !exists {
		dbURL, err := cfg.dbURL()
		if err != nil {
			return NewError(err)
		}

		db, err := sql.Open(pgDriver, dbURL)
		if err != nil {
			return NewError(err)
		}
		defer db.Close()

		err = pingDatabase(db, cfg.Connection.PingIntervals)
		if err != nil {
			return NewError(err)
		}

		_, err = db.Exec(stmntCreateSchemaMigrationsTableFn)
		if err != nil {
			return NewError(err)
		}

		stmnt, err := db.Prepare(stmntCreateSchemaMigrationsTable)
		if err != nil {
			return NewError(err)
		}
		defer stmnt.Close()

		_, err = stmnt.Exec(
			cfg.Migration.Table.Schema,
			cfg.Migration.Table.Name,
		)
		if err != nil {
			return NewError(err)
		}
	}

	return nil
}

func getVersionFromFileName(fileName string) (string, error) {
	baseTokens := strings.Split(fileName, ".")
	subTokens := strings.Split(baseTokens[0], "_")

	// a bit redundant given that we are creating the migration numbers
	// as Unix we need to consider existing migrations (we might just
	// simply need to convert the timestamps to unix)
	i, err := strconv.ParseInt(subTokens[0], 10, 64)
	if err != nil {
		return "", NewError(err)
	}
	t, err := time.Parse(time.RFC3339, time.Unix(i, 0).Format(time.RFC3339))
	if err != nil {
		return "", NewError(err)
	}
	return strconv.FormatInt(t.Unix(), 10), nil
}

type migrationFiles map[int64]string

func (m migrationFiles) Versions() []int64 {
	versions := make([]int64, len(m))
	i := 0
	for k := range m {
		versions[i] = k
		i++
	}
	return versions
}

var isUpMigrationRegex = regexp.MustCompile(`^(.*\.up\.sql)$`)
var isDownMigrationRegex = regexp.MustCompile(`^(.*\.down\.sql)$`)

func getMigrationFiles(mType migrationType, cfg *Config) (migrationFiles, error) {
	mFiles := make(migrationFiles)
	err := filepath.Walk(cfg.Migration.Directory, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".sql" {
			versionStr, err := getVersionFromFileName(filepath.Base(path))
			if err != nil {
				return NewError(err)
			}
			if mType == Forward && isUpMigrationRegex.Match([]byte(path)) {
				versionInt64, err := strconv.ParseInt(versionStr, 10, 64)
				if err != nil {
					return NewError(err)
				}
				mFiles[versionInt64] = path
			}
			if mType == Rollback && isDownMigrationRegex.Match([]byte(path)) {
				versionInt64, err := strconv.ParseInt(versionStr, 10, 64)
				if err != nil {
					return NewError(err)
				}
				mFiles[versionInt64] = path
			}
		}
		return nil
	})
	if err != nil {
		return nil, NewError(err)
	}

	return mFiles, nil
}

func getUnAppliedMigrationFiles(mType migrationType, cfg *Config) (migrationFiles, error) {
	dbURL, err := cfg.dbURL()
	if err != nil {
		return nil, NewError(err)
	}

	db, err := sql.Open(pgDriver, dbURL)
	if err != nil {
		return nil, NewError(err)
	}
	defer db.Close()

	err = pingDatabase(db, cfg.Connection.PingIntervals)
	if err != nil {
		return nil, NewError(err)
	}

	mFiles, err := getMigrationFiles(mType, cfg)
	if err != nil {
		return nil, NewError(err)
	}

	appliedMigrations, err := getAllAppliedMigrations(cfg)
	if err != nil {
		return nil, NewError(err)
	}
	migrations, _ := sliceExclusionInt64s(mFiles.Versions(), appliedMigrations)

	unAppliedMigrations := make(migrationFiles)
	for i := range migrations {
		unAppliedMigrations[migrations[i]] = mFiles[migrations[i]]
	}

	return unAppliedMigrations, nil
}

func getAllAppliedMigrations(cfg *Config) ([]int64, error) {
	dbURL, err := cfg.dbURL()
	if err != nil {
		return nil, NewError(err)
	}

	db, err := sql.Open(pgDriver, dbURL)
	if err != nil {
		return nil, NewError(err)
	}
	defer db.Close()

	_, err = db.Exec(stmntAllSchemaMigrationsFn)
	if err != nil {
		return nil, NewError(err)
	}

	rows, err := db.Query(
		stmntAllSchemaMigrations,
		cfg.Migration.Table.Schema,
		cfg.Migration.Table.Name,
	)
	if err != nil {
		return nil, NewError(err)
	}

	appliedMigrations := make([]int64, 0)
	for rows.Next() {
		var migration int64
		err = rows.Scan(&migration)
		if err != nil {
			return nil, NewError(err)
		}
		appliedMigrations = append(appliedMigrations, migration)
	}

	return appliedMigrations, nil
}
