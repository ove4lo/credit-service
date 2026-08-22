# credit-service

A backend system for handling credit applications, written in Go. Built as two services communicating through a message queue, with authentication, structured logging, tests, graceful shutdown, and observability.

This is a learning-driven project, but built with production-minded practices rather than as a tutorial follow-along.

## Architecture

```
                 POST /applications (JWT)
client ────────────────────────────────►  credit-service (API)
                                             │  saves application (status: new)
                                             │  publishes debt-check task
                                             ▼
                                          RabbitMQ  ──► debt_check queue
                                             │
                                             ▼
                                           worker  (pool of goroutines)
                                             │  reads client's open debt from DB
                                             │  decides approved / rejected
                                             ▼
                                          PostgreSQL (updates status)

Prometheus scrapes /metrics from both services · Grafana visualizes them
```

- **credit-service** — accepts applications over REST, stores them, and publishes a debt-check task to the queue. Responds immediately
- **worker** — consumes tasks concurrently, checks the client's outstanding debt in the database, and writes the decision back

## Stack
- Go (`net/http`, `database/sql`, `log/slog`);
- PostgreSQL (`pgx`);
- RabbitMQ (`amqp091-go`);
- JWT (`golang-jwt`);
- Prometheus + Grafana;
- Docker + Docker Compose.

## Running
Docker is the only requirement

```bash
docker compose up --build
```
Services:
- credit-service API — `http://localhost:4000`
- worker metrics/health — `http://localhost:4001`
- RabbitMQ management — `http://localhost:15672` (guest/guest)
- Prometheus — `http://localhost:9090`
- Grafana — `http://localhost:3000` (admin/admin)

The database schema and seed data are applied automatically on a clean start

## API

**Get a token** (auth is stubbed — see design decisions):
```bash
curl -X POST localhost:4000/login -d '{"username":"alina"}'
```

**Create an application** (requires a token):
```bash
curl -X POST localhost:4000/applications -H "Authorization: Bearer <TOKEN>"  -d '{"client":"Ivan","amount":1000,"term":12}'
```
`id` and `status` are set by the server, not the client. Invalid input returns `422`; missing/invalid token returns `401`.

**Service endpoints:** `GET /healthz` (liveness), `GET /metrics` (Prometheus).

## Design decisions & trade-offs

Decisions I made deliberately, and where I chose to simplify:

- **Two services + a queue instead of one service.** Debt checking is potentially slow, so the API stays responsive: it saves the application and delegates the check via RabbitMQ. This also makes the check independently scalable and resilient — if the worker is down, tasks wait in the queue instead of being lost.
- **Manual ack (not auto-ack).** The worker acknowledges a message only after the decision is written to the DB. If it crashes mid-processing, the message returns to the queue (at-least-once delivery). The stub logic is deterministic so re-processing is safe.
- **Request DTO separate from the domain model.** Incoming JSON is decoded into a struct that has no `id`/`status` fields, so a client cannot set them — protection at the type level, not by convention.
- **Parameterized SQL everywhere** — no string concatenation, so no SQL injection.
- **Errors wrapped with `%w`, logged once** at the handler layer; lower layers add context and return.
- **Graceful shutdown** on both services: stop accepting new work, let in-flight work finish, then close resources (server first, DB last — reverse of dependency order).
- **Metrics live where the event happens** — the API counts received applications, the worker counts decisions (labeled approved/rejected).

Intentional simplifications (would be different in a real system):

- **Auth is stubbed** — `/login` issues a JWT without checking a password. In a real setup, token issuing belongs to a separate auth service; this service only *verifies* tokens.
- **Debt is looked up by client name.** A name is not a stable identifier — real systems identify a client by a unique key (e.g. an ID number) and would model the client as its own entity, with applications and debts referencing it. Name lookup is a simplification for scope.
- **Debt data is seeded, and debts are never closed here.** Closing a debt (via payment) is another service's responsibility; this service only reads debt state.
- **Scoring is a stub** (open debt over a threshold → rejected). The point of the project is the architecture around the decision, not the scoring itself — real logic would replace the stub without changing the surrounding system.

## Observability
Both services expose `/metrics`; Prometheus scrapes them; Grafana visualizes. `/healthz` gives a liveness signal for orchestrators.

## Structure
```text
├── cmd/
│   ├── creditservice/ # API service: handlers, DTO, validation, auth, metrics
│   └── worker/ # queue consumer: worker pool, debt check, decision
├── internal/application/ # domain model, storage layer, DB schema + seed
├── Dockerfile # API image
├── Dockerfile.worker # worker image
├── docker-compose.yml # all services
└── prometheus.yml # scrape config
```

## Roadmap
gRPC between services, Kubernetes deployment, readiness checks (DB/queue reachability), provisioned Grafana dashboards
