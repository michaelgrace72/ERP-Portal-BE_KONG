# 🌐 Portal Service (Backend)

> **Multi-Tenant ERP System with Kong API Gateway & Phantom Token Authentication**

[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/docker-ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![Architecture](https://img.shields.io/badge/architecture-clean-brightgreen?style=flat)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
[![License](https://img.shields.io/badge/license-MIT-yellow?style=flat)](./LICENSE)

This repository hosts the **Portal Backend Service**, the core identity and access management system for the ERP suite. It implements **Clean Architecture** principles and leverages **Kong API Gateway** for secure, scalable multi-tenancy using the **Phantom Token Pattern**.

---

## 📑 Table of Contents

- [🚀 Key Features](#-key-features)
- [🏗️ Architecture Overview](#️-architecture-overview)
- [📁 Project Structure](#-project-structure)
- [🛠️ Environment 1: Local Development](#️-environment-1-local-development)
  - [Option A: Hybrid Mode (Recommended)](#option-a-hybrid-mode-recommended-for-coding-)
  - [Option B: Pure Docker Mode](#option-b-pure-docker-mode-secure--simulation)
- [🚀 Environment 2: Server / Production](#-environment-2-server--production)
- [📊 Port Allocation Reference](#-port-allocation-reference)
- [📚 API Documentation](#-api-documentation)
- [🔧 Operational Guide & Commands](#-operational-guide--commands)

---

## 🚀 Key Features

| Feature                   | Description                                                                                                            |
| :------------------------ | :--------------------------------------------------------------------------------------------------------------------- |
| **🔐 Phantom Token Auth** | Secure OAuth2-style authentication where opaque tokens are exchanged for JWTs at the gateway level (Kong).             |
| **🏢 Multi-Tenancy**      | Built-in support for multiple tenants with strict data isolation and Role-Based Access Control (RBAC).                 |
| **🏛️ Clean Architecture** | Code is organized into independent layers (Entity, Repository, UseCase, Delivery) for maintainability and testability. |
| **⚡ High Performance**   | Powered by Go (Golang) and Gin Web Framework, optimized for speed and concurrency.                                     |
| **🐳 Docker First**       | "Zero-dependency" setup. The entire stack (DB, Cache, Gateway, MQ) runs in Docker containers.                          |

---

## 🏗️ Architecture Overview

The system follows a strict flow where specific responsibilities are delegated to the Gateway (Infrastructure) and the Service (Business Logic).

```
┌──────────────────────────────────────────────────────────┐
│                  Client Applications                     │
│           Web App, Mobile App, Third-party APIs          │
└──────────────────────────────────────────────────────────┘
                            ▲
                            │ HTTP/HTTPS
                            ▼
┌──────────────────────────────────────────────────────────┐
│                   Kong API Gateway                       │
│     • Token Validation (JWT Plugin)                      │
│     • Rate Limiting                                      │
│     • Request/Response Transformation                    │
│     • Consumer Management                                │
└──────────────────────────────────────────────────────────┘
                            ▲
                            │ Proxied Requests
                            ▼
┌──────────────────────────────────────────────────────────┐
│                Portal Service (Go + Gin)                 │
│                                                          │
│   ┌──────────────────────────────────────────────┐       │
│   │         Delivery Layer                       │       │
│   │   HTTP Handlers, Middleware, Routes          │       │
│   └──────────────────────────────────────────────┘       │
│                         ▲                                │
│   ┌──────────────────────────────────────────────┐       │
│   │         Use Case Layer                       │       │
│   │   • Registration (User + Tenant + Kong)      │       │
│   │   • User Management                          │       │
│   └──────────────────────────────────────────────┘       │
│                         ▲                                │
│   ┌─────────────────┬──────────────────────────┐         │
│   │  Repository     │   Gateway Layer           │        │
│   │  • User         │   • Kong Admin Client     │        │
│   │  • Tenant       │   • JWT, BCrypt, OAuth    │        │
│   │  • Membership   │   • Redis, RabbitMQ       │        │
│   └─────────────────┴──────────────────────────┘         │
│                         ▲                                │
│   ┌──────────────────────────────────────────────┐       │
│   │         Entity Layer                         │       │
│   │   User, Tenant, Role, Membership             │       │
│   └──────────────────────────────────────────────┘       │
└──────────────────────────────────────────────────────────┘
                            ▲
                            │
┌──────────────────────────────────────────────────────────┐
│              Data & Infrastructure Layer                 │
│  PostgreSQL • Redis • RabbitMQ • Kong Database           │
└──────────────────────────────────────────────────────────┘
```

<details>
<summary><strong>👇 Click to expand Mermaid Flowchart</strong></summary>

```mermaid
graph TD
    Client[📱 Client App\nWeb / Mobile / 3rd Party] -->|1. HTTP Request + Opaque Token| Kong

    subgraph "Infrastructure Layer"
        Kong[🦍 Kong API Gateway\n(Port 3600)]
        Redis[(🔴 Redis Cache\nSessions & Tokens)]
    end

    Kong -->|2. Validate Opaque Token| Redis
    Kong -->|3. Introspect & Swap for JWT| Portal

    subgraph "Application Layer"
        Portal[🟦 Portal Service\n(Port 3502)]
        PG[(🐘 PostgreSQL\nIdentity Data)]
        RabbitMQ[(🐇 RabbitMQ\nAsync Events)]
    end

    Portal -->|4. Business Logic| PG
    Portal -->|5. Publish Events| RabbitMQ
```

</details>

### 🧩 Architectural Decisions

1.  **Separation of Concerns**:
    - **Kong**: Handles SSL termination, Rate Limiting, and Authentication (Token Validation).
    - **Portal Service**: Handles Business Rules, Authorization (Permissions), and Data persistence.
2.  **Dependency Inversion**: Inner layers (Entities) know nothing about outer layers (Database/HTTP).
3.  **Zero-Trust**: The database is **internal only** within the Docker network (production). Services communicate via governed APIs.

---

## 📁 Project Structure

A quick map of where code lives:

```bash
go-gin-clean/
├── cmd/                                    # 🚀 Entrypoints
│   ├── server/main.go                     #    -> Main HTTP Server
│   └── migrate/main.go                    #    -> Migration CLI Tool
├── internal/                              # 🔒 Private Application Code
│   ├── entity/                            #    🏛️ Domain Models (User, Tenant)
│   ├── repository/                        #    💾 Data Access (Postgres implementation)
│   ├── usecase/                           #    🎯 Business Logic (Registration, Login flow)
│   ├── delivery/                          #    📤 Transport Layer (HTTP Handlers, Routes)
│   ├── gateway/                           #    🌐 External Services (Kong Client, Redis, RabbitMQ)
│   └── container/                         #    🔧 Dependency Injection (IoC)
├── pkg/                                   # 📦 Shared Public Packages (Config, Errors, Utils)
├── migrations/                            # 📜 SQL Database Migrations
├── docker-compose.local.yml               # 🛠️ Local Dev Infrastructure Definition
└── Makefile                               # ⚡ Automation Scripts
```

---

## 🛠️ Environment 1: Local Development

We support two workflows. **Option A** is best for daily coding. **Option B** is best for integration testing.

### Option A: Hybrid Mode (Recommended for Coding) 🔥

> **Why?** Run infrastructure in Docker, but run the Go app **natively** on your machine. This enables **Air (Hot Reload)** and faster debugging.

#### 1. Configure `.env`

Enable `localhost` access for connections:

```bash
cp .env.example .env
```

Edit `.env` to point to localhost:

```ini
DB_HOST=localhost
REDIS_HOST=localhost
RABBITMQ_HOST=localhost
```

#### 2. Open Docker Ports

Modify `docker-compose.local.yml`. **Uncomment** the ports for `portal-db` to allow your host machine to connect:

```yaml
portal-db:
  # ...
  ports:
    - "5432:5432" # ✅ UNCOMMENT THIS
```

#### 3. Start Infrastructure Only

Start the backing services (_excluding_ the app container to avoid port conflict):

```bash
make infra
```

#### 4. Initialize Database

Since the DB is exposed to `localhost:5432`, run migrations directly from your terminal:

```bash
# Apply Migrations
go run cmd/migrate/main.go up

# Seed Initial Data (Roles, System Tenant)
go run cmd/seed/main.go
```

#### 5. Run App with Hot Reload

```bash
# Install Air (once)
go install github.com/cosmtrek/air@latest

# Run App
air
```

---

### Option B: Pure Docker Mode (Secure / Simulation)

> **Why?** Simulation of production environment. No Go installation required on host. Database is secure/internal.

#### 1. Configure `.env`

Use service names for connections:

```bash
cp .env.example .env
```

Edit `.env`:

```ini
DB_HOST=portal-db    # 🔒 Access by container name
REDIS_HOST=redis
```

#### 2. Start Everything

```bash
make up
```

#### 3. Manage Database (via Docker)

Since `localhost:5432` is closed/secure, use our Docker helpers:

```bash
# Run Migrations
docker-compose -f docker-compose.local.yml up portal-migrations

# Seed Data
docker-compose -f docker-compose.local.yml up portal-seeder
```

---

## 🚀 Environment 2: Server / Production

**Goal**: Zero-dependency deployment. The server only needs Docker installed.

### 1. Configuration

Set production environment variables in `.env`:

```ini
APP_ENV=production
gin_mode=release
DB_HOST=postgres       # Matches service name in docker-compose.yml
JWT_SECRET=<strong_random_key>
```

### 2. Build & Deploy

```bash
# Builds optimized containers and starts them
docker-compose up -d --build
```

### 3. Initialization (One-time)

```bash
# 1. Migrate DB
docker-compose run --rm portal-service /app/migrate up

# 2. Seed DB
docker-compose run --rm portal-service /app/seed

# 3. Configure Kong Gateway
./scripts/setup-kong.sh
```

---

## 📊 Port Allocation Reference

### 🏭 ERP Application Services (3500-3599)

| Service             | Port   | Backup Port | Description                          |
| :------------------ | :----- | :---------- | :----------------------------------- |
| **Portal Frontend** | `3500` | `3501`      | Main User Interface                  |
| **Portal Backend**  | `3502` | `3503`      | **(This Service)** Auth & Tenant API |
| ERP Frontend        | `3510` | `3511`      | Main ERP Dashboard                   |
| ERP Inventory       | `3512` | `3513`      | Inventory Management                 |
| ERP Manufacture     | `3514` | `3515`      | Manufacturing Service                |
| ERP General Ledger  | `3516` | `3517`      | General Ledger Service               |
| SCM Frontend        | `3540` | -           | Supply Chain UI                      |

### 🛠️ Infrastructure & Tools (3600-3699)

| Service              | Port   | Protocol | Description                            |
| :------------------- | :----- | :------- | :------------------------------------- |
| **Kong API Gateway** | `3600` | HTTP     | **Main Entry Point** (Public API)      |
| Kong Admin API       | `3602` | HTTP     | Configuration API (Private)            |
| **PgAdmin**          | `5050` | HTTP     | Database GUI (User: `admin@admin.com`) |
| **RabbitMQ UI**      | `3608` | HTTP     | Message Queue Dashboard                |
| **Redis**            | `3606` | TCP      | Cache Store                            |
| **PostgreSQL**       | `5432` | TCP      | Main Database                          |

### 🟢 Running Service Status (Current Docker Stack)

_Based on `docker-compose ps`_

| Service Name     | Container Name    | Status     | Host Port      | Internal Port |
| :--------------- | :---------------- | :--------- | :------------- | :------------ |
| `kong`           | `kong-gateway`    | ✅ Healthy | **3600**, 3602 | 8000, 8001    |
| `kong-database`  | `kong-postgres`   | ✅ Healthy | 3609           | 5432          |
| `portal-db`      | `portal-postgres` | ✅ Healthy | _(Internal)_   | 5432          |
| `pgadmin`        | `portal-pgadmin`  | ✅ Up      | **5050**       | 80            |
| `portal-service` | `portal-service`  | ✅ Up      | **3502**       | 3000          |
| `redis`          | `portal-redis`    | ✅ Healthy | 3606           | 6379          |
| `rabbitmq`       | `portal-rabbitmq` | ✅ Healthy | 3607, **3608** | 5672, 15672   |

---

## 📚 API Documentation

### 🏥 Health Check

- `GET /health` - Service health status (Public)

### 🔐 Authentication (`/api/v1/auth`)

| Method   | Endpoint               | Description                  |
| :------- | :--------------------- | :--------------------------- |
| **POST** | `/login`               | Login with email/password    |
| **POST** | `/register`            | Register new user & tenant   |
| **POST** | `/refresh-token`       | Refresh expired access token |
| **POST** | `/verify-email`        | Verify email token           |
| **POST** | `/send-reset-password` | Request password reset       |
| **POST** | `/reset-password`      | Complete password reset      |
| **POST** | `/oauth2/url`          | Get OAuth2 Login URL         |

**Protected (Token Required):**

- `POST /logout` - Revoke session/token
- `POST /select-tenant` - Context switch to another tenant
- `GET  /session` - Get current session info

### 👤 Profile (`/api/v1/profile`)

- `GET  /` - Get My Profile
- `PUT  /` - Update Profile Info
- `PUT  /change-password` - Update Password

### 👥 User Management (`/api/v1/users`)

- `GET  /me` - Get detailed user info with roles
- `GET  /` - List all users (Admin)
- `POST /` - Create new user (Admin)
- `PUT  /:code` - Update user details
- `PUT  /:code/change-status` - Activate/Suspend user
- `DELETE /:code` - Soft delete user

### 🏢 Tenant & Membership (`/api/v1`)

- `POST /memberships` - Add user to tenant
- `GET  /tenants/:id/members` - List members of tenant
- `PUT  /tenants/:id/roles` - Update member roles

> **Note**: All protected endpoints require header: `Authorization: Bearer <access_token>`

---

## 🔧 Operational Guide & Commands

### 🔌 Service Access URLs

| Service        | Access URL                                                   | Default Credentials         |
| :------------- | :----------------------------------------------------------- | :-------------------------- |
| **Kong Proxy** | [http://localhost:3600](http://localhost:3600)               | -                           |
| **App Health** | [http://localhost:3502/health](http://localhost:3502/health) | -                           |
| **PgAdmin**    | [http://localhost:5050](http://localhost:5050)               | `admin@admin.com` / `admin` |
| **RabbitMQ**   | [http://localhost:3608](http://localhost:3608)               | `guest` / `guest`           |

### 💻 Common CLI Commands

**Docker Management**

```bash
make infra      # Start Core Infra (DB, Kong, Redis, RabbitMQ)
make up         # Start Full Stack (Infra + App)
make down       # Stop & Remove Containers
make logs       # Follow logs for all containers
```

**Database Migrations (Host Mode)**

```bash
make migrate-up          # Apply all migrations
make migrate-down        # Rollback last step
make migrate-create NAME=x # Create new migration file
```

**Database Migrations (Docker Mode)**

```bash
# Secure execution inside container network
docker-compose -f docker-compose.local.yml up portal-migrations
```

**Development**

```bash
make run        # Run App (go run ...)
make build      # Build Binary
make test       # Run Tests
air             # Start with Hot Reload
```

---

## 📜 License

MIT License - Internal Use Only for ERP System.
