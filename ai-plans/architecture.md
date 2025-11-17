# Crypto Market Data System Architecture

## Container Architecture

### 1. Exchange Connector Containers (C++)
- One container per exchange (Binance, Coinbase, Kraken, etc.)
- WebSocket connection to exchange
- Parse JSON market data
- Write directly to PostgreSQL
- Base: Alpine Linux + statically linked C++ binary
- Size target: 15-30 MB per container

### 2. PostgreSQL Database Container
- Store time-series market data
- Tables: trades, orderbook, ticker
- Indexes on timestamp and symbol
- Base: postgres:alpine
- Size: ~80-100 MB

### 3. Backend API Container (Go)
- REST API for querying market data
- Read-only database access
- Endpoints: /trades, /orderbook, /ticker
- Base: scratch or distroless
- Size target: 10-20 MB

### 4. Frontend Container (React + TypeScript)
- React + TypeScript SPA
- Nginx serves static build files
- Base: nginx:alpine
- Size target: 20-30 MB

## Data Flow

```
Exchange WebSocket → Connector Container → PostgreSQL
                                              ↓
                                         Backend API
                                              ↓
                                    Frontend (React + TS)
```

## Tech Stack Summary

| Component | Language/Tech | Base Image | Estimated Size |
|-----------|---------------|------------|----------------|
| Binance Connector | C++ | alpine | 20 MB |
| Coinbase Connector | C++ | alpine | 20 MB |
| Kraken Connector | C++ | alpine | 20 MB |
| PostgreSQL | - | postgres:alpine | 100 MB |
| Backend API | Go | scratch | 15 MB |
| Frontend | React + TypeScript | nginx:alpine | 25 MB |

**Total System**: ~220 MB for all containers

## Key Design Principles

1. Single responsibility per container
2. Statically linked binaries (no runtime dependencies)
3. Minimal base images (Alpine, scratch, distroless)
4. Direct database writes (no message queue for simplicity)
5. Read-only API (no trading logic)

## Database Schema (Simplified)

```sql
CREATE TABLE trades (
    id BIGSERIAL PRIMARY KEY,
    exchange VARCHAR(20),
    symbol VARCHAR(20),
    price NUMERIC,
    quantity NUMERIC,
    timestamp TIMESTAMPTZ,
    side VARCHAR(4)
);

CREATE INDEX idx_trades_time ON trades(timestamp DESC);
CREATE INDEX idx_trades_symbol ON trades(symbol, timestamp DESC);
```

## Network Communication

- Connectors → PostgreSQL: Direct TCP connection
- Backend → PostgreSQL: Read-only connection
- Frontend → Backend: HTTP/REST
- All containers in same Docker network

## Scalability Considerations

- Add more exchange connectors by duplicating container pattern
- Scale backend API horizontally (stateless)
- PostgreSQL can be replaced with TimescaleDB for better time-series performance
- Add Redis cache layer if needed (future enhancement)
