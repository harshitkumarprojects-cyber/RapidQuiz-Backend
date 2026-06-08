# RapidQuiz — Backend

A real-time multiplayer quiz platform backend built with **Go**, **Gin**, **MongoDB**, **Redis**, and **WebSockets**.

Hosts create quizzes, start game sessions with a shareable room code, and participants join to answer questions in real time. Scores are tracked per answer and ranked on a live Redis-backed leaderboard.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go |
| HTTP Framework | Gin |
| Database | MongoDB (via qmgo) |
| Cache / Leaderboard | Redis |
| Real-time | WebSockets (gorilla/websocket) |
| Auth | JWT (HS256) |
| Password Hashing | bcrypt |

---

## Project Structure

```
backend/
├── controllers/        # Request handlers
│   ├── AuthControllers.go
│   ├── GameControlllers.go
│   ├── LeaderboardControllers.go
│   └── QuizControllers.go
├── database/           # DB + Redis connection helpers
│   ├── db.go
│   └── redis.go
├── middlewares/        # JWT auth middleware
│   └── auth.go
├── models/             # Data models
│   ├── Answer.go
│   ├── GameSession.go
│   ├── Participant.go
│   ├── QuestionModel.go
│   ├── QuizModel.go
│   ├── UserModel.go
│   └── WSmessage.go
├── routers/            # Route registration
│   ├── AuthRoutes.go
│   ├── GameRoutes.go
│   ├── LeaderboardRoutes.go
│   └── QuizRoutes.go
├── utils/              # JWT, hashing, helpers
│   ├── hash.go
│   ├── helperFunctions.go
│   └── jwt.go
├── websocket/          # WebSocket hub and handler
│   ├── Broadcast.go
│   ├── Handler.go
│   └── Hub.go
├── main.go
├── go.mod
└── .env.example
```

---

## Setup

### Prerequisites

- Go 1.21+
- MongoDB instance (local or Atlas)
- Redis instance (local or remote)

### Installation

```bash
git clone https://github.com/harshitkumar7525/RapidQuiz.git
cd RapidQuiz/backend
go mod download
```

### Environment Variables

Copy `.env.example` to `.env` and fill in the values:

```env
MONGO_URI=mongodb://localhost:27017
MONGO_DB=rapidquiz
PORT=8080
JWT_SECRET=your_secret_key_here
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
```

| Variable | Description | Default |
|---|---|---|
| `MONGO_URI` | MongoDB connection string | — |
| `MONGO_DB` | MongoDB database name | — |
| `PORT` | HTTP server port | `8080` |
| `JWT_SECRET` | Secret key for signing JWTs | — |
| `REDIS_ADDR` | Redis host and port | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password (leave blank if none) | `""` |

### Running

```bash
go run main.go
```

Or build and run:

```bash
go build -o rapidquiz
./rapidquiz
```

---

## Authentication

Protected routes require a JWT token in the `Authorization` header:

```
Authorization: Bearer <token>
```

Tokens are issued on register and login, and expire after **48 hours**.

---

## API Endpoints

### Auth

#### `POST /auth/register`

Register a new user.

**Request body**
```json
{
  "name": "Alice",
  "email": "alice@example.com",
  "password": "secretpassword"
}
```

**Success — `201 Created`**
```json
{
  "message": "user registered successfully",
  "token": "<jwt>"
}
```

**Errors**

| Status | Reason |
|---|---|
| `400 Bad Request` | Missing or invalid fields, or JSON parse error |
| `409 Conflict` | Email is already registered |
| `500 Internal Server Error` | Failed to save user or generate token |

---

#### `POST /auth/login`

Log in with email and password.

**Request body**
```json
{
  "email": "alice@example.com",
  "password": "secretpassword"
}
```

**Success — `200 OK`**
```json
{
  "message": "login successful",
  "token": "<jwt>"
}
```

**Errors**

| Status | Reason |
|---|---|
| `400 Bad Request` | Missing email or password |
| `401 Unauthorized` | Email not found or password does not match |
| `500 Internal Server Error` | Failed to generate token |

---

### Quizzes

> All routes except `GET /quizzes/:quizId` require `Authorization: Bearer <token>`.

#### `POST /quizzes/`

Create a new quiz. 🔒

**Request body**
```json
{
  "title": "General Knowledge",
  "description": "A fun trivia quiz",
  "questions": [
    {
      "question": "What is the capital of France?",
      "options": ["Berlin", "Madrid", "Paris", "Rome"],
      "correct_answer": "Paris",
      "time_limit": 20
    }
  ]
}
```

