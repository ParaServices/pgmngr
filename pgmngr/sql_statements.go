package pgmngr

var stmntInsertSchemaMigrationFn = `
CREATE FUNCTION pg_temp.create_schema_migration(
    _schema VARCHAR,
    _table_name VARCHAR,
    _schema_migration_verson INT8
) RETURNS VOID AS
$$
BEGIN
  EXECUTE format(
    'INSERT INTO %I.%I(schema_migration_version)
    VALUES (%s)', _schema, _table_name, _schema_migration_verson
  );
END;
$$
language plpgsql;
`
var stmntInsertSchemaMigration = `
SELECT * FROM pg_temp.create_schema_migration(
  CAST(NULLIF($1, NULL) AS VARCHAR),
  CAST(NULLIF($2, NULL) AS VARCHAR),
  CAST(NULLIF($3, NULL) AS INT8)
)
`

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

var stmntCreateDatabaseFn = `
CREATE FUNCTION pg_temp.create_database(
  _host TEXT,
  _port TEXT,
  _template_db TEXT,
  _admin_username TEXT,
  _admin_password TEXT,
  _database TEXT,
  _owner TEXT
) RETURNS INTEGER AS
$$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_database WHERE datname = _database) THEN
    RAISE NOTICE 'database: %s already exists', _database;
  ELSE
    PERFORM dblink_connect('conn',format('host=%I port=%s dbname=%I user=%I password=%I',
      _host,
      _port,
      _template_db,
      _admin_username,
      _admin_password
    ));
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
  CAST(NULLIF($3, NULL) AS TEXT),
  CAST(NULLIF($4, NULL) AS TEXT),
  CAST(NULLIF($5, NULL) AS TEXT),
  CAST(NULLIF($6, NULL) AS TEXT),
  CAST(NULLIF($7, NULL) AS TEXT)
);
`

var stmntDropDatabaseFn = `
CREATE FUNCTION pg_temp.drop_database(
  _host TEXT,
  _port TEXT,
  _template_db TEXT,
  _admin_username TEXT,
  _admin_password TEXT,
  _database TEXT
) RETURNS bool AS
$$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = _database) THEN
    RAISE EXCEPTION 'database: %s does not exists', _database;
  ELSE
    PERFORM dblink_connect('conn',format('host=%I port=%s dbname=%I user=%I password=%I',
      _host,
      _port,
      _template_db,
      _admin_username,
      _admin_password
    ));
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
  CAST(NULLIF($1, NULL) AS TEXT),
  CAST(NULLIF($2, NULL) AS TEXT),
  CAST(NULLIF($3, NULL) AS TEXT),
  CAST(NULLIF($4, NULL) AS TEXT),
  CAST(NULLIF($5, NULL) AS TEXT),
  CAST(NULLIF($6, NULL) AS TEXT)
);
`

var stmntDBexists = `
SELECT EXISTS(
  SELECT 1 FROM pg_catalog.pg_database WHERE lower(datname) = lower($1)
);
`
