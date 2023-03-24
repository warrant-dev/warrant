# Self-hosting Warrant with PostgreSQL

This guide will cover how to self-host Warrant with PostgreSQL as the datastore. Note that Warrant only supports versions of PostgreSQL >= 14.7.

## Docker Compose (Recommended)

The following [Docker Compose](https://docs.docker.com/compose/) manifest will create a PostgreSQL database, setup the database schema required by Warrant, and start Warrant. You can also accomplish this by running Warrant with [Kubernetes](https://kubernetes.io/).

```yaml
version: "3.9"
services:
  datastore:
    image: postgres:14.7
    environment:
      POSTGRES_PASSWORD: replace_with_password
      POSTGRES_DB: warrant
    ports:
      - 5432:5432
    healthcheck:
      test: ["CMD", "pg_isready", "-d", "warrant"]
      timeout: 5s
      retries: 10

  migrate-datastore:
    image: migrate/migrate
    command:
      [
        "-source",
        "github://warrant-dev/warrant/migrations/datastore/postgres",
        "-database",
        "postgres://postgres:replace_with_password@datastore/warrant?sslmode=disable",
        "up",
      ]
    depends_on:
      datastore:
        condition: service_healthy

  web:
    image: warrantdev/warrant
    ports:
      - 8000:8000
    depends_on:
      migrate-datastore:
        condition: service_completed_successfully
    environment:
      WARRANT_PORT: 8000
      WARRANT_LOGLEVEL: 1
      WARRANT_ENABLEACCESSLOG: "true"
      WARRANT_DATASTORE: postgres
      WARRANT_DATASTORE_POSTGRES_USERNAME: postgres
      WARRANT_DATASTORE_POSTGRES_PASSWORD: replace_with_password
      WARRANT_DATASTORE_POSTGRES_HOSTNAME: datastore
      WARRANT_DATASTORE_POSTGRES_DATABASE: warrant
      WARRANT_DATASTORE_POSTGRES_SSLMODE: disable
      WARRANT_API_KEY: replace_with_api_key
```

## Running the Binary

### Setup and Configure PostgreSQL

Install and run the [PostgreSQL Installer](https://www.postgresql.org/download/) for your OS to install and start PostgreSQL. For MacOS users, we recommend [installing PostgreSQL using homebrew](https://formulae.brew.sh/formula/postgresql@14).

### Migrate the Database Schema

Once you've setup and started your database, you can setup the database schema using [golang-migrate](https://github.com/golang-migrate/migrate).

Migrate to the latest schema:

```bash
migrate -source github://warrant-dev/warrant/migrations/datastore/postgres -database postgres://postgres:replace_with_password@/warrant?sslmode=disable up
```

### Download and Run the Binary

Download the latest warrant binary for your architecture from [here](https://github.com/warrant-dev/warrant/releases/latest). Then unzip the tarball and run the `warrant` executable.

```bash
tar -xvf <name_of_tarball>
cd <path_to_untarred_directory>
./warrant
```
