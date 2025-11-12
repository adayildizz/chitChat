Implementation Report: Ricart-Agrawala Algorithm

# 2. Algorithm Overview

## 2.1 Ricart-Agrawala Algorithm

The Ricart-Agrawala algorithm is a permission-based distributed mutual exclusion algorithm that requires \*\*2(N-1) messages per critical section access, where N is the number of nodes. This is an improvement over Lamport's algorithm which requires 3(N-1) messages.

Key Properties:

- **Mutual Exclusion**: Guaranteed through permission-based access
- **Deadlock Freedom**: All requests are eventually granted
- **Fairness**: Requests are ordered by Lamport timestamps with node ID tie-breaking
- **No Starvation**: Every request eventually receives all required permissions

### 2.2 Algorithm Steps

1. **Requesting Critical Section:**

   - Node increments its Lamport clock and records the timestamp
   - Broadcasts REQUEST message to all other nodes with its timestamp
   - Waits for REPLY messages from all other nodes
   - Enters critical section when all replies are received

2. **Receiving a Request:**

   - Node updates its Lamport clock based on received timestamp
   - If not requesting CS, reply immediately
   - If requesting CS, compare timestamps:
     - If incoming request has lower timestamp (or same timestamp with smaller node ID), reply immediately
     - Otherwise, defer the reply until exiting CS

3. **Exiting Critical Section:**
   - Node exits CS
   - Sends REPLY to all deferred requests

## 3. Implementation Details

### 3.1 Architecture

The implementation consists of three main components:

#### 3.1.1 Node Structure

```go
type Node struct {
    node_id          string
    isRequesting     bool
    mu               sync.Mutex
    deferred         map[string]bool
    lamport          int64
    requestTS        int64
    remainingReplies int
    peers            map[string]string
    clients          map[string]ra.RAServerClient
    cond             *sync.Cond
    addr             string
}
```

**Key Fields:**

- `lamport`: Logical clock value
- `requestTS`: Timestamp of current request
- `remainingReplies`: Counter for pending replies
- `deferred`: Map of nodes whose replies are deferred
- `isRequesting`: Flag indicating if node is requesting CS

#### 3.1.2 Core Functions

**Lamport Clock Management:**

- `bump()`: Increments local Lamport clock
- `onReceive(timestamp)`: Updates clock on message receipt: `max(local, received) + 1`

**Request Handling:**

- `RequestCriticalSection(work func())`: Main entry point for CS access
- `BroadcastRequest(reqTs)`: Sends REQUEST to all peers
- `RequestCS()`: gRPC handler for incoming requests

**Reply Handling:**

- `SendReply(to)`: Sends REPLY to a specific node
- `ReplyCS()`: gRPC handler for incoming replies

**Ordering:**

- `tupleLess(tsJ, j, tsI, i)`: Compares (timestamp, nodeID) tuples for ordering

### 3.2 Message Protocol

**Request Message:**

```protobuf
message Request {
    int64 Timestamp = 1;
    string From = 2;
}
```

**Reply Message:**

```protobuf
message Reply {
    int64 Timestamp = 1;
    string From = 2;
}
```

### 3.4 Synchronization Strategy

The implementation uses locking:

- Mutex protects shared state (Lamport clock, request status, deferred map)
- Condition variable (`sync.Cond`) for efficient waiting on replies
- Lock is held only when accessing/modifying shared state
- Network I/O performed outside critical sections to avoid blocking

## 4. Testing Methodology

### 4.1 Test Scenarios

**1. Sequential Access Test**

- Nodes request CS one after another
- **Expected**: Each node enters immediately after previous exits
- **Verifies**: Basic mutual exclusion

**2. Concurrent Access Test**

- Multiple nodes request CS simultaneously
- **Expected**: Nodes enter in timestamp order
- **Verifies**: Fairness and ordering

**3. Deferred Reply Test**

- Node A requests, then Node B requests while A is in CS
- **Expected**: B's request is deferred, receives reply when A exits
- **Verifies**: Deferred reply mechanism

**4. Failure Handling Test**

- Kill a node while others are requesting
- **Expected**: Remaining nodes don't deadlock
- **Verifies**: Error handling and deadlock prevention

### 4.2 Verification Criteria

- ✅ Only one `ENTER_CS` event at a time across all nodes
- ✅ Lamport clocks are monotonically increasing
- ✅ All requests eventually receive all replies
- ✅ Deferred requests are flushed after CS exit
- ✅ No deadlocks or infinite waiting

## 5. Results

### 5.1 Correctness Verification

The implementation successfully demonstrates:

1. **Mutual Exclusion**: Testing with 3 nodes shows only one node enters CS at a time
2. **Fairness**: Nodes with lower timestamps enter CS first
3. **No Deadlocks**: System handles peer failures without hanging
4. **Proper Ordering**: Lamport clocks maintain causal ordering

### 5.2 Performance Characteristics

- **Message Complexity**: 2(N-1) messages per CS access (optimal for permission-based algorithms)
- **Latency**: Depends on network RTT and processing time
- **Throughput**: Limited by sequential CS access, but fair scheduling ensures no starvation

### 5.3 Example Execution Trace

```
[L=1][node=A][EVENT=RequestStart][reqTs=1][need=2]
[L=1][node=A][EVENT=RequestSent][to=B][reqTs=1]
[L=1][node=A][EVENT=RequestSent][to=C][reqTs=1]
[L=2][node=B][EVENT=RequestRecv][from=A][ts=1][decision=REPLIED]
[L=2][node=B][EVENT=ReplySent][to=A]
[L=2][node=C][EVENT=RequestRecv][from=A][ts=1][decision=REPLIED]
[L=2][node=C][EVENT=ReplySent][to=A]
[L=3][node=A][EVENT=ReplyRecv][from=B][outstanding=1]
[L=3][node=A][EVENT=ReplyRecv][from=C][outstanding=0]
[L=3][node=A][EVENT=ENTER_CS]
[node=A][EVENT=EXIT_CS]
```

This trace demonstrates:

- Proper Lamport clock updates
- All peers receive and respond to requests
- Node enters CS only after receiving all replies
- Clean exit from CS
