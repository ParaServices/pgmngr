package pgmngr

var stmntAllSchemaMigrationsFn = `
CREATE FUNCTION pg_temp.get_all_schema_migrations(
    _schema_name TEXT,
    _table_name TEXT
) RETURNS TABLE (
    schema_migration_version INT8
) AS
$$
BEGIN
    RETURN QUERY
    EXECUTE format(
	'SELECT t.schema_migration_version
	 FROM %I.%I t
	', _schema_name, _table_name
    );
END;
$$
language plpgsql;
`

var stmntAllSchemaMigrations = `
SELECT * FROM pg_temp.get_all_schema_migrations(
    CAST(NULLIF($1, NULL) AS TEXT),
    CAST(NULLIF($2, NULL) AS TEXT)
);
`

var stmntSchemaMigrationTableExists = `
SELECT EXISTS (
  SELECT 1
  FROM information_schema.tables
  WHERE table_schema = CAST(NULLIF($1, NULL) AS TEXT)
  AND table_name = CAST(NULLIF($2, NULL) AS TEXT)
);
`

var stmntCreateSchemaMigrationsTableFn = `
CREATE FUNCTION pg_temp.create_schema_migrations_table(
    _schema TEXT,
    _database TEXT
) RETURNS INTEGER AS
$$
BEGIN
  IF EXISTS (SELECT 1
      FROM information_schema.tables
      WHERE table_schema = CAST(NULLIF($1, NULL) AS TEXT)
      AND table_name = CAST(NULLIF($2, NULL) AS TEXT)
    ) THEN
    RAISE NOTICE 'table: %s.%s already exists', _schema, _database;
  ELSE
    EXECUTE format('
       CREATE TABLE IF NOT EXISTS %I.%I (
         schema_migration_version INT8 NOT NULL,
         created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT (NOW() AT TIME ZONE ''UTC'') NOT NULL,
         CONSTRAINT schema_migrations_pk PRIMARY KEY (schema_migration_version)
       )', _schema, _database
    );
    RETURN 1;
  END IF;
END;
$$
language plpgsql;
`

var stmntCreateSchemaMigrationsTable = `
  SELECT *
  FROM pg_temp.create_schema_migrations_table(
    CAST(NULLIF($1, NULL) AS TEXT),
    CAST(NULLIF($2, NULL) AS TEXT)
  );
`

var stmntCreateExtensionDBLink = `
  CREATE EXTENSION IF NOT EXISTS dblink;
`

var stmntDropExtensionDBLink = `
  DROP EXTENSION IF NOT EXISTS dblink;
`

var stmntCreateDatabaseFn = `
CREATE FUNCTION pg_temp.create_database(
    _template_db TEXT,
    _database TEXT,
    _owner TEXT
) RETURNS INTEGER AS
$$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_database WHERE datname = _database) THEN
    RAISE NOTICE 'database: %s already exists', _database;
  ELSE
    PERFORM dblink_connect('conn','host=localhost port=5432 dbname=postgres user=pgmngr password=pgmngr');
    PERFORM dblink_exec(
      'conn',
      format('CREATE DATABASE %I WITH OWNER = %I', _database, _owner)::TEXT
    );
    PERFORM dblink_disconnect('conn');
    RETURN 1;
  END IF;
END;
$$
language plpgsql;
`

var stmntCreateDatabase = `
SELECT * FROM pg_temp.create_database(
  CAST(NULLIF($1, NULL) AS TEXT),
  CAST(NULLIF($2, NULL) AS TEXT),
  CAST(NULLIF($3, NULL) AS TEXT)
);
`

var stmntDropDatabaseFn = `
CREATE FUNCTION pg_temp.drop_database(
  _database TEXT
) RETURNS bool AS
$$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = _database) THEN
    RAISE EXCEPTION 'database: %s does not exists', _database;
  ELSE
    PERFORM dblink_connect('conn','host=localhost port=5432 dbname=postgres user=pgmngr password=pgmngr');
    PERFORM dblink_exec(
      'conn',
      format('DROP DATABASE %I', _database),
      true
    );
    PERFORM dblink_disconnect('conn');
    RETURN 1;
  END IF;
END;
$$
language plpgsql;
`

var stmntDropDatabase = `
SELECT * FROM pg_temp.drop_database(
    CAST(NULLIF($1, NULL) AS TEXT)
);
`

var stmntDBexists = `
SELECT EXISTS(
    SELECT 1 FROM pg_catalog.pg_database WHERE lower(datname) = lower($1)
);
`
