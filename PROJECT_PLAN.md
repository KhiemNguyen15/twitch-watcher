# Twitch Watcher - Project Plan

## Overview
Twitch Watcher is a microservice-based, event-driven application that listens for Twitch events (e.g., new streams for specified games or streamers) and sends notifications to designated Discord servers via webhooks. It features a Discord-authenticated web frontend where users can configure which streams should trigger notifications.

---

## Architecture Summary

- **Architecture Type**: Microservices (Monorepo layout)
- **Event Coordination**: RabbitMQ
- **CI/CD**:
  - CI: GitHub Actions
  - CD: Keel (for automated deployments on K8s)
- **Deployment**: Kubernetes cluster (already provisioned and managed)
- **Frontend Framework**: Vite or Next.js
- **Backend Tools**: Go or Rust

---

## Tech Stack

| Component               | Technology      |
|-------------------------|-----------------|
| Frontend                | Vite or Next.js |
| Backend Services        | Go or Rust      |
| Authentication          | Discord OAuth2  |
| Message Broker          | RabbitMQ        |
| Relational Database     | PostgreSQL      |
| Caching / Rate Limiting | DragonflyDB     |

---

## Monorepo Structure

```
twitch-watcher/
│
├── .github/
│   └── workflows/           # GitHub Actions CI/CD workflows
│
├── infra/                   # Kubernetes manifests, Helm charts, secrets
│   ├── k8s/
│   ├── helm/
│   └── secrets/
│
├── services/                # Backend microservices
│   ├── stream-poller/
│   ├── auth-service/
│   ├── user-service/
│   └── notification-dispatcher/
│
├── frontend/                # React frontend app (Vite or Next.js)
│
├── libs/                    # Shared libraries and schemas
│   ├── types/
│   ├── utils/
│   └── db/
│
├── scripts/                 # Dev and deploy scripts
│
├── docker/                  # Shared Dockerfiles, docker-compose (if used)
│
├── .env.example             # Environment variable templates
├── README.md
└── PROJECT_PLAN.md
```

---

## Domain Overview

### 1. Stream Discovery

**Purpose**: Detect and monitor Twitch streams through polling integration.

**Responsibilities**:
- Poll Twitch API (`/helix/streams`) for stream activity.
- Manage rate limits and pagination.
- Emit `StreamStarted` events to RabbitMQ.

**Microservices**:
- `stream-poller`

---

### 2. Notification

**Purpose**: Deliver stream-related alerts to users through Discord webhooks.

**Responsibilities**:
- Subscribe to `StreamStarted` events.
- Query user preferences to find target Discord webhooks.
- Format and send messages via Discord.
- Handle retries or failures gracefully.

**Microservices**:
- `notification-dispatcher`

---

### 3. User Preferences

**Purpose**: Allow users to manage which games or streamers they want to be notified about.

**Responsibilities**:
- Store tracked streamers and games per user.
- Store Discord webhook configurations.
- Provide a CRUD API for managing preferences.
- Validate incoming data (e.g., valid Twitch IDs, Discord webhooks).

**Microservices**:
- `user-preferences`

---

### 4. Authentication

**Purpose**: Authenticate and authorize users accessing the system.

**Responsibilities**:
- Authenticate users via Discord OAuth or a third-party service (e.g., Clerk).
- Manage tokens or session validation.
- Secure access to user-specific data and preferences.

**Microservices**:
- `auth-service` or integration with Clerk/Auth0/Firebase Auth

---

### 5. Web Dashboard (Optional BFF)

**Purpose**: Frontend UI for users to manage their configuration.

**Responsibilities**:
- Display tracking status and logs.
- Allow users to manage watchlist and Discord webhooks.
- Authenticate users and secure access.
- Use GraphQL or REST to proxy backend services.

**Microservices**:
- `graphql-api` (BFF)
- `frontend-app` (React, Vite, etc.)

---

### 6. Event Bus (Infrastructure Domain)

**Purpose**: Enable asynchronous communication between services.

**Responsibilities**:
- Route events like `StreamStarted`, `StreamEnded`, `DeliveryFailed`.
- Allow microservices to publish/subscribe without tight coupling.
- Ensure reliability and delivery guarantees.

**Technology**:
- RabbitMQ

---

### Domain-to-Service Mapping Summary

| Domain             | Purpose                       | Microservice(s)                      |
|--------------------|-------------------------------|--------------------------------------|
| Stream Discovery   | Detect live streams           | `stream-poller`                      |
| Notification       | Deliver alerts                | `notification-dispatcher`            |
| User Preferences   | Manage user watchlists        | `user-service`                       |
| Authentication     | User auth and access control  | `auth-service` or 3rd-party auth     |
| Web Dashboard      | UI and frontend gateway       | `graphql-api`, `frontend-app`        |
| Event Bus          | Event communication layer     | RabbitMQ (infrastructure component)  |

---

## Core Microservices

1. **Twitch Listener**
   - Polls Twitch API for new stream events
   - Validates & forwards events into RabbitMQ

2. **Notification Dispatcher**
   - Listens to RabbitMQ
   - Sends Discord webhook messages
   
3. **User Service**
   - Stores user configuration (e.g., target game/streamer, Discord webhook)
   - Manages webhook creation/deletion

4. **Auth Service**
   - Manages Discord OAuth2 login flow
   - Persists user tokens & validates permissions

5. **Frontend UI**
   - React-based interface
   - Allows users to:
     - Log in via Discord
     - Create/edit/delete webhook subscriptions

---

## Feature List

- [ ] Login via Discord
- [ ] View/manage existing webhooks
- [ ] Subscribe to Twitch streams by streamer or game
- [ ] Discord webhook notification setup
- [ ] Real-time event propagation via RabbitMQ

---

## Database Design (PostgreSQL)

### Users Table
- id (UUID)
- discord_user_id (string)
- access_token (string)
- refresh_token (string)
- created_at (timestamp)

### Webhooks Table
- id (UUID)
- user_id (FK to Users)
- discord_webhook_url (string)
- is_active (boolean)
- created_at (timestamp)

### Subscriptions Table
- id (UUID)
- webhook_id (FK to Webhooks)
- type (enum: 'streamer', 'game')
- target_id (string)
- twitch_subscription_id (string)
- created_at (timestamp)

---

## CI/CD

### GitHub Actions (CI)
- Lint, test, and build microservices
- Build/push Docker images
- Trigger builds only on relevant path changes via `paths:` config

### Keel (CD)
- Watches container registries
- Auto-deploys updated services to Kubernetes

---

## Deployment

- All services deployed on K3s-based Kubernetes cluster
- Internal communication via RabbitMQ (brokered pub/sub)
- Services exposed via Ingress and Traefik
- Secrets managed via Kubernetes secrets (or Vault in future)

