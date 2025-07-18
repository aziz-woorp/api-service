<!--
⚙️ SETUP
-->

<p align="center">
  <img src="https://img.shields.io/badge/Setup-Guide-success?logo=make" alt="Setup Guide" />
</p>

# ⚙️ Setup & Local Development

[← Back to README](../README.md) | [Architecture](architecture.md) | [Routes](routes.md) | [Middleware](middleware.md)

---

> **Get started with the API Service in just a few steps!**

---

## 🛠️ Prerequisites

- [Go 1.23+](https://golang.org/dl/)
- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/)

---

## 🗂️ Environment Setup

1. **Clone the repository**  
   ```bash
   git clone https://github.com/fraiday-org/api-service.git
   cd api-service
   ```

2. **Configure environment variables**  
   Copy or edit the provided `.env.dev` and `.env.prod` in the `env/` directory.

   > **Tip:**  
   > Use `.env.dev` for development and `.env` for production.

---

## 🏃 Running the Application

### 🧑‍💻 Local Development (with local MongoDB)

```bash
make run
```

- Ensure MongoDB is running locally at the URI specified in your `.env.dev`.

---

### 🐳 Modular Docker Compose (Recommended)

```bash
# Start all services for development
ENV_FILE=env/.env.dev PROFILE=dev make docker-up

# Start all services for production
ENV_FILE=env/.env.prod PROFILE=prod make docker-up

# Stop all services
make docker-down

# Build Docker image for API only
make docker-build

# Build and run a future worker service (example)
make build-worker
make run-worker
```

> **Note:**  
> Docker Compose will spin up all defined services (API, MongoDB, and future microservices) using modular profiles and service defaults.

---

### 🏗️ Build Docker Image (API)

```bash
make docker-build
```

---

## 🧹 Cleaning Up

```bash
make clean
```

---

## 🛠️ Troubleshooting

- **Port already in use:**  
  Change `APP_PORT` in your `.env.dev` or stop the process using the port.

- **MongoDB connection errors:**  
  Ensure MongoDB is running and the URI in your `.env` is correct.

- **Permission issues on Linux/Mac:**  
  Try running Docker commands with `sudo` or adjust user permissions.

---

> **Need more help?**  
> See [Maintainer](maintainer.md) for contact info.

---

<p align="center">
  <b>Explore more:</b>  
  🏗️ [Architecture](architecture.md) | 🛣️ [Routes](routes.md) | 🧩 [Middleware](middleware.md)
</p>
