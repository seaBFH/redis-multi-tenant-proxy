# Redis Multi-Tenant Proxy

A high-performance Redis proxy server designed for multi-tenant environments. This proxy enables secure tenant isolation through key prefixing while providing seamless Redis cluster support.

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
- ❌ **Cross-node operations not implemented**: Operations spanning multiple Redis nodes are not yet supported
- 🔄 **Planned features**:
  - Enhanced `MGET` and `MSET` operations with cross-node support
  - `SCAN` command implementation with tenant-aware key iteration
  - Advanced cluster topology awareness
  - Rate limiting per tenant
  - Metrics and monitoring endpoints

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
./bin/hashpassword
# Enter your password when prompted
# Copy the generated hash to your config file
```

## Usage

### Starting the Proxy

```bash
# Using config file
./bin/proxy --config config.yaml

# Using environment variables (if no config file)
export REDIS_LISTEN_ADDR=":6380"
export REDIS_CLUSTER_NODES="redis1:6379,redis2:6379,redis3:6379"
./bin/proxy
```

### Client Connection

#### Python Example

```python
import redis

# Connect to the proxy with tenant credentials
client = redis.Redis(
    host='localhost',
    port=6380,
    username='tenant1',
    password='secure_password1',
    decode_responses=True
)

# All operations are automatically prefixed with "tenant1:"
client.set('user:123:profile', 'John Doe')  # Key becomes "tenant1:user:123:profile"
profile = client.get('user:123:profile')     # Retrieves "tenant1:user:123:profile"

# Multi-key operations
client.mset({
    'user:123:email': 'john@example.com',
    'user:123:age': '30'
})

values = client.mget(['user:123:email', 'user:123:age'])
print(values)  # ['john@example.com', '30']
```

#### Node.js Example

```javascript
const redis = require('redis');

const client = redis.createClient({
    host: 'localhost',
    port: 6380,
    username: 'tenant1',
    password: 'secure_password1'
});

await client.connect();

// Tenant-isolated operations
await client.set('session:abc123', 'user_data');
const session = await client.get('session:abc123');
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

```yaml
version: '3.8'

services:
  redis-proxy:
    build: .
    ports:
      - "6380:6380"
    volumes:
      - ./config.yaml:/etc/redis-proxy/config.yaml
    depends_on:
      - redis-node1
      - redis-node2
      - redis-node3

  redis-node1:
    image: redis:7-alpine
    ports:
      - "7001:6379"
    command: redis-server --cluster-enabled yes --cluster-node-timeout 5000

  redis-node2:
    image: redis:7-alpine
    ports:
      - "7002:6379"
    command: redis-server --cluster-enabled yes --cluster-node-timeout 5000

  redis-node3:
    image: redis:7-alpine
    ports:
      - "7003:6379"
    command: redis-server --cluster-enabled yes --cluster-node-timeout 5000
```

### Production Considerations

1. **Security**: Always use bcrypt password hashes in production
2. **Monitoring**: Monitor proxy logs for performance and errors
3. **Scaling**: Deploy multiple proxy instances behind a load balancer
4. **Redis Cluster**: Ensure proper Redis cluster configuration and monitoring
5. **Network**: Use secure networks and consider TLS termination at load balancer level

## Development

### Project Structure

```
.
├── cmd/
│   ├── proxy/          # Main proxy application
│   └── hashpassword/   # Password hashing utility
├── internal/
│   ├── config/         # Configuration management
│   ├── proxy/          # Proxy server implementation
│   ├── redisops/       # Redis operation handling
│   └── tenant/         # Tenant management
├── config/
│   └── config.yaml     # Example configuration
├── example/
│   └── client/         # Example client implementations
├── Dockerfile          # Container definition
├── docker-compose.yml  # Development environment
└── Makefile           # Build targets
```

### Building and Testing

```bash
# Build all binaries
make bin

# Build Docker image
make image

# Run tests (when available)
go test ./...

# Format code
go fmt ./...

# Install dependencies
go mod download
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

[Add your license information here]

## Support

For issues and questions:
1. Check the [Issues](https://github.com/seabfh/redis-multi-tenant-proxy/issues) page
2. Create a new issue with detailed information
3. Include configuration, logs, and steps to reproduce

---

**Note**: This project is actively developed with focus on multi-tenant Redis deployments. Cross-node operations and enhanced cluster features are planned for future releases.