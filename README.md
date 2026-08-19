# credit-service

A backend service for handling credit applications, written in Go. Built incrementally, from a single service toward a microservice system

Current state: accepts credit applications over a REST API with input validation, JWT auth, structured logging, and PostgreSQL, all running via Docker Compose

## Stack
- Go (`net/http`, `database/sql`, `log/slog`);
- PostgreSQL (`pgx` driver);
- JWT (`golang-jwt`);
- Docker + Docker Compose.

## Running
Docker is the only requirement

```bash
docker compose up --build
```
Two containers start: the app and PostgreSQL. The app waits for the database to be healthy, and the table is created automatically from `internal/application/schema.sql`

The service listens on `http://localhost:4000`.

## API

### Get a token
Auth is stubbed for now (no password check) — it issues a JWT for a given username. In a real setup, token issuing belongs to a separate auth service
```bash
curl -X POST localhost:4000/login -d '{"username":"alina"}'
```
Response: `{"token":"eyJ..."}`

### Create an application (requires a token)
```bash
curl -X POST localhost:4000/applications \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"client":"Alina","amount":1000,"term":12}'
```
Response:
```json
{"id":1,"client":"Alina","amount":1000,"term":12,"status":"new"}
```
`id` is assigned by the database; `status` (`new`) is set by the server. 
The client cannot set them — the request body is decoded into a separate input struct that has no such fields

Requests without a valid token get `401` or invalid input (empty client, non-positive amount or term) gets `422`

## Structure
```text
├── cmd/creditservice/     # entry point, HTTP layer
│   ├── main.go            # startup: wiring and launch
│   └── server.go          # HTTP layer: handlers, request DTO, validation, auth middleware
├── internal/application/  # application model, storage layer (PostgreSQL), DB schema
├── Dockerfile             # multi-stage build
└── docker-compose.yml     # app + PostgreSQL
```

## Security & practices
- Parameterized SQL queries (no string concatenation);
- Input validated before touching the database;
- JWT verification middleware on protected endpoints; signing algorithm is checked explicitly;
- Secrets (`DATABASE_DSN`, `JWT_SECRET`) are read from the environment, not hardcoded;
- Structured JSON logs to stdout; errors are wrapped with context (`%w`) and logged once;

## Configuration
Set via environment variables (in `docker-compose.yml`): `DATABASE_DSN`, `JWT_SECRET`

## Roadmap
Graceful shutdown, a second service (debt check) via a message queue, metrics, and integration tests