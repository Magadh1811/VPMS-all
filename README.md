# Vehicles Parking Management System — Backend (Go + Gin)

The backend for the Vehicles Parking Management System is developed in **Go (Golang)** using the **Gin web framework**. It provides RESTful APIs to handle user authentication, vehicles, parking spots, and bookings.

This backend is **deployed on Railway**, and the database is hosted on **Supabase (PostgreSQL)**.

---

## Overview

* **Language:** Go (Golang)
* **Framework:** Gin
* **Database:** PostgreSQL (Supabase)
* **Auth:** JWT (HMAC) + role-based middleware
* **Deployment:** Railway (backend) + Docker support

---

## Project Structure

```
Backend-Go/
├── cmd/server/main.go    # Entry point
├── internal/
│   ├── config/           # Env & config loading
│   ├── db/               # DB connection
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # JWT + RBAC
│   └── router/           # Gin router setup
├── .env                  # Local env vars
├── Dockerfile
├── go.mod / go.sum
└── README.md (this)
```

---

## API Routes (from `router.Setup`)

### Public

| Method | Path           | Description           |
| ------ | -------------- | --------------------- |
| POST   | `/auth/signup` | Register user         |
| POST   | `/auth/login`  | Login and receive JWT |

### Authenticated (JWT required)

| Method | Path                       | Description                          |
| ------ | -------------------------- | ------------------------------------ |
| GET    | `/user/me`                 | Current user profile                 |
| POST   | `/vehicles`                | Add a vehicle (plate, type)          |
| POST   | `/parking/book`            | Book a spot (by spotId & vehicleId)  |
| POST   | `/parking/release/:spotId` | Release an active booking for a spot |
| GET    | `/parking/history`         | User booking history                 |

### Admin (JWT + `role=admin`)

| Method | Path                 | Description                              |
| ------ | -------------------- | ---------------------------------------- |
| POST   | `/parking-lots`      | Create parking lot                       |
| POST   | `/parking-spots`     | Create spot (lot, level, number, status) |
| DELETE | `/parking-spots/:id` | Delete spot                              |
| GET    | `/parking/occupancy` | Occupancy snapshot/metrics               |
| GET    | `/parking/reports`   | Reporting endpoints                      |

> Unknown routes return: `404 { error: { code: "NOT_FOUND", message: "route not found" } }`

---

## Auth Flow

1. **Signup** → create user → (hash stored)
2. **Login** → returns **JWT** in response; use `Authorization: Bearer <token>` for protected routes

---

##  Middleware

* `AuthJWT(cfg)` — validates JWT and sets user context
* `RequireRole("admin")` — ensures admin-only access

---

##  Running Locally

```bash
go run ./cmd/server
# or
PORT=8080 GIN_MODE=release go run ./cmd/server
```

### Docker

```bash
docker build -t parking-backend .
docker run -p 8080:8080 --env-file .env parking-backend
```


##  Notes

* DB constraints ensure **one active booking per spot** and **per vehicle** (enforced in DB).
* Spot states: `AVAILABLE`, `OCCUPIED`, `DISABLED`.
* Email uniqueness is case-insensitive.

---

**Author:** Magadh Singh · **Email:** [magadh18.singh@gmail.com](mailto:magadh18.singh@gmail.com) · **GitHub:** [https://github.com/Magadh1811](https://github.com/Magadh1811)
