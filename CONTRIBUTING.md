# Contributing to OptriSpace Backend

[![Tests](https://github.com/optriment/optrispace-backend/actions/workflows/tests.yml/badge.svg)](https://github.com/optriment/optrispace-backend/actions/workflows/tests.yml)
[![Spell](https://github.com/optriment/optrispace-backend/actions/workflows/spell.yml/badge.svg)](https://github.com/optriment/optrispace-backend/actions/workflows/spell.yml)

The development branch is `develop`.\
This is the default branch that all Pull Requests (PR) should be made against.

Requirements:

* [Docker](https://www.docker.com/products/docker-desktop/) version 20 or higher
* Docker Compose version 1.29 or higher
* GNU Make version 4

## Prepare

Please follow instructions below to install backend locally.

1. [Fork](https://help.github.com/articles/fork-a-repo/)
   this repository to your own GitHub account

2. [Clone](https://help.github.com/articles/cloning-a-repository/)
   it to your local device

3. Create a new branch:

    ```sh
    git checkout -b YOUR_BRANCH_NAME
    ```

4. Install the dependencies with:

    ```sh
    go install
    ```

## Start the database

```sh
cd ops/docker-compose-dev
docker compose up postgres
```

It will start PostgreSQL on localhost:65432.\
See `./ops/docker-compose-dev/docker-compose.yml`

## Start the web server

In the project root directory simply run:

```sh
make run
```

It will start the backend on [http://localhost:8080/](http://localhost:8080/).

## Populate test database

In the project root directory simply run:

```sh
make seed
```

## Working with database migrations

## Writing tests

## Running tests

All test-specific commands provided by `Makefile` use `test-` prefix:

* `test-start-database`
* `test-start-backend`
* `test-integration`

All required environment variables for testing purpose are set in `Makefile`.

In the first terminal run the following command to start database:

```sh
make test-start-database
```

It will start PostgreSQL on 65433 port.

After that in another terminal session start the backend server:

```sh
make test-start-backend
```

It will start applicaton's backend on tcp/8081 port.

Now you are able to run integration tests:

```sh
make test-integration
```

## Swagger

### View Swagger specification

The application supports Swagger 2.0 specification by using [swag](https://github.com/swaggo/swag).

After starting the application, you can investigate into API just open link [localhost:8080](http://localhost:8080/swagger/index.html). But be aware about actual published port and host to visit interactive the Swagger specification.

### Edit Swagger specification

After changing Swagger related declarative comments, you have to format these comments and update specification. For this, issue command:

```
make swag-fmt swag-init
```

Be patient while swag tool be downloaded check result. All specification changes must be committed into repository. Use git command for it.
