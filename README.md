# Twitch Watcher

Monitors Twitch and sends Discord webhook alerts when new streams go live for subscribed games or streamers.

## Architecture

```
User ──► subscription-service ──► PostgreSQL
               │
               │ GET /internal/subscriptions/active
               ▼
         stream-poller  ──► twitch.streams.raw (NATS JetStream)
               │
               ▼
         stream-filter  ──► twitch.streams.new  (NATS JetStream)
               │
               ▼
      notification-dispatcher ──► Discord webhook
```

---

## Configuration

Each service is configured entirely through environment variables.

### subscription-service

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | ✅ | — | PostgreSQL connection string, e.g. `postgres://user:pass@host:5432/dbname` |
| `INTERNAL_API_KEY` | ✅ | — | Shared secret used by stream-poller to call the internal endpoint |
| `HTTP_ADDR` | | `:8080` | Address the HTTP server listens on |

### stream-poller

| Variable | Required | Default | Description |
|---|---|---|---|
| `TWITCH_CLIENT_ID` | ✅ | — | Twitch application Client ID |
| `TWITCH_CLIENT_SECRET` | ✅ | — | Twitch application Client Secret |
| `INTERNAL_API_KEY` | ✅ | — | Must match the value set in subscription-service |
| `SUBSCRIPTION_SVC_URL` | | `http://localhost:8080` | Base URL of subscription-service |
| `NATS_URL` | | `nats://localhost:4222` | NATS server URL |
| `VALKEY_ADDR` | | `localhost:6379` | Valkey/Redis address for game ID cache |
| `POLL_INTERVAL_SECONDS` | | `60` | How often to poll the Twitch API |

### stream-filter

| Variable | Required | Default | Description |
|---|---|---|---|
| `NATS_URL` | | `nats://localhost:4222` | NATS server URL |
| `VALKEY_ADDR` | | `localhost:6379` | Valkey/Redis address for seen-stream deduplication |

### notification-dispatcher

| Variable | Required | Default | Description |
|---|---|---|---|
| `NATS_URL` | | `nats://localhost:4222` | NATS server URL |

---

## Local Development

### 1. Start infrastructure

```bash
docker compose up -d   # requires a compose file — see below
```

Or start each dependency manually:

```bash
# NATS with JetStream
docker run -d --name nats -p 4222:4222 nats:latest -js

# Valkey
docker run -d --name valkey -p 6379:6379 valkey/valkey:latest

# PostgreSQL
docker run -d --name postgres -p 5432:5432 \
  -e POSTGRES_USER=twitch_watcher \
  -e POSTGRES_PASSWORD=secret \
  -e POSTGRES_DB=twitch_watcher \
  postgres:16-alpine
```

### 2. Run the database migration

```bash
psql postgres://twitch_watcher:secret@localhost:5432/twitch_watcher \
  -f services/subscription-service/migrations/001_create_subscriptions.up.sql
```

### 3. Export environment variables

```bash
# subscription-service
export DATABASE_URL="postgres://twitch_watcher:secret@localhost:5432/twitch_watcher"
export INTERNAL_API_KEY="local-dev-secret"

# stream-poller (in a separate shell)
export TWITCH_CLIENT_ID="<your-client-id>"
export TWITCH_CLIENT_SECRET="<your-client-secret>"
export INTERNAL_API_KEY="local-dev-secret"

# stream-filter and notification-dispatcher use defaults; nothing extra required
```

### 4. Run each service

```bash
make run-subscription-service
make run-stream-poller
make run-stream-filter
make run-notification-dispatcher
```

---

## Kubernetes (Helm)

### Prerequisites

- Namespace: `kubectl apply -f deploy/namespaces/twitch-watcher.yaml`
- Infrastructure:
  ```bash
  helm upgrade --install nats nats/nats \
    -n twitch-watcher -f deploy/infrastructure/nats/values.yaml

  helm upgrade --install valkey bitnami/valkey \
    -n twitch-watcher -f deploy/infrastructure/valkey/values.yaml

  helm upgrade --install postgresql bitnami/postgresql \
    -n twitch-watcher -f deploy/infrastructure/postgresql/values.yaml
  ```

### Required Secrets

Create these secrets **before** installing the service charts.

#### `twitch-credentials`
```bash
kubectl create secret generic twitch-credentials \
  -n twitch-watcher \
  --from-literal=TWITCH_CLIENT_ID=<your-client-id> \
  --from-literal=TWITCH_CLIENT_SECRET=<your-client-secret>
```

#### `internal-api-key`
```bash
kubectl create secret generic internal-api-key \
  -n twitch-watcher \
  --from-literal=INTERNAL_API_KEY=<random-secret>
```

#### `postgresql-secret`
```bash
kubectl create secret generic postgresql-secret \
  -n twitch-watcher \
  --from-literal=DATABASE_URL="postgres://twitch_watcher:<pass>@postgresql:5432/twitch_watcher" \
  --from-literal=postgres-password=<admin-pass> \
  --from-literal=password=<user-pass>
```

#### `valkey-secret`
```bash
kubectl create secret generic valkey-secret \
  -n twitch-watcher \
  --from-literal=password=<valkey-password> \
  --from-literal=VALKEY_ADDR="valkey-primary:6379"
```

### Install service charts

```bash
helm upgrade --install subscription-service deploy/helm/subscription-service \
  -n twitch-watcher --set image.tag=<version>

helm upgrade --install stream-poller deploy/helm/stream-poller \
  -n twitch-watcher --set image.tag=<version>

helm upgrade --install stream-filter deploy/helm/stream-filter \
  -n twitch-watcher --set image.tag=<version>

helm upgrade --install notification-dispatcher deploy/helm/notification-dispatcher \
  -n twitch-watcher --set image.tag=<version>
```

---

## Obtaining Twitch API credentials

1. Go to [dev.twitch.tv/console](https://dev.twitch.tv/console) and create an application.
2. Set the OAuth redirect URL to `http://localhost` (not used, but required).
3. Copy the **Client ID** and generate a **Client Secret**.

The app uses the [Client Credentials flow](https://dev.twitch.tv/docs/authentication/getting-tokens-oauth/#client-credentials-grant-flow) — no user login required.

---

## API Reference

### Public endpoints

```
POST   /v1/subscriptions
       Body: { "discord_webhook": "https://discord.com/api/webhooks/...",
               "watch_type": "game" | "streamer",
               "watch_target": "Fortnite" | "ninja" }

GET    /v1/subscriptions/{id}

DELETE /v1/subscriptions/{id}

GET    /v1/health
```

### Internal endpoint (stream-poller only)

```
GET    /internal/subscriptions/active
       Header: X-Internal-API-Key: <INTERNAL_API_KEY>
```
