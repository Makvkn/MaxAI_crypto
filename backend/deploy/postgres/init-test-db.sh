#!/bin/sh
# Creates the database used by the integration test harness, so `make up`
# followed by `make test-integration` works on a fresh machine.
set -e

psql --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE DATABASE maxai_test OWNER $POSTGRES_USER;
EOSQL
