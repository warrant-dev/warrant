# Self-hosting Warrant with MySQL

This guide will cover how to self-host Warrant with MySQL as the datastore. Note that Warrant only supports versions of MySQL >= 8.0.32.

## Docker Compose (Recommended)

The following [Docker Compose](https://docs.docker.com/compose/) manifest will create a MySQL database, setup the database schema required by Warrant, and start Warrant. You can also accomplish this by running Warrant with [Kubernetes](https://kubernetes.io/).

```yaml
version: "3.9"
services:
  datastore:
    image: mysql:8.0.32
    environment:
      MYSQL_USER: replace_with_username
      MYSQL_PASSWORD: replace_with_password
      MYSQL_DATABASE: warrant
    ports:
      - 3306:3306
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      timeout: 5s
      retries: 10

  migrate-datastore:
    image: migrate/migrate
    command:
      [
        "-source",
        "github://warrant-dev/warrant/migrations/datastore/mysql",
        "-database",
        "mysql://replace_with_username:replace_with_password@tcp(datastore:3306)/warrant",
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
      WARRANT_DATASTORE: mysql
      WARRANT_DATASTORE_MYSQL_USERNAME: replace_with_username
      WARRANT_DATASTORE_MYSQL_PASSWORD: replace_with_password
      WARRANT_DATASTORE_MYSQL_HOSTNAME: datastore
      WARRANT_DATASTORE_MYSQL_DATABASE: warrant
```

## Running the Binary

### Setup and Configure MySQL

Follow the [MySQL Installation Guide](https://dev.mysql.com/doc/mysql-installation-excerpt/8.0/en/) for your OS to install and start MySQL. For MacOS users, we recommend [installing MySQL using homebrew](https://formulae.brew.sh/formula/mysql).

### Migrate the Database Schema

Once you've setup and started your database, you can setup the database schema using [golang-migrate](https://github.com/golang-migrate/migrate).

Migrate to the latest schema:

```bash
migrate -source github://warrant-dev/warrant/migrations/datastore/mysql -database mysql://replace_with_username@/warrant up
```

### Download and Run the Binary

Download the latest warrant binary for your architecture from [here](https://github.com/warrant-dev/warrant/releases/latest). Then unzip the tarball and run the `warrant` executable.

```bash
tar -xvf <name_of_tarball>
cd <path_to_untarred_directory>
./warrant
```
