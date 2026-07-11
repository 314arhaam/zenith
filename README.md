# Zenith

[![CI](https://github.com/314arhaam/zenith/actions/workflows/ci.yml/badge.svg)](https://github.com/314arhaam/zenith/actions/workflows/ci.yml)
[![E2E Test](https://github.com/314arhaam/zenith/actions/workflows/e2e.yml/badge.svg)](https://github.com/314arhaam/zenith/actions/workflows/e2e.yml)

> A lightweight service registry consisting of a Go REST API server and a Cobra-powered CLI client. To keep Github Actions VMs up in a controlled manner, to deploy platforms.

Zenith is a small distributed service management project written in Go. It consists of two independent applications:

* **Zenith Server** — exposes a REST API for managing services.
* **Zenith CLI** — a Cobra-based command-line application for interacting with the server.

The project demonstrates:

* REST API development in Go
* Cobra CLI application design
* End-to-end testing with GitHub Actions
* Multi-runner integration testing
* Tailscale-based networking between CI runners

---

# Architecture

```
                   +----------------------+
                   |    Zenith CLI        |
                   |   (Cobra Client)     |
                   +----------+-----------+
                              |
                         HTTP REST API
                              |
                              v
                   +----------------------+
                   |    Zenith Server     |
                   |    Service Registry  |
                   +----------------------+
```

---

# Features

## Server

* REST API
* Register services
* Remove services
* Query service status
* Health endpoint
* In-memory service registry
* Thread-safe data access

## Client

* Built with Cobra
* Simple command structure
* Supports remote servers through `--url`
* Continuous health checking
* Continuous service monitoring

---

# Project Structure

```
.
├── client/
│   ├── cmd/
│   └── main.go
│
├── server/
│   └── main.go
│
├── handlers/
├── models/
├── tests/
└── go.mod
```

---

# Building

## Server

```bash
go build -o zenith_server server/main.go
```

## Client

```bash
go build -o zenith_client client/main.go
```

---

# Running

Start the server:

```bash
./zenith_server 8080
```

Use the CLI:

```bash
./zenith_client --url http://localhost:8080 ping
```

---

# CLI Commands

```
add         Add a service
check       Continuously check service availability
ping        Ping the server
remove      Remove a service
status      Query service status
```

Global flag:

```
--url
```

Default:

```
http://0.0.0.0:8080
```

Examples:

```bash
# Ping server
zenith_client ping

# Register a service
zenith_client add api

# Check a service
zenith_client check api

# Check every registered service
zenith_client check

# Service status
zenith_client status api

# Remove a service
zenith_client remove api
```

---

# REST API

| Method | Endpoint  | Description               |
| ------ | --------- | ------------------------- |
| GET    | `/ping`   | Health endpoint           |
| POST   | `/add`    | Register a service        |
| DELETE | `/remove` | Remove a service          |
| GET    | `/status` | Query service status      |
| GET    | `/check`  | Check registered services |

Example:

```bash
curl -X POST http://localhost:8080/add \
    -H "Content-Type: application/json" \
    -d '{"service_name":"example"}'
```

---

# End-to-End Testing

Zenith includes an end-to-end GitHub Actions workflow that validates the complete distributed system.

The workflow performs the following steps:

1. Build the server binary.
2. Build the CLI binary.
3. Upload both artifacts.
4. Start the server on a dedicated GitHub Actions runner.
5. Connect all runners using Tailscale.
6. Register the main server service.
7. Wait until the server becomes reachable.
8. Launch two independent client runners.
9. Register services from each client.
10. Continuously verify service availability.
11. Remove services one by one.
12. Finally remove the server service and complete the shutdown sequence.

This simulates a small distributed environment where multiple machines communicate with a central Zenith server over a private Tailscale network.

---

# Example Workflow

```
Server
   │
   ├── Register "main_service"
   │
   ├──────────────┐
   │              │
Client-01      Client-02
   │              │
Register      Register
   │              │
Check          Check
   │              │
Remove         Remove
   └──────┬───────┘
          │
      Remove server
```

---

# Technologies

* Go
* Cobra CLI
* net/http
* GitHub Actions
* Tailscale

---

# License

MIT License.
