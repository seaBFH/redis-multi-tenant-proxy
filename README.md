# Redis Multi-Tenant Proxy

A Redis proxy server designed for multi-tenant environments. This proxy enables secure tenant isolation through key prefixing while providing seamless Redis cluster support.

## Features

### Current Features
- **Multi-tenant isolation**: Automatic key prefixing based on authenticated tenant
- **Redis Cluster support**: Full support for Redis cluster deployments with automatic redirection handling
- **Secure authentication**: Bcrypt password hashing with per-tenant credentials
- **Connection pooling**: Efficient connection management and reuse
- **Command routing**: Intelligent command routing with MOVED/ASK redirection support
- **Docker support**: Ready-to-deploy Docker containers

### Supported Redis Commands
- Basic key operations: `GET`, `SET`, `DEL`, `EXISTS`, `EXPIRE`, `TTL`, `INCR`, `DECR`
- Hash operations: `HGET`, `HSET`, `HMGET`, `HMSET`, `HDEL`
- List operations: `LPUSH`, `RPUSH`, `LPOP`, `RPOP`, `LRANGE`
- Set operations: `SADD`, `SREM`, `SMEMBERS`
- Sorted set operations: `ZADD`, `ZREM`, `ZRANGE`, `ZRANK`
- Multi-key operations: `MGET`, `MSET`
- Utility commands: `PING`, `INFO`

### Limitations & Roadmap
- ❌ **NOT PRODUCTION READY**: This project is in early development and should not be used in production environments
- ❌ **Cross-node operations not implemented**: Operations spanning multiple Redis nodes are not yet supported
- 🔄 **Planned features**:
  - Test and Benchmark
  - Enhanced `MGET` and `MSET` operations with cross-node support
  - `SCAN` command implementation with tenant-aware key iteration
  - Metrics and monitoring endpoints
  - Command filtering and blacklisting
  - Rate limiting per tenant

## Architecture

```
Client Applications
        ↓
Redis Multi-Tenant Proxy (Port 6380)
        ↓
Redis Cluster (Nodes on Port 6379)
```

The proxy acts as an intermediary between Redis clients and a Redis cluster, automatically:
1. Authenticating clients based on tenant credentials
2. Adding tenant-specific prefixes to all keys
3. Routing commands to appropriate Redis cluster nodes
4. Handling cluster redirections transparently

## Installation

### Prerequisites
- Go 1.25.1 or later
- Redis cluster setup (3+ nodes recommended)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/seabfh/redis-multi-tenant-proxy.git
cd redis-multi-tenant-proxy

# Build the application
make bin

# The binaries will be available in the bin/ directory
./bin/proxy --config config.yaml
```

### Docker Deployment

```bash
# Build the Docker image
make image

# Run with Docker Compose
docker-compose up -d
```

## Configuration

### Basic Configuration (`config.yaml`)

```yaml
# Proxy server configuration
listen_addr: ":6380"

# Redis Cluster configuration
cluster_nodes:
  - "redis-node1:6379"
  - "redis-node2:6379"
  - "redis-node3:6379"

# Authentication settings
auth_enabled: true

# Tenant configurations
tenants:
  tenant1:
    prefix: "tenant1:"
    password: "secure_password1"
    
  tenant2:
    prefix: "tenant2:"
    # Bcrypt hashed password (recommended)
    password_hash: "$2a$10$CxOSXe2LKT5g.nvBrno2Guz9R3yP6Z/io0xS5ZfAZjQxULoHDtRvW"
    
  production:
    prefix: "prod:"
    password_hash: "$2a$10$XQiOPG2Fj6GdJsRQxZJVZerugKxGC9Ky0mXyMsrXtFJLhGT8YKJkW"
    rate_limit: 1000
    max_connections: 50

# Performance settings
max_connections: 1000
log_level: "info"
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `listen_addr` | Address and port for the proxy server | `:6380` |
| `cluster_nodes` | List of Redis cluster node addresses | Required |
| `auth_enabled` | Enable/disable authentication | `true` |
| `tenants` | Tenant configuration map | `{}` |
| `max_connections` | Maximum concurrent connections | `1000` |
| `log_level` | Logging level (debug, info, warn, error) | `info` |

### Tenant Configuration

Each tenant can be configured with:
- `prefix`: Key prefix for tenant isolation (e.g., "tenant1:")
- `password`: Plain text password (not recommended for production)
- `password_hash`: Bcrypt hashed password (recommended)
- `rate_limit`: Optional rate limiting (requests per second)
- `max_connections`: Optional connection limit per tenant

### Generating Password Hashes

Use the included utility to generate secure password hashes:

```bash
./bin/hashpassword -password=your_password
# Enter your password when prompted
# Copy the generated hash to your config file
```

## Usage

### Starting the Proxy

```bash
./bin/proxy --config ./config/config.yaml
```

### Command Line Testing

```bash
# Connect using redis-cli
redis-cli -h localhost -p 6380

# Authenticate
AUTH tenant1 secure_password1

# Use Redis commands normally
SET user:1 "John Doe"
GET user:1
HSET user:1:profile name "John Doe" email "john@example.com"
HGET user:1:profile name
```

## Deployment

### Docker Compose

```bash
# launch the proxy and a Redis cluster
docker compose up -d
```
