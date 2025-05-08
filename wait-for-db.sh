#!/bin/sh

# Wait for Postgres to be ready before proceeding
echo "Checking if Postgres is ready at $DB_HOST:$DB_PORT..."

until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; do
  echo "Waiting for Postgres to be ready..."
  sleep 2
done

echo "Postgres is ready, starting the application..."
exec "$@"

