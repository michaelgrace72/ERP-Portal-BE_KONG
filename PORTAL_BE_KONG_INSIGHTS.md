# 🧠 Portal-BE_Kong System Context & Insights

## 📋 Executive Summary

**Portal-BE_Kong** is the central identity and access management (IAM) backend for the ERP suite. It is designed with **Clean Architecture** in Go (Gin framework) and uses **Kong API Gateway** as the enforcement point for security. It implements the **Phantom Token Pattern**, exchanges opaque tokens for JWTs at the gateway level, and handles multi-tenancy with strict data isolation.

---

## 🏗️ Architecture & Design Patterns

### 1. Clean Architecture Layers

The project adheres to strict separation of concerns:

- **Entity Layer** (`internal/entity`): Pure domain models (User, Tenant, Role). No external dependencies.
- **Repository Layer** (`internal/repository`): Data access implementation (PostgreSQL/Gorm).
- **Use Case Layer** (`internal/usecase`): Application specific business rules. Connects Repositories, Gateways, and Services.
- **Delivery Layer** (`internal/delivery/http`): HTTP handlers and routing.
- **Gateway/Infrastructure** (`internal/gateway`, `internal/infrastructure`): Interfaces to external systems (Kong, Redis, RabbitMQ, Cloudinary).

### 2. The Phantom Token Pattern

This is a critical security architectural decision:

1.  **Client** sends credentials to `POST /auth/phantom-login`.
2.  **Portal** verifies credentials, generates a **Random Reference Token (Opaque)**, saves session data in **Redis** with this token as key, and returns the opaque token to Client.
3.  **Client** makes API request with `Authorization: Bearer <opaque_token>`.
4.  **Kong** intercepts the request. It calls the **Introspection Endpoint** (`POST /auth/introspect`) on the Portal.
5.  **Portal** validates the opaque token against Redis, retrieves the user profile & claims, and returns a JSON object.
6.  **Kong** (using a plugin) converts this JSON into a signed **JWT** and injects it into the upstream request to the services.
7.  **Services** (including Portal itself) only see and validate the JWT.

### 3. Multi-Tenancy

- **Structure**: Users belong to **Tenants** via **Memberships**.
- **Context Switching**: Users can switch active tenants via `POST /auth/select-tenant`.
- **Authorization**: RBAC (Role-Based Access Control) is applied per tenant.

---

## 🛠️ Infrastructure Stack

| Service            | Role        | Tech Details                                                           |
| :----------------- | :---------- | :--------------------------------------------------------------------- |
| **Portal Service** | Core Logic  | Go 1.22+, Gin, Gorm                                                    |
| **Kong Gateway**   | API Gateway | Routes traffic, handles auth (JWT/Phantom), Rate limiting              |
| **PostgreSQL**     | Primary DB  | Stores Users, Tenants, Roles, Memberships                              |
| **Redis**          | Fast Cache  | Stores **Sessions** (Opaque Tokens), cached user permissions           |
| **RabbitMQ**       | Event Bus   | Publishes `UserCreated`, `TenantUpdated` events for other ERP services |

---

## 🔌 API & Logic Breakdown

### 🔐 Authentication (`internal/usecase/auth_usecase.go`)

- **`POST /api/v1/auth/phantom-login`**:
  - Validates email/password (Bcrypt).
  - Checks if user is Active.
  - Generates opaque string (Reference Token).
  - Stores `User`, `ActiveTenant`, `Permissions` in Redis (TTL ~30m).
  - Returns Reference Token.
- **`POST /api/v1/auth/introspect`**:
  - **CRITICAL**: Called by Kong, not users.
  - Lookup token in Redis.
  - Returns payload: `{ "active": true, "sub": "user_id", "exp": 123, "scope": "..." }`.

### 📝 Registration (`internal/usecase/registration_usecase.go`)

- **`POST /api/v1/auth/register`**:
  - **Transaction**:
    1. Create `User` in Postgres.
    2. Create a default `Tenant` (Organization) for the user.
    3. Create `Membership` (Admin role) for that tenant.
  - **Kong Integration**:
    - Creates a "Consumer" in Kong for the new user (via `kongClient`).
    - Ensures rate-limiting tiers are applied.

### 👥 User Management (`internal/usecase/user_management_usecase.go`)

- **`POST /api/v1/users`** (Admin Create):
  - Creates a user (status: Pending/Invited).
  - Sends invitation email (mock or real).
- **`GET /api/v1/users/me`**:
  - Returns detailed profile + all memberships + current active tenant roles.

### 🔄 Data & Event Flow

1.  **Write Operations** (Create User/Tenant) -> Write to Postgres -> Publish Event to RabbitMQ (`user.created`).
2.  **Read Operations** -> Check Redis Cache -> Fallback to Postgres.

---

## � Frontend Implementation Guide

