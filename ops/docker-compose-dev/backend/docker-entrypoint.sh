#!/bin/sh -eu

cd /app/pkg/internal/db/db-migrations

make migrate-up

cd /app

_term() { 
  echo "Caught SIGTERM signal!" 
  kill -TERM "$child" 2>/dev/null
}

trap _term TERM INT

go run . run --config /app/app/docker-compose-dev/backend/dev.yaml -d $DB_URL &

child=$!
wait "$child"
