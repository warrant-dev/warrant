# Running Warrant with MySQL

This guide covers how to set up MySQL as a datastore for Warrant.

Note: Please first refer to the [development guide](/development.md) to ensure that your Go environment is set up and you have checked out the Warrant source or [downloaded a binary](https://github.com/warrant-dev/warrant/releases).

## Install MySQL

Follow the [MySQL Installation Guide](https://dev.mysql.com/doc/mysql-installation-excerpt/8.0/en/) for your OS to install and start MySQL. For MacOS users, we recommend [installing MySQL using homebrew](https://formulae.brew.sh/formula/mysql).

## Warrant configuration

The Warrant server requires certain configuration, defined either within a `warrant.yaml` file (located within the same directory as the binary) or via environment variables. This configuration includes some common variables and some MySQL specific variables. Here's a sample config:

### Sample `warrant.yaml` config

```yaml
port: 8000
logLevel: 1
enableAccessLog: true
autoMigrate: true
authentication:
  apiKey: replace_with_api_key
datastore:
  mysql:
    username: replace_with_username
    password: replace_with_password
    hostname: 127.0.0.1
    database: warrant
```

Note: You can create a database via the mysql command line and configure it as the `database` attribute under `datastore`.

You can also customize your database connection by providing a [DSN (Data Source Name)](https://github.com/go-sql-driver/mysql#dsn-data-source-name). If provided, this string is used to open the given database rather using the individual variables, i.e. `user`, `password`, `hostname`.

```yaml
datastore:
  mysql:
    dsn: root:@tcp(127.0.0.1:3306)/warrant?parseTime=true
```
Note: `parseTime=true` must be included when providing a DSN to parse `DATE` and `DATETIME` values to `time.Time`.

## Running db migrations

Warrant uses [golang-migrate](https://github.com/golang-migrate/migrate) to manage sql db migrations. If the `autoMigrate` config flag is set to true, the server will automatically run migrations on start. If you prefer managing migrations and upgrades manually, please set the `autoMigrate` flag to false.

You can [install golang-migrate yourself](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) and run the MySQL migrations manually:

```shell
migrate -path ./migrations/datastore/mysql/ -database mysql://username:password@hostname/warrant up
```
