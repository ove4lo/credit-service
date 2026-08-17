# credit-service

A backend service for handling credit applications, written in Go. Built incrementally, from a single service toward a microservice system

Current stage: accepting an application over a REST API and storing it in PostgreSQL, all running via Docker Compose

## Stack
- Go (`net/http`, `database/sql`);
- PostgreSQL (`pgx` driver);
- Docker + Docker Compose.

## Running
Docker is the only requirement

```bash
docker compose up --build
```
Two containers start: the app and PostgreSQL. The app waits for the database to be healthy, and the table is created automatically from `internal/application/schema.sql`

The service listens on `http://localhost:4000`.

## API

### Create an application
```bash
curl -X POST localhost:4000/applications \
  -d '{"client":"Alina","amount":1000,"term":12}'
```
Response:
```json
{"id":1,"client":"Alina","amount":1000,"term":12,"status":"new"}
```
`id` is assigned by the database; `status` (`new`) is set by the server — the client does not provide them

## Structure
```text
├── cmd/creditservice/     # entry point, HTTP layer
├── internal/application/  # application model, storage layer (PostgreSQL), DB schema
├── Dockerfile             # multi-stage build
└── docker-compose.yml     # app + PostgreSQL
```

## Configuration
The database address is passed via the `DATABASE_DSN` environment variable (set in `docker-compose.yml`) and is not stored in code

## Roadmap
Logging, JWT auth and validation, graceful shutdown, a second service (debt check) via a message queue, metrics, and integration tests
