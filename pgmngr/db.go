package pgmngr

import (
	"database/sql"
	"fmt"
)

const pgDriver = "postgres"
const templateDB = "postgres"

func dbExists(cfg Config) (bool, error) {
	cfgTemplate := cfg
	cfgTemplate.Connection.Database = templateDB
	dbURL, err := cfgTemplate.dbURL()
	if err != nil {
		return false, NewError(err)
	}

	db, err := sql.Open(pgDriver, dbURL)
	if err != nil {
		return false, NewError(err)
	}
	defer db.Close()

	err = pingDatabase(db, cfg)
	if err != nil {
		return false, NewError(err)
	}

	stmnt, err := db.Prepare(stmntDBexists)
	if err != nil {
		return false, NewError(err)
	}

	row := stmnt.QueryRow(cfg.Connection.Database)

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

// CreateDatabase creates a database using the database name from the connection information
func CreateDatabase(cfg Config) error {
	cfgTemplate := cfg
	cfgTemplate.Connection.Database = templateDB
	dbURL, err := cfgTemplate.dbURL()
	if err != nil {
		return NewError(err)
	}

	db, err := sql.Open(pgDriver, dbURL)
	if err != nil {
		return NewError(err)
	}
	defer db.Close()

	err = pingDatabase(db, cfg)
	if err != nil {
		return NewError(err)
	}

	exists, err := dbExists(cfg)
	if err != nil {
		return NewError(err)
	}

	if exists {
		return NewError(
			fmt.Errorf(
				"database: %s already exists",
				cfg.Connection.Database),
		)
	}

	_, err = db.Exec(stmntCreateExtensionDBLink)
	if err != nil {
		return NewError(err)
	}
	defer db.Exec(stmntCreateExtensionDBLink)

	_, err = db.Exec(stmntCreateDatabaseFn)
	if err != nil {
		return NewError(err)
	}

	stmnt, err := db.Prepare(stmntCreateDatabase)
	if err != nil {
		return NewError(err)
	}
	defer stmnt.Close()

	_, err = stmnt.Exec(
		templateDB,
		cfg.Connection.Database,
		cfg.Connection.Username,
	)
	if err != nil {
		return NewError(err)
	}

	return nil
}

// DropDatabase ...
func DropDatabase(cfg Config) error {
	cfgTemplate := cfg
	cfgTemplate.Connection.Database = templateDB
	dbURL, err := cfgTemplate.dbURL()
	if err != nil {
		return NewError(err)
	}

	db, err := sql.Open(pgDriver, dbURL)
	if err != nil {
		return NewError(err)
	}
	defer db.Close()

	err = pingDatabase(db, cfg)
	if err != nil {
		return NewError(err)
	}

	exists, err := dbExists(cfg)
	if err != nil {
		return NewError(err)
	}

	if !exists {
		return NewError(
			fmt.Errorf(
				"database: %s does not exist",
				cfg.Connection.Database,
			),
		)
	}

	_, err = db.Exec(stmntCreateExtensionDBLink)
	if err != nil {
		return NewError(err)
	}
	defer db.Exec(stmntDropExtensionDBLink)

	_, err = db.Exec(stmntDropDatabaseFn)
	if err != nil {
		return NewError(err)
	}

	stmnt, err := db.Prepare(stmntDropDatabase)
	if err != nil {
		return NewError(err)
	}
	defer stmnt.Close()

	_, err = stmnt.Exec(cfg.Connection.Database)
	if err != nil {
		return NewError(err)
	}

	return nil
}
