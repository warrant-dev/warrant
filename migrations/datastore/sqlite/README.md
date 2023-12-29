# Running Warrant with SQLite

This guide covers how to set up SQLite as a datastore for Warrant.

Note: Please first refer to the [development guide](/development.md) to ensure that your Go environment is set up and you have checked out the Warrant source or [downloaded a binary](https://github.com/warrant-dev/warrant/releases).

## Install SQLite

Many operating systems (like MacOS) come with SQLite pre-installed. If you already have SQLite installed, you can skip to the next step. If you don't already have SQLite installed, [install it](https://www.tutorialspoint.com/sqlite/sqlite_installation.htm). Once installed, you should be able to run the following command to print the currently installed version of SQLite:

```bash
sqlite3 --version
```

## Warrant configuration

The Warrant server requires certain configuration, defined either within a `warrant.yaml` file (located within the same directory as the binary) or via environment variables. This configuration includes some common variables and some SQLite specific variables. Here's a sample config:

### Sample `warrant.yaml` config

```yaml
port: 8000
logLevel: 1
enableAccessLog: true
autoMigrate: true
authentication:
  apiKey: replace_with_api_key
datastore:
  sqlite:
    database: warrant
    inMemory: true
```

Note: By default, SQLite will create a database file for the datastore. The filename is configurable using the `database` property under `datastore`. Specifying the `inMemory` option under `datastore` will create the database file in memory and will not persist it to the filesystem. When running Warrant with the `inMemory` configuration, **any data in Warrant will be lost once the Warrant process is shutdown/killed**.

Unlike `mysql` and `postgresql`, `sqlite` currently does not support manually running db migrations on the command line via golang-migrate. Therefore, you should keep `autoMigrate` set to true in your Warrant config so that the server runs migrations as part of startup.
