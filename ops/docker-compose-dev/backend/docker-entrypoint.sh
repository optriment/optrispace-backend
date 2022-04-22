#!/bin/sh -eu

cd /app/pkg/db/db-migrations

env DB_URL=$DB_URL make migrate-up

cd /app

_term() { 
  echo "Caught SIGTERM signal!" 
  kill -TERM "$child" 2>/dev/null
}

trap _term TERM INT

go run . run --config /app/ops/docker-compose-dev/backend/dev.yaml -d $DB_URL &

child=$!
wait "$child"
