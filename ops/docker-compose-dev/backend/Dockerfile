FROM golang:1.18-bullseye

COPY . /app

WORKDIR /app

ENTRYPOINT [ "/app/ops/docker-compose-dev/backend/docker-entrypoint.sh" ]