**Question rules:**
- `question` — required, non-empty string
- `options` — required, minimum 2 entries
- `correct_answer` — required; must exactly match one of the `options`
- `time_limit` — optional integer (seconds); defaults to `30` if omitted or `<= 0` during scoring

**Success — `201 Created`**
```json
{
  "message": "quiz created successfully"
}
```

**Errors**

| Status | Reason |
|---|---|
| `400 Bad Request` | Missing title, no questions, or a question fails validation |
| `401 Unauthorized` | Missing or invalid token |
| `500 Internal Server Error` | Database write failed |

---

#### `GET /quizzes/`

Get all quizzes created by the authenticated user. 🔒

**Success — `200 OK`**
```json
[
  {
    "id": "664abc123...",
    "title": "General Knowledge",
    "description": "A fun trivia quiz",
    "created_by": "663aaa...",
    "questions": [...],
    "created_at": "2024-05-10T12:00:00Z",
    "updated_at": "2024-05-10T12:00:00Z"
  }
]
```

Returns an empty array `[]` if the user has no quizzes.

**Errors**

| Status | Reason |
|---|---|
| `401 Unauthorized` | Missing or invalid token |
| `500 Internal Server Error` | Database read failed |

---

#### `GET /quizzes/:quizId`

Get a single quiz by ID. Public — no token required.

**Success — `200 OK`**
```json
{
  "id": "664abc123...",
  "title": "General Knowledge",
  "description": "A fun trivia quiz",
  "created_by": "663aaa...",
  "questions": [
    {
      "question": "What is the capital of France?",
      "options": ["Berlin", "Madrid", "Paris", "Rome"],
      "correct_answer": "Paris",
      "time_limit": 20
    }
  ],
  "created_at": "2024-05-10T12:00:00Z",
  "updated_at": "2024-05-10T12:00:00Z"
}
```

**Errors**

| Status | Reason |
|---|---|
| `400 Bad Request` | Invalid quiz ID format |
| `404 Not Found` | Quiz does not exist |

---

#### `PATCH /quizzes/:quizId`

Update an existing quiz. Only the quiz creator can update it. 🔒

**Request body** — same shape as `POST /quizzes/`, all fields optional except validation rules still apply to any questions provided.

**Success — `200 OK`**
```json
{
  "message": "quiz updated successfully"
}
```

**Errors**

| Status | Reason |
|---|---|
| `400 Bad Request` | Invalid quiz ID, invalid JSON, or question validation failure |
| `401 Unauthorized` | Missing or invalid token |
| `404 Not Found` | Quiz not found or user is not the creator |
| `500 Internal Server Error` | Database write failed |

---

#### `DELETE /quizzes/:quizId`

Delete a quiz. Only the quiz creator can delete it. 🔒

**Success — `200 OK`**
```json
{
  "message": "quiz deleted successfully"
}
```

**Errors**

| Status | Reason |
|---|---|
| `400 Bad Request` | Invalid quiz ID format |
| `401 Unauthorized` | Missing or invalid token |
| `404 Not Found` | Quiz not found or user is not the creator |
| `500 Internal Server Error` | Database delete failed |

---

### Games

#### `POST /games/create`

Start a new game session for a quiz. Only the quiz creator can start it. 🔒

**Request body**
```json
{
  "quiz_id": "664abc123..."
}
```

**Success — `201 Created`**
```json
{
  "message": "game session created successfully",
  "room_code": "A1B2C3",
  "game_id": "665def456...",
  "status": "waiting",
  "currentQuestion": 0
}
```

Game statuses: `waiting` → `running` → `paused` / `ended`

**Errors**

| Status | Reason |
|---|---|
| `400 Bad Request` | Missing `quiz_id` or invalid ID format |
| `401 Unauthorized` | Missing or invalid token |
| `403 Forbidden` | Authenticated user is not the quiz creator |
| `404 Not Found` | Quiz does not exist |
| `500 Internal Server Error` | Failed to generate room code or save session |

---

#### `POST /games/join`

Join a game session using a room code.

**Request body**
```json
{
  "room_code": "A1B2C3",
  "name": "Bob"
}
```

**Success — `200 OK`**
```json
{
  "message": "joined successfully",
  "participant_id": "666ghi789...",
  "game_id": "665def456..."
}
```

Save `participant_id` — it is required when submitting answers.

**Errors**

| Status | Reason |
|---|---|
| `400 Bad Request` | Missing `room_code` or `name` |
| `404 Not Found` | Room not found or game has ended |
| `409 Conflict` | Display name already taken in this room |
| `500 Internal Server Error` | Failed to save participant |

---

#### `GET /ws/:roomCode`

Open a WebSocket connection to a game room.

