#!/bin/sh -eu

# this script is for starting from inside appropriate docker container

cd /app/src/pkg/internal/db/db-migrations

sleep 1s # wait till DB started though... Hope this will be

env DB_URL=postgres://postgres:postgres@postgres:5432/optrwork?sslmode=disable make migrate-up

cd /app/src

_term() { 
  echo "Caught SIGTERM signal!" 
  kill -TERM "$child" 2>/dev/null
}

trap _term TERM INT

go run . run -L trace --server.host :8080 --pprof.hostport :5005 --server.trace=true &

child=$! 
wait "$child"
