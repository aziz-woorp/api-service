<!--
🚀 RELEASE NOTES v1.0
-->

# 🚀 Release Notes: API Service v1.0

**Date:** 2025-07-18

---

## 1. Repository & Codebase

- **Directory Structure:**  
  - Modular, clean, and extensible.  
  - Clear separation: `cmd/`, `internal/`, `env/`, `docs/`, and root configs.

- **Guidelines:**  
  - Follows clean architecture and Go best practices.
  - All internal imports are relative for private/local use.

- **Documentation:**  
  - Comprehensive and visually structured.
  - Covers usage, setup, routes, middleware, and architecture.
  - Centralized in `README.md` and `/docs`.

- **Swagger Docs:**  
  - Ready for integration at `/docs` for API exploration (future-ready).

- **Middleware:**  
  - Centralized stack: Request ID, Logger, Recovery, CORS, Error Handler.
  - All requests pass through middleware for security and traceability.

- **Local Setup:**  
  - Simple with Makefile: `make build`, `make run`.
  - Environment configs in `env/`.

- **Docker & Docker-Compose:**  
  - Modular, multi-stage Dockerfile.
  - Compose supports profiles, service defaults, and future microservices.
  - One command to spin up the stack:  
    `ENV_FILE=env/.env.dev PROFILE=dev make docker-up`

- **Authentication:**  
  - Token-based (HMAC, per-user secret).
  - `/auth/login` is open; all other endpoints are protected.
  - Usage documented in [Authentication](../../auth.md).

- **Code Guidelines:**  
  - Idiomatic Go, modular, and ready for extension.
  - Linting and testing supported via Makefile.

---

## 2. Database

- **MongoDB**  
  - Used for all persistent data.
  - User authentication and future business logic ready.
  - Managed via Docker Compose for local/dev.

---

## 3. Environment

- **dev/prod**  
  - Environment-specific configs in `env/.env.dev` and `env/.env.prod`.
  - Easy switching via Makefile and Compose.

---

**Summary:**  
This release delivers a robust, modular, and production-ready API service. The codebase is clean, extensible, and fully documented for both developers and stakeholders. All aspects—from setup to security—are designed for clarity, maintainability, and ease of use.
