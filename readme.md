# Chirpy API

Chirpy is a small Twitter-like HTTP API written with Go's standard `net/http` package. It stores users, chirps, and refresh tokens in PostgreSQL.

Base URL when running locally: `http://localhost:8080`

## Requirements

- Go 1.26 or later
- PostgreSQL
- A database schema created from the migrations in `sql/schema`

## Configuration and startup

Create a `.env` file in the repository root:

```env
DB_URL=postgres://USER:PASSWORD@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
JWT_SECRET=replace-with-a-long-random-secret
POLKA_KEY=replace-with-the-webhook-api-key
```

Apply the migrations before starting the server. For example, with Goose:

```sh
goose -dir sql/schema postgres "$DB_URL" up
go run .
```

The server listens on port `8080`. A Nix development shell is also provided and includes Goose, sqlc, and PostgreSQL tooling:

```sh
nix develop
```

`PLATFORM=dev` enables the development-only reset endpoint. Do not use that setting in an environment where data must be preserved.

## Conventions

- Request and response bodies use JSON unless noted otherwise.
- IDs are UUIDs and timestamps are RFC 3339 timestamp strings.
- Authenticated endpoints expect `Authorization: Bearer <access-token>`.
- Access tokens are HS256 JWTs and expire one hour after login or refresh.
- Refresh tokens are opaque strings and expire 60 days after login.
- Error responses generally use this shape:

```json
{"error":"description"}
```

## API reference

### Health check

#### `GET /api/healthz`

Returns `200 OK` with a plain-text body of `OK` when the server is available.

```sh
curl http://localhost:8080/api/healthz
```

### Create a user

#### `POST /api/users`

Creates a user. Email addresses must be unique.

Request body:

```json
{
  "email": "alice@example.com",
  "password": "correct-horse-battery-staple"
}
```

Response: `201 Created`

```json
{
  "id": "a1b2c3d4-e5f6-4789-8abc-def012345678",
  "created_at": "2025-01-01T12:00:00Z",
  "updated_at": "2025-01-01T12:00:00Z",
  "email": "alice@example.com",
  "is_chirpy_red": false
}
```

### Log in

#### `POST /api/login`

Verifies an email and password, then returns the user, an access token, and a refresh token.

Request body:

```json
{
  "email": "alice@example.com",
  "password": "correct-horse-battery-staple"
}
```

Response: `200 OK`

```json
{
  "id": "a1b2c3d4-e5f6-4789-8abc-def012345678",
  "created_at": "2025-01-01T12:00:00Z",
  "updated_at": "2025-01-01T12:00:00Z",
  "email": "alice@example.com",
  "is_chirpy_red": false,
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "opaque-refresh-token"
}
```

Invalid credentials return `401 Unauthorized`.

### Update the current user

#### `PUT /api/users`

Updates the email address and password of the user identified by the access token.

Headers:

```http
Authorization: Bearer <access-token>
Content-Type: application/json
```

Request body:

```json
{
  "email": "alice.new@example.com",
  "password": "a-new-password"
}
```

Response: `200 OK` with the user object described in [Create a user](#create-a-user).

### Create a chirp

#### `POST /api/chirps`

Creates a chirp for the authenticated user.

Headers:

```http
Authorization: Bearer <access-token>
Content-Type: application/json
```

Request body:

```json
{
  "body": "Hello, Chirpy!"
}
```

A chirp body can be at most 140 bytes. Whole-word, case-insensitive occurrences of `kerfuffle`, `sharbert`, and `fornax` are replaced with `****` before storage.

Response: `201 Created`

```json
{
  "id": "b1b2c3d4-e5f6-4789-8abc-def012345678",
  "created_at": "2025-01-01T12:01:00Z",
  "updated_at": "2025-01-01T12:01:00Z",
  "body": "Hello, Chirpy!",
  "user_id": "a1b2c3d4-e5f6-4789-8abc-def012345678"
}
```

### List chirps

#### `GET /api/chirps`

Returns all chirps as a JSON array. No authentication is required.

Optional query parameters:

| Parameter | Description |
| --- | --- |
| `author_id` | Limit results to chirps from this user UUID. |
| `sort` | Use `desc` for newest-first ordering. Any other value uses the database's default ordering. |

Examples:

```sh
curl http://localhost:8080/api/chirps
curl 'http://localhost:8080/api/chirps?author_id=a1b2c3d4-e5f6-4789-8abc-def012345678&sort=desc'
```

Response: `200 OK`

```json
[
  {
    "id": "b1b2c3d4-e5f6-4789-8abc-def012345678",
    "created_at": "2025-01-01T12:01:00Z",
    "updated_at": "2025-01-01T12:01:00Z",
    "body": "Hello, Chirpy!",
    "user_id": "a1b2c3d4-e5f6-4789-8abc-def012345678"
  }
]
```

### Get a chirp

#### `GET /api/chirps/{chirpID}`

Returns a single chirp. No authentication is required.

```sh
curl http://localhost:8080/api/chirps/b1b2c3d4-e5f6-4789-8abc-def012345678
```

Response: `200 OK` with a chirp object. An invalid UUID returns `400 Bad Request`; an unknown chirp returns `404 Not Found`.

### Delete a chirp

#### `DELETE /api/chirps/{chirpID}`

Deletes a chirp only when it belongs to the authenticated user.

Headers:

```http
Authorization: Bearer <access-token>
```

Response: `204 No Content`.

Returns `401 Unauthorized` for a missing or invalid token, `403 Forbidden` if the chirp belongs to another user, and `404 Not Found` if it does not exist.

### Refresh an access token

#### `POST /api/refresh`

Exchanges a valid, non-revoked refresh token for a new one-hour access token.

Headers:

```http
Authorization: Bearer <refresh-token>
```

Response: `200 OK`

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

Invalid, expired, or revoked refresh tokens return `401 Unauthorized`.

### Revoke a refresh token

#### `POST /api/revoke`

Revokes the supplied refresh token. A revoked token can no longer be used with `/api/refresh`.

Headers:

```http
Authorization: Bearer <refresh-token>
```

Response: `204 No Content`.

### Polka webhook

#### `POST /api/polka/webhooks`

Processes subscription events from Polka. It currently handles `user.upgraded` by setting the referenced user's `is_chirpy_red` field to `true`. Other event names are acknowledged without changes.

Headers:

```http
Authorization: ApiKey <POLKA_KEY>
Content-Type: application/json
```

Request body:

```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "a1b2c3d4-e5f6-4789-8abc-def012345678"
  }
}
```

Response: `204 No Content`. Missing or incorrect API keys return `401 Unauthorized`; an unknown user returns `404 Not Found` for `user.upgraded` events.

## Development endpoints

These endpoints support local development and are not part of the public API.

### `GET /admin/metrics`

Returns an HTML page with the number of requests served under `/app/` since startup or the most recent reset.

### `POST /admin/reset`

Available only when `PLATFORM=dev`. It resets the `/app/` request count and deletes every user from the database, including related refresh tokens. It returns `200 OK` on success and `403 Forbidden` for other platform values.

### `GET /app/...`

Serves static files from the repository root. Requests to this route are counted by the metrics endpoint.
