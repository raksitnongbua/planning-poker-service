# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run the service
go run main.go

# Run with hot-reload (requires air: go install github.com/cosmtrek/air@latest)
air

# Build
go build -o ./tmp/main .

# Test
go test ./...

# Run a single test package
go test ./internal/core/usecase/room/...
```

## Environment

Copy `.env.example` to `.env` and set:
- `FIREBASE_CREDENTIALS` — Firebase service account credentials as a JSON string
- `NEXTAUTH_SECRET` — Secret used to decrypt NextAuth.js session tokens from the frontend

The service starts on port **8080**.

## Architecture

The service is a BFF (Backend For Frontend) for [Corgi Planning Poker](https://www.corgiplanningpoker.com). It follows a layered architecture:

```
main.go
  → configs.Init()       # load env vars
  → repository.Init()    # connect Firestore
  → protocol.ServeREST() # start Fiber HTTP server
```

### Layers

- **`protocol/`** — HTTP router. Registers all Fiber routes and the WebSocket upgrade middleware.
- **`internal/core/handler/`** — HTTP/WS handlers. Parse requests, call usecases, return responses. Subpackages: `health`, `room`, `room_socket`, `user`.
- **`internal/core/usecase/`** — Business logic. Subpackages: `room` (CRUD), `room_socket` (real-time actions), `profile` (session decryption), `id_generator`, `timer`, `common`.
- **`internal/core/domain/`** — Core entities: `Room`, `Member`, `Profile`. Domain methods (JoinRoom, RevealCards, Restart, etc.) live directly on these structs.
- **`internal/core/auth/session/`** — Decrypts NextAuth.js JWE session cookies using HKDF + AES-GCM to extract the user profile.
- **`internal/repository/`** — Firestore access. `repository.go` holds the global `RoomsColRef`; `repository/room/` contains all room queries.
- **`configs/`** — Parses required env vars into a global `configs.Conf` struct at startup.
- **`constants/`** — Shared string constants (cookie names, encryption info).

### API Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/guest/sign-in` | Generate a guest UUID |
| POST | `/api/v1/new-room` | Create a new room |
| GET | `/api/v1/room/recent-rooms/:id` | Get recent rooms for a user |
| DELETE | `/api/v1/rooms/expired` | Clean up rooms inactive >30 days |
| WS | `/ws/room/:uid/:id` | WebSocket connection for a room |

### WebSocket Protocol

The WebSocket connection at `/ws/room/:uid/:id` uses JSON messages with `{ "action": string, "payload": any }`.

**Client → Server actions:** `JOIN_ROOM`, `UPDATE_ESTIMATED_VALUE`, `REVEAL_CARDS`, `RESET_ROOM`

**Server → Client actions:** `UPDATE_ROOM` (broadcasts updated room state to all room members), `NEED_TO_JOIN` (sent on connect if user is not yet a member)

All connected clients are stored in a global `map[*websocket.Conn]bool` with a mutex; broadcasts filter by `roomId` stored in `c.Locals`.

### Room State Machine

Room `Status` transitions:
- `VOTING` → `REVEALED_CARDS` (via `REVEAL_CARDS` action)
- `REVEALED_CARDS` → `VOTING` (via `RESET_ROOM` action)

Rooms are deleted by the cleanup endpoint after 30 days of inactivity (`roomRetention` constant in `repository/room/room_repository.go`).
