#!/bin/bash

set -euo pipefail

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-validator}"
DB_PASSWORD="${DB_PASSWORD:-val1dat0r}"
DB_NAME="${DB_NAME:-project-sem-1}"
DB_TABLE="${DB_TABLE:-prices}"

if ! command -v psql >/dev/null 2>&1; then
  echo "psql не найден. Установите PostgreSQL клиент."
  exit 1
fi

run_psql() {
  if PGPASSWORD="" psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -d postgres -c '\q' >/dev/null 2>&1; then
    psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -d postgres -v ON_ERROR_STOP=1 "$@"
    return
  fi

  if sudo -n true >/dev/null 2>&1; then
    sudo -u postgres psql -v ON_ERROR_STOP=1 "$@"
    return
  fi

  echo "Нет доступа к пользователю postgres. Запустите скрипт от пользователя с правами на создание БД."
  exit 1
}

run_psql <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${DB_USER}') THEN
    CREATE ROLE ${DB_USER} LOGIN PASSWORD '${DB_PASSWORD}';
  END IF;
END
\$\$;

DO \$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = '${DB_NAME}') THEN
    CREATE DATABASE "${DB_NAME}" OWNER ${DB_USER};
  END IF;
END
\$\$;
SQL

PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 <<SQL
CREATE TABLE IF NOT EXISTS ${DB_TABLE} (
  id BIGINT,
  name TEXT,
  category TEXT,
  price DOUBLE PRECISION,
  create_date DATE
);
SQL

echo "База данных подготовлена."
