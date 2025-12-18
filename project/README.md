# Concurrent Chat Server

Real-time chat server in Go demonstrating goroutines, channels, and synchronization primitives.

## What it does

Multi-user chat room where messages are broadcast to all connected clients in real-time. Handles concurrent connections without blocking using Go's concurrency model.

## Architecture

The server uses a hub-and-spoke model:

- **Main thread**: Runs HTTP server, spawns goroutines for each request
- **Broadcast manager**: Single goroutine that coordinates all message distribution (ChatRoom.Run)
- **Client handlers**: One goroutine per connected browser, keeps SSE connection alive
- **Monitor**: Background goroutine logging statistics every 5 seconds

Total goroutines: 3 + N (where N is number of connected users)

## Synchronization mechanisms

### Channels (message passing)

```go
broadcast  chan Message  // POST handler -> broadcast manager
register   chan *Client  // Client connects -> broadcast manager
unregister chan *Client  // Client disconnects -> broadcast manager
```

Each client has a personal channel that the broadcast manager writes to. This prevents direct goroutine-to-goroutine communication.

### Mutexes (protecting shared state)

```go
sync.RWMutex  // Protects clients map (many readers, one writer)
sync.Mutex    // Protects statistics counters
```

RWMutex allows multiple goroutines to read the client list simultaneously while broadcasting, but only one can modify it during registration/removal.

## Project structure

```
chat.go      Core concurrency logic: ChatRoom, channels, mutexes
handlers.go  HTTP endpoints: /post, /stream, /stats
monitor.go   Statistics logging goroutine
main.go      Server initialization and startup
index.html   Web UI
```

## Running

```bash
nix develop . # setting up environment
go run .
```

Open http://localhost:6969 in multiple browser tabs. Type messages in one tab, see them appear in all tabs instantly.

## How it works

Message flow from user to all clients:

1. User submits message via POST /post
2. Handler puts message on broadcast channel
3. ChatRoom.Run receives from channel
4. Saves to history, increments stats (with mutex)
5. Iterates client map (with RWMutex read lock)
6. Sends message to each client's personal channel
7. Each client goroutine receives from its channel
8. Pushes to browser via Server-Sent Events

Only ChatRoom.Run modifies shared state. Other goroutines communicate through channels. This prevents race conditions.

## Testing with curl

Post a message:

```bash
curl -X POST http://localhost:6969/post \
  -H "Content-Type: application/json" \
  -d '{"username":"Alice","content":"Hello from terminal"}'
```

Listen to message stream:

```bash
curl -N http://localhost:6969/stream
```

Get statistics:

```bash
curl http://localhost:6969/stats
```

## Key concurrency patterns

**Producer-consumer**: Multiple POST handlers produce messages, ChatRoom.Run consumes them.

**Fan-out broadcasting**: One sender (ChatRoom.Run) distributes to many receivers (client goroutines).

**Single-writer principle**: Only one goroutine modifies the client map, preventing data races.

**Non-blocking sends**: When broadcasting, uses select with default to skip slow clients instead of blocking.

## Assignment requirements

- Minimum 3 threads: Yes (main + broadcast + monitor + client handlers)
- At least 2 user task threads: Yes (one handleStream goroutine per user)
- Manager/coordinator thread: Yes (ChatRoom.Run)
- Synchronization primitives: Yes (channels + RWMutex + Mutex)
- Real-world concurrency benefit: Yes (handles thousands of simultaneous users)
- User interface: Yes (web browser)
- Technical documentation: This file

## Implementation notes

Go's HTTP server automatically creates a goroutine for each incoming request. This means handlePost and handleStream run concurrently without explicit goroutine creation in those functions.

The broadcast manager uses an infinite select loop. It runs for the lifetime of the server. Each case handles one type of event (registration, message, disconnection).

Client channels are buffered (capacity 10) to prevent slow clients from blocking the broadcast manager. If a client's buffer fills, messages are dropped for that client only.