**URL example:** `ws://localhost:8080/ws/A1B2C3`

Once connected, any message sent by one client is broadcast to all other clients in the same room. The server uses the `WSMessage` shape:

```json
{
  "type": "string",
  "data": "<any>"
}
```

The connection is closed automatically when the client disconnects or a write error occurs.

---

### Leaderboard & Answers

#### `POST /games/:gameId/answer`

Submit an answer for a question. The game must be in `running` status.

**Request body**
```json
{
  "participant_id": "666ghi789...",
  "question_index": 0,
  "answer": "Paris"
}
```

- `question_index` — zero-based index into the quiz's `questions` array
- `answer` — the exact string of the chosen option

**Scoring:** correct answers earn `100 + time_limit` points (using the question's `time_limit`, defaulting to `30` if unset).

**Success — `200 OK`**
```json
{
  "is_correct": true,
  "score": 120,
  "message": "correct answer!"
}
```

Or for a wrong answer:
```json
{
  "is_correct": false,
  "score": 0,
  "message": "wrong answer"
}
```

**Errors**

| Status | Reason |
|---|---|
| `400 Bad Request` | Invalid game ID, missing fields, or invalid question index |
| `400 Bad Request` | Game is not currently running |
| `404 Not Found` | Game or quiz not found |
| `409 Conflict` | Answer already submitted for this question by this participant |
| `500 Internal Server Error` | Failed to save answer or update leaderboard |

---

#### `GET /games/:gameId/leaderboard`

Get the top 20 participants for a game, ranked by total score.

**Success — `200 OK`**
```json
{
  "game_id": "665def456...",
  "leaderboard": [
    {
      "rank": 1,
      "participant_id": "666ghi789...",
      "name": "Bob",
      "score": 340
    },
    {
      "rank": 2,
      "participant_id": "666jkl012...",
      "name": "Carol",
      "score": 220
    }
  ]
}
```

Scores are served from Redis and update in real time as answers are submitted. Leaderboard data expires after **24 hours**.

**Errors**

| Status | Reason |
|---|---|
| `400 Bad Request` | Invalid game ID format |
| `500 Internal Server Error` | Redis read failed |

---

## Data Models

### User
| Field | Type | Notes |
|---|---|---|
| `id` | ObjectID | Auto-generated |
| `name` | string | |
| `email` | string | Unique |
| `password` | string | bcrypt hashed; hidden from JSON responses |
| `created_at` | time | |
| `updated_at` | time | |

### Quiz
| Field | Type | Notes |
|---|---|---|
| `id` | ObjectID | Auto-generated |
| `title` | string | Required |
| `description` | string | Optional |
| `created_by` | ObjectID | Set from JWT on creation |
| `questions` | []Question | At least 1 required |
| `created_at` | time | |
| `updated_at` | time | |

### Question
| Field | Type | Notes |
|---|---|---|
| `question` | string | Required |
| `options` | []string | Min 2 |
| `correct_answer` | string | Must match one of `options` |
| `time_limit` | int | Seconds; defaults to 30 if unset |

### GameSession
| Field | Type | Notes |
|---|---|---|
| `id` | ObjectID | Auto-generated |
| `quiz_id` | ObjectID | |
| `host_id` | ObjectID | Set from JWT |
| `room_code` | string | 6-char alphanumeric, e.g. `A1B2C3` |
| `status` | string | `waiting`, `running`, `paused`, `ended` |
| `current_question` | int | Zero-based index |
| `started_at` | time? | Optional |
| `ended_at` | time? | Optional |

### Participant
| Field | Type | Notes |
|---|---|---|
| `id` | ObjectID | Auto-generated |
| `game_id` | ObjectID | |
| `name` | string | Unique within a room |
| `joined_at` | time | |

### Answer
| Field | Type | Notes |
|---|---|---|
| `id` | ObjectID | Auto-generated |
| `game_id` | ObjectID | |
| `participant_id` | ObjectID | |
| `question_index` | int | |
| `answer` | string | The chosen option text |
| `is_correct` | bool | |
| `score` | int | 0 if wrong; `100 + time_limit` if correct |
| `answered_at` | time | |

---

## MongoDB Collections

| Collection | Description |
|---|---|
| `users` | Registered user accounts |
| `quizzes` | Quiz definitions with questions |
| `game_sessions` | Active and historical game sessions |
| `participants` | Players who joined a game |
| `answers` | Submitted answers per participant per question |

---

## Redis Keys

| Key pattern | Type | TTL | Description |
|---|---|---|---|
| `leaderboard:<gameId>` | Sorted Set | 24h | Participant scores; member = `participant_id`, score = cumulative points |