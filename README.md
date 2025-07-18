<p align="center">
  <img src="https://img.shields.io/badge/Go-1.23-blue?logo=go" alt="Go 1.23" />
  <img src="https://img.shields.io/badge/Gin-Framework-green?logo=go" alt="Gin" />
  <img src="https://img.shields.io/badge/Docker-Ready-blue?logo=docker" alt="Docker" />
  <img src="https://img.shields.io/badge/MongoDB-Repository-brightgreen?logo=mongodb" alt="MongoDB" />
</p>

<h1 align="center">🚀 Genie : AI Interactions </h1>
<p align="center"><b>Go REST API based on Gin Web Framework</b></p>

---

> **Welcome!**  
> This is your entry point to the API Service documentation.  
> Use the navigation below to explore architecture, routes, middleware, setup, and more.

---

## 📚 Table of Contents

| Section | Description |
|---------|-------------|
| 🏗️ [Architecture](docs/architecture.md) | Visual overview of the system and components |
| 🛣️ [Routes](docs/routes.md) | All API endpoints, requests & responses |
| 🧩 [Middleware](docs/middleware.md) | Middleware stack and usage |
| ⚙️ [Setup](docs/setup.md) | How to set up and run the application |
| 🧑‍💻 [Maintainer](docs/maintainer.md) | Code owner and contact |
| 🔐 [Authentication](docs/auth.md) | Username/password authentication |

---

## ✨ Features

- **Gin Framework** for fast, idiomatic HTTP APIs
- **Clean Architecture** for maintainability and testability
- **Centralized Logging** with zap
- **Robust Middleware**: Request ID, logging, recovery, CORS, error handler
- **Health Endpoints**: `/health` (detailed), `/ping` (simple)
- **MongoDB Data Layer**: Repository pattern, ready for business logic
- **Environment-based Config**: `.env.dev` and `.env.prod`
- **Docker & Compose**: Local and production orchestration
- **Makefile**: Streamlined local development

---

## 🚦 Quick Start

> **Requirements:**  
> Go 1.23+, Docker, Docker Compose, Make

```bash
# Clone the repo
git clone https://github.com/fraiday-org/api-service.git

# Build and run the API locally
make build
make run

# Or use Docker Compose (modular, multi-service)
ENV_FILE=env/.env.dev PROFILE=dev make docker-up

# Stop all services
make docker-down

# Build Docker image for API only
make docker-build

# For more, see [Setup](docs/setup.md)
```

---

## 🗂️ Project Structure

See the [Architecture](docs/architecture.md) doc for a full breakdown.

```
cmd/            # Entrypoint
internal/       # API, middleware, handlers, repository, models
env/            # Environment configs
Dockerfile      # Docker build
docker-compose.yml
Makefile
docs/           # 📄 All documentation
```

---

## 📝 Documentation

- For a deep dive into the system, see the docs above or jump to:
  - [Architecture](docs/architecture.md)
  - [Routes](docs/routes.md)
  - [Middleware](docs/middleware.md)
  - [Setup](docs/setup.md)
  - [Maintainer](docs/maintainer.md)
  - [Authentication](docs/auth.md)

---

## 🐳 Modular Docker & Compose

- **Dockerfile**: Multi-stage, supports build args for service name and env file.
- **docker-compose.yml**: Modular, supports profiles, service defaults, and future microservices.
- **Makefile**: Modular targets for build, run, lint, test, compose up/down, and per-service actions.

See [Setup](docs/setup.md) for full details and advanced usage.
