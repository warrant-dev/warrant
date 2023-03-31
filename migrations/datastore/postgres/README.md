# Self-hosting Warrant with PostgreSQL

This guide will cover how to self-host Warrant with PostgreSQL as the datastore. Note that Warrant only supports versions of PostgreSQL >= 14.7.

## Docker Compose (Recommended)

The following [Docker Compose](https://docs.docker.com/compose/) manifest will create a PostgreSQL database, setup the database schema required by Warrant, and start Warrant. You can also accomplish this by running Warrant with [Kubernetes](https://kubernetes.io/).

```yaml
version: "3.9"
services:
  database:
    image: postgres:14.7
    environment:
      POSTGRES_PASSWORD: replace_with_password
    ports:
      - 5432:5432
    healthcheck:
      test: ["CMD", "pg_isready", "-d", "warrant"]
      timeout: 5s
      retries: 10

  web:
    image: warrantdev/warrant
    ports:
      - 8000:8000
    depends_on:
      database:
        condition: service_healthy
    environment:
      WARRANT_PORT: 8000
      WARRANT_LOGLEVEL: 1
      WARRANT_ENABLEACCESSLOG: "true"
      WARRANT_DATASTORE: postgres
      WARRANT_DATASTORE_POSTGRES_USERNAME: postgres
      WARRANT_DATASTORE_POSTGRES_PASSWORD: replace_with_password
      WARRANT_DATASTORE_POSTGRES_HOSTNAME: database
      WARRANT_DATASTORE_POSTGRES_DATABASE: warrant
      WARRANT_DATASTORE_POSTGRES_SSLMODE: disable
      WARRANT_EVENTSTORE: postgres
      WARRANT_EVENTSTORE_POSTGRES_USERNAME: postgres
      WARRANT_EVENTSTORE_POSTGRES_PASSWORD: replace_with_password
      WARRANT_EVENTSTORE_POSTGRES_HOSTNAME: database
      WARRANT_EVENTSTORE_POSTGRES_DATABASE: warrant_events
      WARRANT_EVENTSTORE_POSTGRES_SSLMODE: disable
      WARRANT_API_KEY: replace_with_api_key
      WARRANT_AUTHENTICATION_PROVIDER: replace_with_authentication_name
      WARRANT_AUTHENTICATION_PUBLICKEY: replace_with_authentication_public_key
      WARRANT_AUTHENTICATION_USER_ID_CLAIM: replace_with_authentication_user_id_claim
      WARRANT_AUTHENTICATION_TENANT_ID_CLAIM: replace_with_authentication_tenant_id_claim
```

## Running the Binary

### Setup and Configure PostgreSQL

Install and run the [PostgreSQL Installer](https://www.postgresql.org/download/) for your OS to install and start PostgreSQL. For MacOS users, we recommend [installing PostgreSQL using homebrew](https://formulae.brew.sh/formula/postgresql@14).

### Download the Binary

Download the latest warrant binary for your architecture from [here](https://github.com/warrant-dev/warrant/releases/latest). Then unzip the tarball and run the `warrant` executable.

```bash
tar -xvf <name_of_tarball>
```

### Create `warrant.yaml` Configuration

Create a file called `warrant.yaml` in the directory containing the Warrant binary.

```bash
cd <path_to_untarred_directory>
touch warrant.yaml
```

```yaml
# warrant.yaml
port: 8000
logLevel: 0
enableAccessLog: "true"
apiKey: replace_with_api_key
datastore:
  postgres:
    username: replace_with_username
    password: replace_with_password
    hostname: localhost
    database: warrant
    sslmode: disable
eventstore:
  postgres:
    username: replace_with_username
    password: replace_with_password
    hostname: localhost
    database: warrant_events
    sslmode: disable
```

### Run the Binary

```bash
./warrant
```
