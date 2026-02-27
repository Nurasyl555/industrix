#!/bin/bash
set -e
COMMAND=${1:-up}
SERVICE=${2:-identity_db}
VERSION=${3}
DB_HOST=${POSTGRES_HOST:-localhost}
DB_PORT=${POSTGRES_PORT:-5432}
DB_USER=${POSTGRES_USER:-postgres}
DB_PASSWORD=${POSTGRES_PASSWORD:-devpassword}
DB_NAME=${SERVICE}
MIGRATIONS_DIR="./migrations/${SERVICE}"
DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"
if ! command -v migrate &> /dev/null; then
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi
case $COMMAND in
    up) migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" up ;;
    down) migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" down 1 ;;
    *) exit 1 ;;
esac
