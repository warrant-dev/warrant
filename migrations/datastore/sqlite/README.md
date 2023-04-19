# Running Warrant with SQLite

This guide will cover how to run Warrant with SQLite as the datastore/eventstore. Note that running Warrant with SQLite requires that you build Warrant from source.

## Running the Binary

### Install SQLite

Many operating systems (like MacOS) come with SQLite pre-installed. If you already have SQLite installed, you can skip to the next step. If you don't already have SQLite installed, [install it](https://www.tutorialspoint.com/sqlite/sqlite_installation.htm). Once installed, you should be able to run the following command to print the currently installed version of SQLite:

```bash
sqlite3 --version
```

### Install Go

[Install Go](https://go.dev/doc/install).

### Build Warrant From Source

Clone the Warrant repository from GitHub or download and unzip the tarball containing the source code for the [latest Warrant release](https://github.com/warrant-dev/warrant/releases/latest).

```bash
tar -xvf <name_of_tarball>
```

Navigate to the `cmd/warrant` directory and run `make dev` to build Warrant. This will create a file called `warrant` that contains the executable.

### Create `warrant.yaml` Configuration

Create a file called `warrant.yaml` in the directory containing the Warrant binary. Add properties to configure SQLite as both the datastore and evenstore for Warrant.

```yaml
# warrant.yaml
port: 8000
logLevel: 0
enableAccessLog: true
apiKey: replace_with_api_key
autoMigrate: true
datastore:
  sqlite:
    database: warrant
    inMemory: true
eventstore:
  sqlite:
    database: warrantEvents
    inMemory: true
```

NOTE: By default, SQLite will create a database file for both the database and eventstore. The filenames are configurable using the `database` property under `datastore` and `eventstore`. Specifying the `inMemory` option under `datastore` or `eventstore` will bypass creation of a database file and run the SQLite database completely in memory. When running Warrant with the `inMemory` configuration, **any data in Warrant will be lost once the Warrant process is shutdown/killed**.

### Run the Executable

```bash
./warrant
```
