# Twitch Watcher - Stream Poller Microservice Plan

## Overview
The `stream-poller` microservice is responsible for detecting when Twitch streams go live for specific games or streamers. It polls the Twitch API at regular intervals, detects new stream activity, and publishes structured event messages to RabbitMQ. Active streams are tracked via DragonflyDB for change detection.

---

## Responsibilities
- Poll Twitch API for active streams by `game_id` and `user_id`
- Batch poll requests in groups of up to 100 IDs to reduce request volume
- Handle Twitch API pagination
- Manage concurrency using goroutines with rate-limited throttling
- Track currently active streams using DragonflyDB
- Detect new stream starts and publish events to RabbitMQ
- Authenticate to Twitch using an App Access Token (client_credentials grant)
- Include a `/healthz` endpoint for liveness checks
- Build as a Docker container for deployment

---

## Polling Strategy
- Uses `GET https://api.twitch.tv/helix/streams` endpoint
- Batches `user_id` and `game_id` requests in chunks of 100
- Concurrent polling via goroutines and `sync.WaitGroup`
- Rate-limiting with a semaphore (max ~13 requests/sec to respect 800 req/min limit)
- Pagination support using `pagination.cursor`
- Configurable polling interval via environment variable (e.g., every 30 seconds)

---

## Change Detection
- Uses DragonflyDB (Redis-compatible) for stream tracking
- For each poll, cache active streamers per `game_id` or `user_id`
- Compare new list with previous set stored in DragonflyDB
- Difference = streamers who just went live
- New streamers trigger a message published to RabbitMQ

---

## RabbitMQ Publishing
- Queue: `stream.online`
- Message format: JSON conforming to internal schema
```json
{
  "user_id": "123456",
  "user_login": "streamername",
  "game_id": "7890",
  "title": "Going Live Now!",
  "started_at": "2025-04-23T16:00:00Z"
}
```
- RabbitMQ handles retries and dead-lettering

---

## Authentication
- Uses Twitch App Access Token via `client_credentials` grant
- Validates token before each batch request
- Automatically renews token on expiration (401 response)

---

## Endpoints
| Path       | Method | Purpose         |
|------------|--------|-----------------|
| `/healthz` | GET    | Liveness check  |

---

## Folder Structure
```
stream-poller/
├── cmd/
│   └── stream-poller/          # Main application entrypoint
│       └── main.go
├── internal/
│   ├── poller/                 # Game and user poll logic
│   │   ├── game.go
│   │   └── user.go
│   ├── twitch/                 # Twitch API client + auth
│   │   ├── client.go
│   │   └── auth.go
│   ├── publisher/              # RabbitMQ publisher
│   │   └── rabbitmq.go
│   ├── storage/                # DragonflyDB (Redis) access
│   │   └── dragonfly.go
│   └── config/                 # Config loading and validation
│       └── env.go
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

---

## Configuration (.env)
```env
POLL_INTERVAL_SECONDS=30
TWITCH_CLIENT_ID=your_client_id
TWITCH_CLIENT_SECRET=your_client_secret
DRAGONFLY_URL=redis://localhost:6379
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

---

## Future Enhancements
- Prometheus metrics integration
- Optional Redis-based deduplication layer for rapid deployments
- Dynamic target management via webhook-manager or user UI
- Alerting for API failures or poll failures