This section details the specific flows the frontend must implement to interact with the Portal Backend.

### 1. Authentication Flow

The system uses **Phantom Tokens**. The frontend holds an opaque string (Reference Token) which is useless without the backend session.

#### **A. Login Process**

1.  **Request**: `POST /api/v1/auth/phantom-login`
    ```json
    {
    	"email": "user@example.com",
    	"password": "secret_password"
    }
    ```
2.  **Scenario 1: Single Tenant (or Tenant Specified)**

    - **Response (200 OK)**:
      ```json
      {
      	"status": "success",
      	"message": "Login successful",
      	"data": {
      		"access_token": "ref_...",
      		"expires_in": 1800,
      		"token_type": "Bearer",
      		"user": { "id": 1, "email": "..." },
      		"tenant": { "id": 10, "name": "Tech Corp" }
      	}
      }
      ```
    - **Action**: Store `access_token`. Redirect to Dashboard.

3.  **Scenario 2: Multi-Tenant (Ambiguous)**

    - **Response (200 OK - Note: It is 200, not 300)**:
      ```json
      {
      	"status": "success",
      	"message": "Tenant selection required",
      	"data": {
      		"message": "User has multiple tenants...",
      		"requires_choice": true,
      		"tenants": [
      			{ "id": 10, "name": "Company A", "role": "Administrator" },
      			{ "id": 20, "name": "Company B", "role": "Member" }
      		]
      	}
      }
      ```
    - **Action**: Present a "Select Organization" screen to the user.

4.  **Scenario 3: Completing Selection**
    - User picks a tenant.
    - **Request**: `POST /api/v1/auth/phantom-login` (or `/auth/select-tenant` if already logged in context? - No, use Login for initial auth)
    - _Correction_: `Login` endpoint handles tenant selection if `tenant_id` is passed.
    ```json
    {
    	"email": "user@example.com",
    	"password": "secret_password",
    	"tenant_id": 20
    }
    ```

### 2. Multi-Tenancy & Context Switching

#### **Key Concept: Token = Tenant Context**

A token is strictly bound to **one specific tenant**. You cannot use the same token to access data from two different tenants.

#### **Switching Tenants**

Since the token is bound to a tenant, "switching" tenants effectively means **logging in again** to the target tenant.

- **Endpoint**: `POST /api/v1/auth/select-tenant`
- **Payload**: Requires `email`, `password`, and `tenant_id`.
- **UX Implication**: The "Switch Tenant" feature might need to prompt for a password confirmation for security, unless the frontend application temporarily caches the password (not recommended) or the user is just directed to the "Select Organization" screen.

### 3. Authorization (RBAC) Implementation

The Frontend needs to know _what_ to show/hide. The Login response gives you the _Token_, but not the _Permissions_.

#### **Recommended Bootstrap Flow**

After a successful login:

1.  **Call**: `GET /api/v1/users/me`
2.  **Headers**: `Authorization: Bearer <access_token>`
3.  **Response**:
    ```json
    {
    	"data": {
    		"id": 1,
    		"name": "User Name",
    		"memberships": [
    			{
    				"tenant_id": 10,
    				"tenant_name": "Tech Corp",
    				"role_name": "Administrator",
    				"permissions": ["user:create", "report:view"] // <--- USE THIS
    			}
    		]
    	}
    }
    ```
4.  **Logic**:
    - Find the `membership` in the response where `tenant_id` matches the current active tenant (from login response).
    - Load the `permissions` array into your Frontend State (e.g., Redux/Context/Zustand).
    - **Guard Components**: `<Can permission="user:create"> <CreateUserButton /> </Can>`

### 4. Handling 401 Unauthorized

Since tokens expire (default 30 mins) or can be revoked:

1.  **Intercept 401** responses.
2.  **Try Refresh**: Call `POST /api/v1/auth/refresh` with the current token.
3.  **Success**: Retry original request.
4.  **Fail**: Redirect to Login.

---

## �📂 Key File Map

Use this to locate logic quickly:

- **Entry Point**: `cmd/server/main.go` -> Sets up container, database, and router.
- **Router**: `internal/delivery/http/route/route.go` -> Defines all API paths and middleware.
- **Dependency Injection**: `internal/infrastructure/container.go` -> Wires userRepo -> userUseCase -> userHandler.
- **Kong Client**: `internal/gateway/kong/client.go` -> Logic to talk to Kong Admin API (Port 8001).
- **Session Logic**: `internal/gateway/session/service.go` -> Redis interaction for Phantom Tokens.

## 🚀 Environment & Ports

- **Portal Backend**: Port `3502`
- **Kong Proxy**: Port `3600` (Public Entry)
- **Kong Admin**: Port `3602`
- **Postgres**: Port `3605` (Internal: 5432)
- **Redis**: Port `3606`
