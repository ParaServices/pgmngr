# pgmngr

`pgmngr` or Postgres Manager is a package that facilitates migration of a
given [Postgres] database.

This package is inspired from [pgmgr](https://github.com/rnubel/pgmgr.git)
which is a great package but had some features that are missing that I needed.

One of the features was the independence from external dependencies like
`postgresql-client` or `postgresql`. This package does not require that and
solely relies on the `sql/db` package for creating and dropping databases.

To use the package, simply run:

```
$ go run main.go [help]

```

or if compiled

```
$ pgmngr [help]
```

TODO:

 - [] Schema dump
 - [] Rollback migration

[Postgres]: https://postgresql.org
