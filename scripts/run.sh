#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

#Значения данным переменным задаю при запуске данного скрипта в консоли
SERVER_HOST="${SERVER_HOST:-}"
SERVER_USER="${SERVER_USER:-}"
SSH_PORT="${SSH_PORT:-22}"
SSH_KEY_PATH="${SSH_KEY_PATH:-}"

if [[ -z "$SERVER_HOST" || -z "$SERVER_USER" ]]; then
  echo "SERVER_HOST или SERVER_USER не заданы. Запускаю локально."
  exec go run "${ROOT_DIR}/main.go"
fi

SSH_OPTS=(-p "$SSH_PORT")
if [[ -n "$SSH_KEY_PATH" ]]; then
  SSH_OPTS+=(-i "$SSH_KEY_PATH")
fi

tar -C "$ROOT_DIR" -czf - main.go go.mod go.sum internal scripts \
  | ssh "${SSH_OPTS[@]}" "${SERVER_USER}@${SERVER_HOST}" "mkdir -p ~/price-service && tar -xzf - -C ~/price-service"

ssh "${SSH_OPTS[@]}" "${SERVER_USER}@${SERVER_HOST}" "cd ~/price-service && ./scripts/prepare.sh && nohup go run main.go > app.log 2>&1 &"

echo "$SERVER_HOST"
