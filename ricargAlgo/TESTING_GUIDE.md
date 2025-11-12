# Testing Guide for Ricart-Agrawala Algorithm

## Prerequisites

1. Make sure you have Go installed
2. Install dependencies: `go mod tidy`
3. Generate gRPC code (if needed): `protoc --go_out=. --go-grpc_out=. grpc/proto.proto`

## Quick Start

### Option 1: Manual Testing (Recommended for understanding)

#### Windows (PowerShell):
```powershell
# Terminal 1 - Node A
go run . --id A --listen :50051 --mode manual

# Terminal 2 - Node B  
go run . --id B --listen :50052 --mode manual

# Terminal 3 - Node C
go run . --id C --listen :50053 --mode manual
```

#### Linux/Mac:
```bash
# Terminal 1 - Node A
go run . --id A --listen :50051 --mode manual

# Terminal 2 - Node B
go run . --id B --listen :50052 --mode manual

# Terminal 3 - Node C
go run . --id C --listen :50053 --mode manual
```

Or use the test scripts:
- Windows: `.\test_script.ps1`
- Linux/Mac: `chmod +x test_script.sh && ./test_script.sh`

### Option 2: Demo Mode (Automatic)

```powershell
# Terminal 1
go run . --id A --listen :50051 --mode demo

# Terminal 2
go run . --id B --listen :50052 --mode demo

# Terminal 3
go run . --id C --listen :50053 --mode demo
```

In demo mode, each node automatically requests CS every 2 seconds.

## Test Scenarios

### 1. **Basic Mutual Exclusion Test**
- Start all 3 nodes
- Press ENTER in Node A to request CS
- **Expected**: Node A enters CS, does work, exits
- **Verify**: Only one node is in CS at a time (check logs for `ENTER_CS` and `EXIT_CS`)

### 2. **Concurrent Request Test**
- Start all 3 nodes
- Press ENTER in Node A, B, and C almost simultaneously (within 1 second)
- **Expected**: 
  - Nodes with lower timestamps (or lexicographically smaller IDs) get priority
  - Only one enters CS at a time
  - Others wait and enter in order
- **Verify**: Check Lamport timestamps in logs - requests with lower timestamps should enter first

### 3. **Sequential Request Test**
- Start all 3 nodes
- Press ENTER in Node A, wait for it to exit
- Press ENTER in Node B, wait for it to exit
- Press ENTER in Node C
- **Expected**: Each node enters CS immediately after the previous one exits
- **Verify**: No waiting/deferring should occur

### 4. **Deferred Reply Test**
- Start all 3 nodes
- Press ENTER in Node A (don't wait)
- Immediately press ENTER in Node B
- **Expected**: 
  - Node A enters CS first
  - Node B's request is deferred by Node A
  - When Node A exits, it sends reply to Node B
  - Node B then enters CS
- **Verify**: Look for `DEFERRED` and `ReplyFlush` events in logs

### 5. **Tie-Breaking Test**
- Start all 3 nodes
- Try to get two nodes to request at exactly the same Lamport timestamp
- **Expected**: Node with lexicographically smaller ID (A < B < C) should win
- **Verify**: Check `tupleLess` function behavior in logs

### 6. **Lamport Clock Consistency**
- Monitor the Lamport clock values in all logs
- **Expected**: 
  - When a node receives a message, it updates its clock: `max(local_clock, received_clock) + 1`
  - Clocks should be consistent across all nodes
- **Verify**: Check `[L=...]` values in log entries

## What to Look For in Logs

### Correct Behavior Indicators:
- ✅ `[EVENT=RequestStart]` followed by `[EVENT=RequestSent]` to all peers
- ✅ `[EVENT=RequestRecv]` with `decision=REPLIED` or `decision=DEFERRED`
- ✅ `[EVENT=ReplyRecv]` decrements `outstanding` counter
- ✅ `[EVENT=ENTER_CS]` appears only when `outstanding=0`
- ✅ `[EVENT=EXIT_CS]` followed by `[EVENT=ReplyFlush]` if there were deferred requests
- ✅ Only ONE node in CS at any given time
- ✅ Lamport timestamps are monotonically increasing

### Error Indicators:
- ❌ Multiple nodes in CS simultaneously
- ❌ Node stuck waiting (outstanding replies never reach 0)
- ❌ Lamport clock going backwards
- ❌ Missing replies (outstanding counter stuck)
- ❌ Deadlock (all nodes waiting)

## Example Correct Log Sequence

```
[L=1][node=A][EVENT=RequestStart][reqTs=1][need=2]
[L=1][node=A][EVENT=RequestSent][to=B][reqTs=1]
[L=1][node=A][EVENT=RequestSent][to=C][reqTs=1]
[L=2][node=B][EVENT=RequestRecv][from=A][ts=1][decision=REPLIED][reqTs=0][requesting=false]
[L=2][node=B][EVENT=ReplySent][to=A]
[L=2][node=C][EVENT=RequestRecv][from=A][ts=1][decision=REPLIED][reqTs=0][requesting=false]
[L=2][node=C][EVENT=ReplySent][to=A]
[L=3][node=A][EVENT=ReplyRecv][from=B][outstanding=1]
[L=3][node=A][EVENT=ReplyRecv][from=C][outstanding=0]
[L=3][node=A][EVENT=ENTER_CS]
[node=A][EVENT=EXIT_CS]
```

## Troubleshooting

### Nodes can't connect:
- Check that ports 50051, 50052, 50053 are not in use
- Verify `peers.json` has correct addresses
- Ensure nodes start in order (wait 1-2 seconds between starts)

### Deadlock:
- Check if any node failed to send requests (look for ERROR messages)
- Verify all nodes are running
- Check network connectivity

### Incorrect ordering:
- Verify Lamport clock updates correctly
- Check that `tupleLess` function works correctly
- Ensure timestamps are compared properly

## Advanced Testing

### Test with 4+ nodes:
Edit `peers.json` to add more nodes:
```json
{
  "A": "localhost:50051",
  "B": "localhost:50052",
  "C": "localhost:50053",
  "D": "localhost:50054"
}
```

### Test failure scenarios:
- Kill a node while others are requesting (test the fix for deadlock)
- Start nodes in different orders
- Test with slow network (use `--slow-reply-ms` flag if implemented)

