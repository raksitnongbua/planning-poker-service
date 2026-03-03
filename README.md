# Planning Poker Service

> Backend service for [Corgi Planning Poker](https://www.corgiplanningpoker.com) — a collaborative estimation tool for agile teams.

![Version](https://img.shields.io/badge/version-v1.2.0-blue)
![Go](https://img.shields.io/badge/Go-1.21.4-00ADD8?logo=go&logoColor=white)
![Fiber](https://img.shields.io/badge/Fiber-v2.52.0-00ACD7?logo=go&logoColor=white)
![Firebase](https://img.shields.io/badge/Firebase-v3.13.0-FFCA28?logo=firebase&logoColor=black)
![Firestore](https://img.shields.io/badge/Firestore-v1.14.0-FFCA28?logo=googlecloud&logoColor=white)
![WebSocket](https://img.shields.io/badge/WebSocket-enabled-brightgreen)

---

## Overview

A Planning Poker backend built with **Go** using a **Backend For Frontend (BFF)** architecture. It provides REST and WebSocket APIs to support real-time collaborative estimation sessions.

## Tech Stack

| Technology | Version | Purpose |
|---|---|---|
| [Go](https://golang.org) | 1.21.4 | Primary language |
| [Fiber](https://gofiber.io) | v2.52.0 | HTTP framework |
| [Firebase](https://firebase.google.com) | v3.13.0 | Auth & app integration |
| [Firestore](https://cloud.google.com/firestore) | v1.14.0 | Database |
| [WebSocket](https://github.com/gofiber/contrib/websocket) | v1.3.0 | Real-time communication |
| [JWX](https://github.com/lestrrat-go/jwx) | v2.0.19 | JWT handling |

## Getting Started

### Prerequisites

- Go 1.21.4+
- Firebase project with Firestore enabled

### Run locally

```bash
go run main.go
```

## Contributing

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and commit them.
4. Push to your fork and open a pull request.

Happy Planning! 🚀
