<!--
🧩 MIDDLEWARE
-->

<p align="center">
  <img src="https://img.shields.io/badge/Middleware-Stack-blue?logo=stackshare" alt="Middleware Stack" />
</p>

# 🧩 Middleware

[← Back to README](../README.md) | [Architecture](architecture.md) | [Routes](routes.md) | [Setup](setup.md)

---

> **Middleware** are the backbone of request processing, providing logging, error handling, security, and more.

---

## 🧩 Middleware Stack

| Middleware      | Icon | Description                                      |
|-----------------|------|--------------------------------------------------|
| Request ID      | 🆔   | Attaches a unique ID to each request             |
| Logger          | 📋   | Logs request/response details with zap           |
| Recovery        | 🛡️   | Recovers from panics, logs errors                |
| CORS            | 🌐   | Enables Cross-Origin Resource Sharing            |
| Error Handler   | ❗   | Centralizes error responses and formatting       |

---

## 🆔 Request ID

- **Purpose:** Assigns a unique `X-Request-ID` to every request for traceability.
- **How it works:** Checks for incoming `X-Request-ID` header or generates a new UUID.
- **Usage:** Used in logs and error responses for correlation.

```go
func RequestID() gin.HandlerFunc {
    // ...
}
```

---

## 📋 Logger

- **Purpose:** Structured logging of all requests and responses.
- **How it works:** Logs method, path, status, latency, client IP, and request ID using zap.
- **Best Practice:** Use logs for monitoring, debugging, and auditing.

```go
func Logger(logger *zap.Logger) gin.HandlerFunc {
    // ...
}
```

---

## 🛡️ Recovery

- **Purpose:** Prevents panics from crashing the server.
- **How it works:** Catches panics, logs stack trace, returns 500 error with request ID.
- **Tip:** Always keep this middleware enabled in production.

```go
func Recovery(logger *zap.Logger) gin.HandlerFunc {
    // ...
}
```

---

## 🌐 CORS

- **Purpose:** Allows cross-origin requests for APIs.
- **How it works:** Sets `Access-Control-Allow-*` headers for all responses.
- **Config:** Default is permissive (`*`), restrict in production as needed.

```go
func CORS() gin.HandlerFunc {
    // ...
}
```

---

## ❗ Error Handler

- **Purpose:** Centralizes error formatting and responses.
- **How it works:** Catches errors set in Gin context, returns JSON with error and request ID.
- **Best Practice:** Use for consistent error responses across all endpoints.

```go
func ErrorHandler() gin.HandlerFunc {
    // ...
}
```

---

> **All middleware are applied globally and in order, ensuring every request is logged, traced, and safely handled.**

---
