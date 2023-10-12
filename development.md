# Local development

## Set up Go

Warrant is written in Go. Prior to cloning the repo and making any code changes, please ensure that your local Go environment is set up. Refer to the appropriate instructions for your platform [here](https://go.dev/).

## Fork & clone repository

We follow GitHub's fork & pull request model. If you're looking to make code changes, it's easier to do so on your own fork and then contribute pull requests back to the Warrant repo. You can create your own fork of the repo [here](https://github.com/warrant-dev/warrant/fork).

If you'd just like to checkout the source to build & run, you can clone the repo directly:

```shell
git clone git@github.com:warrant-dev/warrant.git
```

Note: It's recommended you clone the repository into a directory relative to your `GOPATH` (e.g. `$GOPATH/src/github.com/warrant-dev`)

## Server configuration
To set up your server config file, refer to our [configuration doc](/configuration.md).

## Build binary & start server

After the datastore, eventstore and configuration are set, build & start the server:

```shell
cd cmd/warrant
make dev
./bin/warrant
```

## Make requests

Once the server is running, you can make API requests using curl, any of the [Warrant SDKs](/README.md#sdks), or your favorite API client:

```shell
curl -g "http://localhost:port/v1/object-types" -H "Authorization: ApiKey YOUR_KEY"
```

# Running tests

## Unit tests

```shell
go test -v ./...
```

## End-to-end API tests

The Warrant repo contains a suite of e2e tests that test various combinations of API requests. These tests are defined in json files within the `tests/` dir and are executed using [APIRunner](https://github.com/warrant-dev/apirunner). These tests can be run locally:

### Install APIRunner

```shell
go install github.com/warrant-dev/apirunner/cmd/apirunner@latest
```

### Define test configuration

APIRunner tests run based on a simple config file that you need to create in the `tests/` directory:

```shell
touch tests/apirunner.conf
```

Add the following to your `tests/apirunner.conf` (replace with your server url and api key):

```json
{
    "baseUrl": "YOUR_SERVER_URL",
    "headers": {
        "Authorization" : "ApiKey YOUR_API_KEY"
    }
}
```

### Run tests

First, make sure your server is running:

```shell
./bin/warrant
```

In a separate shell, run the tests:

```shell
cd tests/
apirunner . '.*'
```
