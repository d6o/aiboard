# AIBoard

A kanban board built for humans and AI agents alike. Every feature in the UI is backed by a complete REST API, making it easy for LLM-powered agents to create, update, and manage work alongside people.

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![Postgres](https://img.shields.io/badge/PostgreSQL-17-4169E1?logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)

---

## Quick Start

```bash
make run
```

Open [http://localhost:8080](http://localhost:8080). That's it.

This builds the Go binary, starts PostgreSQL, runs migrations, seeds default users and tags, and serves the app. No manual setup required.

## What It Does

AIBoard is a three-column kanban board (**Todo**, **Doing**, **Done**) with:

- **Cards** with title, description, priority (1-5), reporter, assignee, tags, subtasks, and comments
- **Drag-and-drop** between columns
- **Subtasks** with completion tracking and a 20-item limit per card
- **Tags** as reusable color-coded labels
- **Comments** with @mention support and autocomplete
- **Notifications** triggered by @mentions, subtask completion, and cards moved to Done
- **Activity log** recording every mutation for full auditability
- **Idempotency** via `Idempotency-Key` header to prevent duplicate creates on retries

There is no authentication. Users identify themselves by passing a `user_id` with their requests. A user picker in the header lets you switch identities in the UI.

## API Reference

All endpoints return JSON with this structure:

```json
// Success
{ "data": { ... } }

// Error
{ "error": { "code": "VALIDATION_ERROR", "message": "...", "fields": [...] } }
```

### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/users` | List all users |
| GET | `/api/users/{id}` | Get a user |
| POST | `/api/users` | Create a user |
| DELETE | `/api/users/{id}` | Delete a user (fails if user is reporter/assignee on any card) |

### Cards

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/cards` | List cards (filter: `column`, `assignee_id`, `reporter_id`, `tag_id`, `priority`) |
| GET | `/api/cards/{id}` | Get card with subtasks, comments, and tags |
| POST | `/api/cards` | Create a card |
| PUT | `/api/cards/{id}` | Update a card |
| DELETE | `/api/cards/{id}` | Delete a card |
| PATCH | `/api/cards/{id}/move` | Move card to a different column |

### Subtasks

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/cards/{cardID}/subtasks` | List subtasks |
| POST | `/api/cards/{cardID}/subtasks` | Create a subtask |
| PUT | `/api/cards/{cardID}/subtasks/{id}` | Update a subtask (title, completed) |
| DELETE | `/api/cards/{cardID}/subtasks/{id}` | Delete a subtask |
| PATCH | `/api/cards/{cardID}/subtasks/reorder` | Reorder subtasks |

### Tags

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/tags` | List all tags |
| POST | `/api/tags` | Create a tag |
| DELETE | `/api/tags/{id}` | Delete a tag |
| POST | `/api/cards/{cardID}/tags` | Attach a tag to a card |
| DELETE | `/api/cards/{cardID}/tags/{tagID}` | Detach a tag from a card |

### Comments

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/cards/{cardID}/comments` | List comments |
| POST | `/api/cards/{cardID}/comments` | Create a comment (parses @mentions) |
| DELETE | `/api/cards/{cardID}/comments/{id}` | Delete a comment |

### Notifications

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/users/{userID}/notifications` | List notifications (filter: `unread=true`) |
| PATCH | `/api/users/{userID}/notifications/{id}/read` | Mark one as read |
| PATCH | `/api/users/{userID}/notifications/read-all` | Mark all as read |

### Activity Log

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/activity` | List activity (filter: `card_id`, `user_id`, `action`) |

### Board Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/board/reset` | Wipe all data (users, cards, tags, everything) for a fresh start |

## API Examples

**Create a card:**

```bash
curl -X POST http://localhost:8080/api/cards \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Implement login page",
    "description": "Build the login form with validation",
    "priority": 2,
    "column": "todo",
    "reporter_id": "<user-id>",
    "assignee_id": "<user-id>",
    "user_id": "<user-id>"
  }'
```

**Move a card to Done:**

```bash
curl -X PATCH http://localhost:8080/api/cards/<card-id>/move \
  -H "Content-Type: application/json" \
  -d '{"column": "done", "user_id": "<user-id>"}'
```

**Post a comment with @mention:**

```bash
curl -X POST http://localhost:8080/api/cards/<card-id>/comments \
  -H "Content-Type: application/json" \
  -d '{"content": "Hey @Alice, this is ready for review", "user_id": "<user-id>"}'
```

**Retry-safe create with idempotency:**

```bash
curl -X POST http://localhost:8080/api/cards \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: unique-request-id-123" \
  -d '{"title": "Safe to retry", "priority": 3, "column": "todo", "reporter_id": "<user-id>", "assignee_id": "<user-id>"}'
```

## Deployment

### Docker Compose (local development)

```bash
make run          # build and start everything
make logs         # tail logs
make stop         # stop containers
make clean        # stop and delete database volume
```

### Kubernetes / k3s

Deploy with a single command using the included kustomize manifests:

```bash
make deploy
```

This creates an `aiboard` namespace with a Postgres StatefulSet (1Gi PVC) and the app Deployment. To access the app locally:

```bash
make k8s-port-forward   # forwards localhost:8080 -> aiboard service
```

To pin a specific image version instead of `latest`:

```bash
cd k8s && kustomize edit set image ghcr.io/d6o/aiboard:1.2.0
kubectl apply -k k8s/
```

To expose publicly, uncomment the ingress in `k8s/kustomization.yaml` and set your hostname in `k8s/ingress.yaml`:

```yaml
# k8s/kustomization.yaml
resources:
  - namespace.yaml
  - postgres.yaml
  - app.yaml
  - ingress.yaml    # uncomment this line

# k8s/ingress.yaml — change the host
spec:
  rules:
    - host: board.yourdomain.com
```

To tear it all down:

```bash
make undeploy
```

### CI/CD

Every merge to `main` triggers the GitHub Actions release pipeline:

1. **Test** — builds and vets the Go code
2. **Release** — parses conventional commits, bumps the semver tag, and creates a GitHub Release with changelog
3. **Docker** — builds multi-arch images (`linux/amd64` + `linux/arm64`) and pushes to GHCR with semver tags (`1.2.3`, `1.2`, `1`, `latest`)

Use [Conventional Commits](https://www.conventionalcommits.org/) to control versioning:

| Prefix | Bump | Example |
|--------|------|---------|
| `feat:` | minor | `feat: add card search` |
| `fix:` | patch | `fix: subtask count off by one` |
| `feat!:` | major | `feat!: redesign API response format` |

## Architecture

```
cmd/server/          Entry point
internal/
  config/            Environment-based configuration
  database/          Connection, migrations, seed data
  model/             Domain types and error definitions
  store/             Database access (raw SQL, no ORM)
  service/           Business logic, validation, notifications
  handler/           HTTP handlers with JSON responses
  server/            Route registration and wiring
static/              Frontend (HTML + CSS + vanilla JS)
```

Each layer depends on the one below it through interfaces defined at the consumer side. Stores are injected into services, services into handlers. No globals, no init functions, no singletons.

## Make Targets

| Command | Description |
|---------|-------------|
| `make run` | Build and start with Docker Compose |
| `make stop` | Stop containers |
| `make restart` | Full rebuild and restart |
| `make logs` | Tail container logs |
| `make clean` | Stop containers and delete database volume |
| `make build` | Build without starting |
| `make deploy` | Deploy to Kubernetes with kustomize |
| `make undeploy` | Remove from Kubernetes |
| `make k8s-port-forward` | Forward localhost:8080 to the k8s service |
| `make k8s-logs` | Tail app logs in Kubernetes |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/aiboard?sslmode=disable` | PostgreSQL connection string |
| `PORT` | `8080` | HTTP server port |

## Seed Data

On first run, the database is populated with default **tags**: bug, feature, enhancement, urgent, design, backend, frontend (each with a distinct color).

No default users are created. Create users through the API or the UI before using the board.

To start completely fresh at any time, call `POST /api/board/reset`. This wipes all data including users, cards, tags, and activity. Default tags are re-seeded on the next server restart.
