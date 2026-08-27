# Sharing Vision Article API

A REST API for the Sharing Vision full-stack test. The service is written in Go, uses MySQL through `database/sql`, and keeps HTTP, business, validation, and persistence responsibilities in separate layers.

## Features

- Create, list, read, update, and delete articles
- Manual validation with field-specific messages
- Raw parameterized SQL without an ORM or `SELECT *`
- Stable pagination ordered by `created_date DESC, id DESC`
- A maximum page size of 100
- Per-query context timeouts
- Configured MySQL connection pooling
- SQL migrations managed by `golang-migrate`
- Graceful HTTP server shutdown

## Requirements

- Go 1.22 or newer
- MySQL 8.0 or a compatible version
- [`Air`](https://github.com/air-verse/air) for development hot reload
- [`golang-migrate`](https://github.com/golang-migrate/migrate) CLI

## Project structure

```text
.
├── cmd/main.go
├── db/migrations
├── internal
│   ├── config
│   ├── database
│   ├── handler
│   ├── model
│   ├── repository
│   ├── service
│   └── validator
├── .air.toml
├── .env.example
├── go.mod
└── README.md
```

The request flow is:

```text
HTTP Handler -> Article Service -> Article Repository -> MySQL
```

## Database setup

Install `golang-migrate` with MySQL support once if it is not already available:

```bash
go install -tags mysql github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Ensure the Go binary directory (usually `%USERPROFILE%\go\bin` on Windows) is included in `PATH`.

The local defaults match the requested environment:

| Setting | Default |
|---|---|
| Host | `localhost:3306` |
| User | `root` |
| Password | empty |
| Database | `sharing_vision_test` |

Create the database once:

```sql
CREATE DATABASE IF NOT EXISTS sharing_vision_test
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;
```

Run the migration from the project root:

```bash
migrate -path db/migrations -database "mysql://root:@tcp(localhost:3306)/sharing_vision_test" up
```

Roll back the migration when needed:

```bash
migrate -path db/migrations -database "mysql://root:@tcp(localhost:3306)/sharing_vision_test" down 1
```

Migration versions are tracked by `golang-migrate`. The SQL also uses `IF NOT EXISTS` and `IF EXISTS` so rerunning a migration command does not overwrite article data.

## Configuration

The defaults work without an `.env` loader. Set environment variables in your shell when different values are needed:

| Variable | Default |
|---|---|
| `SERVER_ADDRESS` | `:8080` |
| `DB_HOST` | `localhost:3306` |
| `DB_USER` | `root` |
| `DB_PASSWORD` | empty |
| `DB_NAME` | `sharing_vision_test` |
| `DB_MAX_OPEN_CONNS` | `25` |
| `DB_MAX_IDLE_CONNS` | `25` |

`.env.example` is provided as a reference. Go does not read it automatically, which avoids adding a runtime dependency only for local configuration.

## Run the service for development

Install Air once if it is not already available:

```bash
go install github.com/air-verse/air@latest
```

Download the project dependencies:

```bash
go mod download
```

Start the development server with hot reload:

```bash
air
```

Development uses Air through `.air.toml`; run `air` instead of `go run`. Air rebuilds the service whenever a Go source file changes and stores temporary binaries in the ignored `tmp/` directory.

The API listens on `http://localhost:8080` by default. The application verifies the database connection during startup and exits with a clear error if MySQL or the database is unavailable.

## API endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/article/` | Create an article |
| `GET` | `/article/{limit}/{offset}` | List articles |
| `GET` | `/article/{id}` | Get one article |
| `PUT` | `/article/{id}` | Replace an article |
| `PATCH` | `/article/{id}` | Replace an article |
| `DELETE` | `/article/{id}` | Delete an article |

Both `PUT` and `PATCH` validate all four writable fields, as required by the test specification.

### Valid article payload

```json
{
  "title": "A sufficiently long article title",
  "content": "This article content must contain at least two hundred characters. Add the complete article text here until the minimum length is reached. Validation counts Unicode characters and ignores surrounding whitespace before saving the value to MySQL.",
  "category": "technology",
  "status": "publish"
}
```

Validation rules:

- `title`: required, at least 20 characters
- `content`: required, at least 200 characters
- `category`: required, at least 3 characters
- `status`: exactly `publish`, `draft`, or `thrash`

### Create an article

```bash
curl -X POST http://localhost:8080/article/ \
  -H "Content-Type: application/json" \
  -d '{"title":"A sufficiently long article title","content":"This article content must contain at least two hundred characters. Add the complete article text here until the minimum length is reached. Validation counts Unicode characters and ignores surrounding whitespace before saving the value to MySQL.","category":"technology","status":"publish"}'
```

A successful create returns `201 Created` and the inserted article, including its generated `id`.

### List articles

```bash
curl http://localhost:8080/article/10/0
```

`limit` must be from 1 through 100, and `offset` must be zero or greater. An empty result is returned as `[]`.

### Get an article

```bash
curl http://localhost:8080/article/1
```

### Update an article

```bash
curl -X PUT http://localhost:8080/article/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"An updated and sufficiently long title","content":"This updated content must also contain at least two hundred characters. Add the complete replacement article here until the required content length is reached before sending this request to the article API endpoint.","category":"backend","status":"draft"}'
```

### Delete an article

```bash
curl -X DELETE http://localhost:8080/article/1
```

## Error responses

Invalid fields return `400 Bad Request` with individual messages:

```json
{
  "error": "validation failed",
  "fields": {
    "title": "title must be at least 20 characters",
    "status": "status must be one of: publish, draft, thrash"
  }
}
```

A missing article returns `404 Not Found`:

```json
{
  "error": "article not found"
}
```

Unexpected server errors are logged internally and return a generic `500 Internal Server Error` response.

## Tests

Run all unit tests:

```bash
go test ./...
```

The service tests use a repository test double and do not require a running MySQL instance.
