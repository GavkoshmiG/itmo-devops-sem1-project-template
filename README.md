# Финальный проект 1 семестра

REST API сервис для загрузки и выгрузки данных о ценах.

## Требования к системе

- Go 1.23+
- PostgreSQL 13+
- zip/unzip

## Установка и запуск

1. Подготовить базу данных:

```bash
./scripts/prepare.sh
```

2. Запустить сервер:

```bash
./scripts/run.sh
```

Сервер слушает `:8080`.

Если PostgreSQL работает не на локальном хосте, можно передать переменные:

```bash
DB_HOST=localhost DB_PORT=5432 DB_USER=validator DB_PASSWORD=val1dat0r DB_NAME=project-sem-1 go run main.go
```

### Запуск на удалённой машине через SSH

Скрипт `scripts/run.sh` копирует проект на сервер и запускает его там:

```bash
SERVER_HOST=<IP-сервера> SERVER_USER=<USER> SSH_KEY_PATH=~/.ssh/<имя-ключа> ./scripts/run.sh
```

## API

### POST /api/v0/prices?type=zip|tar

- Тело запроса: `multipart/form-data`, поле `file` с архивом.
- `type` по умолчанию `zip`.
- Возвращает JSON со статистикой по базе.

### GET /api/v0/prices

- Возвращает ZIP-архив с файлом `data.csv`.

## Тестирование

Директория `sample_data` - это пример директории, которая является разархивированной версией файла `sample_data.zip`

Запуск тестов для продвинутого уровня:

```bash
./scripts/tests.sh 2
```

Для простого и сложного уровней:

```bash
./scripts/tests.sh 1
./scripts/tests.sh 3
```

## Контакт

- tg: **@van_dark01**
