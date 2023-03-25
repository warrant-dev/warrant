# Self-hosting Warrant with MySQL

This guide will cover how to self-host Warrant with MySQL as the datastore. Note that Warrant only supports versions of MySQL >= 8.0.32.

## Docker Compose (Recommended)

The following [Docker Compose](https://docs.docker.com/compose/) manifest will create a MySQL database, setup the database schema required by Warrant, and start Warrant. You can also accomplish this by running Warrant with [Kubernetes](https://kubernetes.io/).

```yaml
version: "3.9"
services:
  database:
    image: mysql:8.0.32
    environment:
      MYSQL_USER: replace_with_username
      MYSQL_PASSWORD: replace_with_password
    ports:
      - 3306:3306
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
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
      WARRANT_DATASTORE: mysql
      WARRANT_DATASTORE_MYSQL_USERNAME: replace_with_username
      WARRANT_DATASTORE_MYSQL_PASSWORD: replace_with_password
      WARRANT_DATASTORE_MYSQL_HOSTNAME: database
      WARRANT_DATASTORE_MYSQL_DATABASE: warrant
      WARRANT_EVENTSTORE: mysql
      WARRANT_EVENTSTORE_MYSQL_USERNAME: replace_with_username
      WARRANT_EVENTSTORE_MYSQL_PASSWORD: replace_with_password
      WARRANT_EVENTSTORE_MYSQL_HOSTNAME: database
      WARRANT_EVENTSTORE_MYSQL_DATABASE: warrantEvents
      WARRANT_API_KEY: replace_with_api_key
```

## Running the Binary

### Setup and Configure MySQL

Follow the [MySQL Installation Guide](https://dev.mysql.com/doc/mysql-installation-excerpt/8.0/en/) for your OS to install and start MySQL. For MacOS users, we recommend [installing MySQL using homebrew](https://formulae.brew.sh/formula/mysql).

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
  mysql:
    username: replace_with_username
    password: replace_with_password
    hostname: 127.0.0.1
    database: warrant
eventstore:
  mysql:
    username: replace_with_username
    password: replace_with_password
    hostname: 127.0.0.1
    database: warrantEvents
```

### Run the Binary

```bash
./warrant
```
