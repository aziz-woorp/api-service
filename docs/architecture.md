<!--
🏗️ ARCHITECTURE
-->

<p align="center">
  <img src="https://img.shields.io/badge/Architecture-Mermaid-blueviolet?logo=mermaid" alt="Architecture" />
</p>

# 🏗️ API Service Architecture

[← Back to README](../README.md) | [Routes](routes.md) | [Middleware](middleware.md) | [Setup](setup.md)

---

> **Overview**  
> The diagram below shows the high-level and component-level architecture of the API Service.

---

```mermaid
flowchart TD
    subgraph Client
      U[User / Client]
    end

    subgraph API["API Service (Gin)"]
      direction TB
      RQ[Request Router]
      MW[Middleware Stack]
      HD[Handlers]
      SRV[Service Layer (future)]
      REPO[Repository Layer]
      CFG[Config Loader]
      LOG[Logger (zap)]
    end

    subgraph DB["MongoDB"]
      MDB[(MongoDB Instance)]
    end

    U -->|HTTP Request| RQ
    RQ -->|Passes through| MW
    MW -->|Invokes| HD
    HD -->|Business Logic| SRV
    SRV -->|Data Access| REPO
    REPO -->|CRUD| MDB
    CFG -->|Env Vars| API
    LOG -->|Logs| API
    HD -->|Logs| LOG
    MW -->|Logs| LOG
```

---

## 🗺️ Legend

- **Client**: User or external system making HTTP requests.
- **Request Router**: Gin's router, matches routes.
- **Middleware Stack**: Request ID, Logger, Recovery, CORS, Error Handler.
- **Handlers**: Endpoint logic (e.g., `/health`, `/ping`).
- **Service Layer**: (Pluggable for business logic, coming soon)
- **Repository Layer**: MongoDB data access.
- **Config Loader**: Loads env/config for the app.
- **Logger**: Centralized logging with zap.
- **MongoDB**: Persistent data store.

---

> **Design Principles:**  
> - All requests flow through middleware before reaching handlers.  
> - Repository pattern for data access.  
> - Centralized logging and configuration.

---

## 🐳 Modular Deployment

- **Dockerfile**: Multi-stage, supports build args for service name and env file.
- **docker-compose.yml**: Modular, supports profiles, service defaults, and future microservices.
- **Makefile**: Modular targets for build, run, lint, test, compose up/down, and per-service actions.

> The architecture is designed for extensibility: add new microservices, scale components, and manage everything with a unified, modular toolchain.
