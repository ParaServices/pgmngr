package pgmngr

import (
	"database/sql"
	"fmt"
)

const pgDriver = "postgres"

func dbExists(cfg Config) (bool, error) {
	cfgTemplate := cfg
	dbURL, err := cfgTemplate.dbAdminURL()
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

	row := stmnt.QueryRow(cfg.Connection.Migration.Database)

	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, NewError(
				fmt.Errorf("database: %s already exists", cfg.Connection.Migration.Database),
			)
		}
		return false, NewError(err)
	}

	return exists, nil
}

// CreateDatabase creates a database using the database name from the connection information
func CreateDatabase(cfg Config) error {
	cfgTemplate := cfg
	dbURL, err := cfgTemplate.dbAdminURL()
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
				cfg.Connection.Admin.Database),
		)
	}

	_, err = db.Exec(stmntCreateExtensionDBLink)
	if err != nil {
		return NewError(err)
	}

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
		cfg.Connection.Admin.Host,
		cfg.Connection.Admin.Port,
		cfg.Connection.Admin.Database,
		cfg.Connection.Admin.Username,
		cfg.Connection.Admin.Password,
		cfg.Connection.Migration.Database,
		cfg.Connection.Migration.Username,
	)
	if err != nil {
		return NewError(err)
	}

	return nil
}

// DropDatabase ...
func DropDatabase(cfg Config) error {
	cfgTemplate := cfg
	dbURL, err := cfgTemplate.dbAdminURL()
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
				cfg.Connection.Migration.Database,
			),
		)
	}

	_, err = db.Exec(stmntCreateExtensionDBLink)
	if err != nil {
		return NewError(err)
	}

	_, err = db.Exec(stmntDropDatabaseFn)
	if err != nil {
		return NewError(err)
	}

	stmnt, err := db.Prepare(stmntDropDatabase)
	if err != nil {
		return NewError(err)
	}
	defer stmnt.Close()

	_, err = stmnt.Exec(
		cfg.Connection.Admin.Host,
		cfg.Connection.Admin.Port,
		cfg.Connection.Admin.Database,
		cfg.Connection.Admin.Username,
		cfg.Connection.Admin.Password,
		cfg.Connection.Migration.Database,
	)
	if err != nil {
		return NewError(err)
	}

	return nil
}

func ResetDatabase(cfg Config) error {
	err := DropDatabase(cfg)
	if err != nil {
		return err
	}
	err = CreateDatabase(cfg)
	if err != nil {
		return err
	}
	err = ApplyMigration(Forward, &cfg)
	if err != nil {
		return err
	}
	return nil
}
