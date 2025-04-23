# Twitch Watcher - Project Plan

## Overview
Twitch Watcher is a microservice-based, event-driven application that listens for Twitch EventSub events (e.g., new streams for specified games or streamers) and sends notifications to designated Discord servers via webhooks. It features a Discord-authenticated web frontend where users can configure which streams should trigger notifications.

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
│   ├── twitch-listener/
│   ├── webhook-manager/
│   ├── auth-service/
│   └── notification-sender/
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

## Core Microservices

1. **Twitch Listener**
   - Handles Twitch EventSub webhook callbacks
   - Validates & forwards events into RabbitMQ

2. **Notification Dispatcher**
   - Listens to RabbitMQ
   - Sends Discord webhook messages
   
3. **Webhook Manager**
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
- [ ] Persistent Twitch EventSub subscriptions
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

