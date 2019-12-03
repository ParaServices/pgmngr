package pgmngr

import (
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"strconv"
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

	log.Println("drop datbase called and cfg", cfg)
	cfgTemplate := cfg
	dbURL, err := cfgTemplate.dbAdminURL()
	if err != nil {
		return NewError(err)
	}
	defer func() {
		log.Println(err)
	}()

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

	// cfg.Connection.Admin.Host,
	// cfg.Connection.Admin.Port,
	// cfg.Connection.Admin.Database,
	// cfg.Connection.Admin.Username,
	// cfg.Connection.Admin.Password,
	// cfg.Connection.Migration.Database,

	// curWd, err := os.Getwd()
	// if err != nil {
	// 	return NewError(err)
	// }
	// sqlFilePath := filepath.Join(curWd, "../resources/postgres/sqlstatements/dropdatabase.sql")

	sqlFilePath := "sqlstatements/dropdatabase.sql"

	// "postgresql://$DB_USER:$DB_PWD@$DB_SERVER/$DB_NAME"
	// export PGPASSWORD=pgmngr&& psql -U pgmngr  -h localhost -d postgres -p 5432 -a -f "../resources/postgres/sqlstatements/dropdatabase.sql" -v  mg_database='pgmngr_test_debitis'

	// cmd := exec.Command("/bin/bash", "-c", "/usr/local/bin/psql", "-U", cfg.Connection.Admin.Username, "-h", cfg.Connection.Admin.Host, "-d", cfg.Connection.Admin.Database, "-p", strconv.Itoa(cfg.Connection.Admin.Port), "-a",
	// 	"-f", sqlFilePath, "-v ", fmt.Sprintf("%s=%s", "mg_database", cfg.Connection.Migration.Database))

	cmdToExecute := fmt.Sprintf("export PGPASSWORD=%s&& psql %s %s %s %s %s %s %s %s %s %s %s %s %s", cfg.Connection.Admin.Password, "-U", cfg.Connection.Admin.Username, "-h", cfg.Connection.Admin.Host, "-d", cfg.Connection.Admin.Database, "-p", strconv.Itoa(cfg.Connection.Admin.Port), "-a", "-f", sqlFilePath, "-v ", fmt.Sprintf("%s=%s", "mg_database", cfg.Connection.Migration.Database))

	// cmd := exec.Command("/bin/bash", "-c", "psql -U pgmngr -h localhost -d postgres -p 5432 -a -f /Users/chandra.kasiraju/workspace/go/src/github.com/ParaServices/pgmngr/resources/postgres/sqlstatements/dropdatabase.sql -v  mg_database=pgmngr_test_quia")
	cmd := exec.Command("/bin/bash", "-c", cmdToExecute)

	log.Println(cmd.String())
	// var out, stderr bytes.Buffer
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	out, err := cmd.CombinedOutput()
	// log.Println(cmd.Output())
	// log.Println(cmd.Run().Error())
	if err != nil {
		log.Printf("Error executing query. Command Output:  %v", err)
		return NewError(err)
	}
	log.Printf("Command Output: %+v\n", string(out))

	// psql -U "pgmngr" -W "pgmngr" -h "postgres"

	// _, err = db.Exec(stmntCreateExtensionDBLink)
	// if err != nil {
	// 	return NewError(err)
	// }

	// if cfg.ForceDropDB == true {
	// 	log.Println("force connection drop is enable. closing all connections now")
	// 	_, err = db.Exec(stmntDropConnectionsFn)
	// 	if err != nil {
	// 		return NewError(err)
	// 	}

	// 	stmnt, err := db.Prepare(stmntDropConnections)
	// 	if err != nil {
	// 		return NewError(err)
	// 	}
	// 	defer stmnt.Close()

	// 	_, err = stmnt.Exec(
	// 		cfg.Connection.Migration.Database,
	// 	)
	// 	if err != nil {
	// 		return NewError(err)
	// 	}
	// }

	// _, err = db.Exec(stmntDropDatabaseFn)
	// if err != nil {
	// 	return NewError(err)
	// }

	// stmnt, err := db.Prepare(stmntDropDatabase)
	// if err != nil {
	// 	return NewError(err)
	// }
	// defer stmnt.Close()

	// _, err = stmnt.Exec(
	// 	cfg.Connection.Admin.Host,
	// 	cfg.Connection.Admin.Port,
	// 	cfg.Connection.Admin.Database,
	// 	cfg.Connection.Admin.Username,
	// 	cfg.Connection.Admin.Password,
	// 	cfg.Connection.Migration.Database,
	// )
	// if err != nil {
	// 	return NewError(err)
	// }

	return nil
}
