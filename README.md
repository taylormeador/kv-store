# Distributed Key-Value Store with Raft Consensus

A learning project implementing a replicated key-value database with the Raft consensus algorithm in Go.

## Overview

Production-quality distributed key-value store featuring:
- **Raft Consensus** - Leader election, log replication, strong consistency
- **Fault Tolerance** - Automatic failover, survives minority node failures
- **Durability** - Persistent Raft log with crash recovery
- **High Performance** - RWMutex-optimized concurrent operations
- **Simple Protocol** - Text-based TCP interface (GET, SET, DELETE, EXISTS)

## Architecture
```
┌─────────────────── Raft Cluster (3 nodes) ──────────────────┐
│                                                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                    Leader Node                       │   │
│  │  ┌──────────────┐        ┌──────────────┐            │   │
│  │  │  KV Server   │◄──────►│  Raft Node   │            │   │
│  │  │  (port 808x) │        │  (port 909x) │            │   │
│  │  │   ┌──────┐   │        │  ┌────────┐  │──► Disk    │   │
│  │  │   │Store │   │        │  │  Log   │  │            │   │
│  │  │   └──────┘   │        │  └────────┘  │            │   │
│  │  └──────────────┘        └──────────────┘            │   │
│  └────────────────┬──────────────────────┬──────────────┘   │
│                   │   Replication        │                  │
│                   ▼                      ▼                  │
│  ┌─────────────────────┐  ┌─────────────────────┐           │
│  │   Follower Node     │  │   Follower Node     │           │
│  │  ┌────┐  ┌──────┐   │  │  ┌────┐  ┌──────┐   │           │
│  │  │ KV │  │ Raft │   │  │  │ KV │  │ Raft │   │           │
│  │  └────┘  └──────┘   │  │  └────┘  └──────┘   │           │
│  └─────────────────────┘  └─────────────────────┘           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Getting Started

### Prerequisites
- Go 1.23+
- Docker & Docker Compose

### Run Cluster
```bash
# Start 3-node cluster
docker-compose up

# Logs show leader election
# One node becomes leader, others become followers
```

### Basic Operations
```bash
# Write to leader
echo "SET foo bar" | nc localhost 8081
# Response: OK

# Read value
echo "GET foo" | nc localhost 8081
# Response: bar

# Delete key
echo "DELETE foo" | nc localhost 8081
# Response: TRUE
```

### Test Fault Tolerance
```bash
# Write data
echo "SET test value" | nc localhost 8081

# Kill leader
docker-compose stop node-1

# New leader elected (200-300ms)
# Data persists on new leader
echo "GET test" | nc localhost 8082
# Response: value

# Restart old leader - catches up automatically
docker-compose start node-1
```

## Protocol
```
SET key value\n    → OK\n
GET key\n          → value\n
DELETE key\n       → TRUE\n | FALSE\n
EXISTS key\n       → TRUE\n | FALSE\n
```

## Changelog

### Phase 5: Distributed Consensus (Current)
- Raft consensus implementation
- Leader election with randomized timeouts
- Log replication with majority quorum
- Log consistency checks and repair
- Persistent Raft log (JSON format)
- Automatic failover and crash recovery

### Phase 4: Async Write Performance
- Background fsync (1-second intervals)
- **Performance:** 1.8x throughput, 467x faster p99 latency
- Graceful shutdown with final fsync

### Phase 3: Concurrency Optimization  
- Upgraded Mutex → RWMutex
- **Performance:** 5.6x faster reads, 53% overall improvement

### Phase 2: Durability
- Write-ahead logging (WAL)
- Synchronous fsync per write
- Replay on startup for crash recovery

### Phase 1: Foundation
- Basic TCP server with text protocol
- In-memory hashmap store
- GET, SET, DELETE, EXISTS operations

## Roadmap

### Phase 6: Snapshots & Log Compaction
- Periodic state snapshots
- Log truncation to bound disk usage
- Fast recovery for far-behind followers

### Phase 7: Transactions
- Multi-key atomic operations
- Optimistic concurrency control
- Snapshot or serializable isolation

### Phase 8: Sharding
- Consistent hashing across multiple Raft groups
- Partition keyspace by shard
- Rebalancing on membership changes

## Testing
```bash
# Unit tests
go test ./...

# Benchmarks
go test -bench=. ./internal/store

# Load test
cd cmd/loadtest
python loadtest.py --clients 10 --operations 1000
```
