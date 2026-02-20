#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE isr;
    CREATE DATABASE be;
EOSQL

echo "Databases 'isr' and 'be' created successfully"
