# Deployment examples

Sample deployment configurations for running Warrant.

## Docker Compose

### Using MySQL as the datastore

This guide will cover how to self-host Warrant with MySQL as the datastore. Note that Warrant only supports versions of MySQL >= 8.0.32.

The following [Docker Compose](https://docs.docker.com/compose/) manifest will create a MySQL database, setup the database schema required by Warrant, and start Warrant. You can also accomplish this by running Warrant with [Kubernetes](https://kubernetes.io/):

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
      WARRANT_ENABLEACCESSLOG: true
      WARRANT_AUTOMIGRATE: true
      WARRANT_CHECK_CONCURRENCY: 4
      WARRANT_CHECK_MAXCONCURRENCY: 1000
      WARRANT_CHECK_TIMEOUT: 1m
      WARRANT_AUTHENTICATION_APIKEY: replace_with_api_key
      WARRANT_DATASTORE_MYSQL_USERNAME: replace_with_username
      WARRANT_DATASTORE_MYSQL_PASSWORD: replace_with_password
      WARRANT_DATASTORE_MYSQL_HOSTNAME: database
      WARRANT_DATASTORE_MYSQL_DATABASE: warrant

```


### Using PostgreSQL as the datastore

This guide will cover how to self-host Warrant with PostgreSQL as the datastore. Note that Warrant only supports versions of PostgreSQL >= 14.7.

The following [Docker Compose](https://docs.docker.com/compose/) manifest will create a PostgreSQL database, setup the database schema required by Warrant, and start Warrant. You can also accomplish this by running Warrant with [Kubernetes](https://kubernetes.io/):

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
      WARRANT_ENABLEACCESSLOG: true
      WARRANT_AUTOMIGRATE: true
      WARRANT_CHECK_CONCURRENCY: 4
      WARRANT_CHECK_MAXCONCURRENCY: 1000
      WARRANT_CHECK_TIMEOUT: 1m
      WARRANT_AUTHENTICATION_APIKEY: replace_with_api_key
      WARRANT_DATASTORE_POSTGRES_USERNAME: postgres
      WARRANT_DATASTORE_POSTGRES_PASSWORD: replace_with_password
      WARRANT_DATASTORE_POSTGRES_HOSTNAME: database
      WARRANT_DATASTORE_POSTGRES_DATABASE: warrant
      WARRANT_DATASTORE_POSTGRES_SSLMODE: disable
```
