# Ricart-Agrawala Algorithm Implementation

A distributed mutual exclusion algorithm implementation using gRPC and Lamport clocks.

## Quick Start

### 1. Build the project

```bash
go mod tidy
go build .
```

### 2. Run nodes manually (3 terminals)

**Terminal 1:**

```bash
go run . --id A --listen :50051 --mode manual
```

**Terminal 2:**

```bash
go run . --id B --listen :50052 --mode manual
```

**Terminal 3:**

```bash
go run . --id C --listen :50053 --mode manual
```

### 3. Test mutual exclusion

- In each terminal, press **ENTER** to request the critical section
- Watch the logs to verify:
  - Only one node enters CS at a time
  - Nodes wait for replies before entering
  - Deferred replies are sent after exiting CS

## Command Line Options

- `--id`: Node identifier (e.g., A, B, C)
- `--listen`: Address to listen on (e.g., :50051)
- `--peers`: Path to peers.json file (default: peers.json)
- `--mode`: `manual` (press ENTER) or `demo` (auto-request every 2s)

## Files

- `main.go`: Node initialization and main loop
- `server.go`: gRPC server handlers (RequestCS, ReplyCS)
- `client.go`: Client code for sending requests/replies
- `peers.json`: Configuration file mapping node IDs to addresses
