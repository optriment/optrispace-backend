# Optrispace backend

Prerequisites:

* docker engine v20
* docker-compose v1.29
* GNU Make v4 in PATH location

## Start development version

If you need to test development version of the backend just clone repo and
issue commands in the root directory:

```sh
make docker-compose-up
```

This will start REST API on [http://localhost:8080/](http://localhost:8080/).

After that you may want to populate database with sample data. Just run:

```sh
make seed
```

It's required sometimes to use clean run the application.
For this just use following command:

```sh
make docker-compose-down
```

## Run integration tests

All test-specific commands provided by `Makefile` use `test-` prefix:

* `test-start-database`
* `test-start-backend`
* `test-integration`

Prerequisites:

* Application with database successfully started
* Go v1.18 installed in PATH location

All required environment variables for testing purpose are set in `Makefile`.

In the first terminal run the following command to start testing database:

```sh
make test-start-database
```

It will run PostgreSQL on 65433 port.

After that in another terminal session start the backend server:

```sh
make test-start-backend
```

It will run applicaton's backend on tcp/8081 port.

Now you are able to run integration tests:

```sh
make test-integration
```
