<!--
🔐 AUTHENTICATION
-->

<p align="center">
  <img src="https://img.shields.io/badge/Auth-Username%20%26%20Password-blueviolet?logo=lock" alt="Authentication" />
</p>

# 🔐 Authentication

[← Back to README](../README.md) | [Architecture](architecture.md) | [Routes](routes.md) | [Middleware](middleware.md) | [Setup](setup.md)

---

> **Username/Password authentication with per-user secret key signing.**

---

## 🧑‍💻 How It Works

- Users are stored in the `users` collection in MongoDB.
- Each user has a `username`, `password`, `secret_key`, and `is_active` status.
- On login, the API:
  1. Accepts `username` and `password` via POST `/auth/login`.
  2. Looks up the user in MongoDB.
  3. Validates the password and `is_active` status.
  4. Uses the user's `secret_key` to sign a session token (HMAC).
  5. Returns the signed token to the client.
- For protected routes:
  - The client must send the token in the `Authorization: Bearer <token>` header.
  - The API validates the token using the user's `secret_key` from MongoDB.

---

## 📝 Example User Document

```json
{
  "username": "fraiday-dev-user",
  "password": "fraiday-dev-pwd",
  "secret_key": "test",
  "created_at": "2020-01-01T00:00:00Z",
  "is_active": true
}
```

---

## 🔑 Login Example

**Request:**
```http
POST /auth/login
Content-Type: application/json

{
  "username": "fraiday-dev-user",
  "password": "fraiday-dev-pwd"
}
```

**Response:**
```json
{
  "token": "<signed_token>"
}
```

---

## 🛡️ Using the Token

Include the token in the `Authorization` header for all protected endpoints:

```http
Authorization: Bearer <signed_token>
```

---

## ⚠️ Security Notes

- Tokens are signed with the user's `secret_key` and expire after 1 hour.
- Passwords are stored as plaintext for demo purposes; use hashing in production.
- Only active users can authenticate.
- Never share your token or secret key.

---

## 🔗 Related Docs

- [Routes](routes.md)
- [Middleware](middleware.md)
- [Setup](setup.md)
- [Maintainer](maintainer.md)

---