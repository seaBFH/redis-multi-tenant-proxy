package proxy

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tidwall/redcon"

	"github.com/seabfh/redis-multi-tenant-proxy/internal/config"
	"github.com/seabfh/redis-multi-tenant-proxy/internal/redisops"
	"github.com/seabfh/redis-multi-tenant-proxy/internal/tenant"
)

// Proxy is the main proxy server
type Proxy struct {
	cfg       *config.Config
	server    *redcon.Server
	redisOps  *redisops.CommandHandler
	tenantMgr *tenant.Manager
	connMutex sync.RWMutex
	connMap   map[redcon.Conn]string // Maps connections to connection IDs
	shutdown  chan struct{}
}

// NewProxy creates a new proxy server
func NewProxy(cfg *config.Config) (*Proxy, error) {
	// Initialize Redis command handler
	redisOps, err := redisops.NewCommandHandler(cfg.ClusterNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis handler: %w", err)
	}

	// Initialize tenant manager with the new tenant config
	tenantMgr := tenant.NewManager(cfg.Tenants, cfg.AuthEnabled)

	return &Proxy{
		cfg:       cfg,
		redisOps:  redisOps,
		tenantMgr: tenantMgr,
		connMap:   make(map[redcon.Conn]string),
		shutdown:  make(chan struct{}),
	}, nil
}

// Start starts the proxy server
func (p *Proxy) Start() error {
	// Create the Redis server
	p.server = redcon.NewServer(p.cfg.ListenAddr, p.handleCommand, p.handleConnect, p.handleDisconnect)

	// Start the server
	return p.server.ListenAndServe()
}

// Shutdown gracefully shuts down the proxy
func (p *Proxy) Shutdown() {
	// Signal shutdown
	close(p.shutdown)

	// Close the Redis server
	if p.server != nil {
		p.server.Close()
	}

	// Close Redis connections
	p.redisOps.Shutdown()
}

// handleConnect handles new client connections
func (p *Proxy) handleConnect(conn redcon.Conn) bool {
	// Generate a unique connection ID
	connID := uuid.New().String()

	// Store the connection mapping
	p.connMutex.Lock()
	p.connMap[conn] = connID
	p.connMutex.Unlock()

	log.Printf("New connection from %s (ID: %s)", conn.RemoteAddr(), connID)
	return true
}

// handleDisconnect handles client disconnections
func (p *Proxy) handleDisconnect(conn redcon.Conn, err error) {
	p.connMutex.Lock()
	connID, exists := p.connMap[conn]
	if exists {
		delete(p.connMap, conn)
	}
	p.connMutex.Unlock()

	if exists {
		p.tenantMgr.ConnectionClosed(connID)
		log.Printf("Connection closed from %s (ID: %s): %v", conn.RemoteAddr(), connID, err)
	}
}

// handleCommand processes Redis commands
func (p *Proxy) handleCommand(conn redcon.Conn, cmd redcon.Command) {
	start := time.Now()

	// Get the connection ID
	p.connMutex.RLock()
	connID, exists := p.connMap[conn]
	p.connMutex.RUnlock()

	if !exists {
		conn.WriteError("ERR internal server error: connection not found")
		return
	}

	// Get command name
	cmdName := ""
	if len(cmd.Args) > 0 {
		cmdName = string(cmd.Args[0])
	}

	// TODO: check if cmdName is valid Redis command

	// Handle AUTH command specially
	if len(cmd.Args) >= 3 && strings.ToUpper(cmdName) == "AUTH" {
		username := string(cmd.Args[1])
		password := string(cmd.Args[2])

		authenticated, err := p.tenantMgr.Authenticate(connID, username, password)
		if err != nil || !authenticated {
			conn.WriteError("ERR invalid username-password pair")
			return
		}

		conn.WriteString("OK")
		return
	}

	// Get tenant prefix for this connection
	tenantPrefix, err := p.tenantMgr.GetPrefix(connID)
	if err != nil {
		conn.WriteError("NOAUTH Authentication required.")
		return
	}

	// Process the command
	result, err := p.redisOps.ProcessCommand(cmd, tenantPrefix)
	if err != nil {
		conn.WriteError("ERR " + err.Error())
		return
	}

	// Write the result back to the client
	p.writeResult(conn, result)

	// Log the command (in a production environment, you'd want to make this configurable)
	duration := time.Since(start)
	log.Printf("Command: %s, Duration: %s", cmdName, duration)
}

// writeResult writes a result to the client connection
func (p *Proxy) writeResult(conn redcon.Conn, result interface{}) {
	switch v := result.(type) {
	case nil:
		conn.WriteNull()
	case string:
		conn.WriteString(v)
	case int64:
		conn.WriteInt64(v)
	case int:
		conn.WriteInt64(int64(v))
	case bool:
		if v {
			conn.WriteInt(1)
		} else {
			conn.WriteInt(0)
		}
	case []interface{}:
		conn.WriteArray(len(v))
		for _, item := range v {
			p.writeResult(conn, item)
		}
	case []string:
		conn.WriteArray(len(v))
		for _, item := range v {
			conn.WriteString(item)
		}
	case error:
		conn.WriteError("ERR " + v.Error())
	default:
		// Try to convert to string
		conn.WriteString(fmt.Sprintf("%v", v))
	}
}
