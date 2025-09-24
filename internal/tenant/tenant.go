package tenant

import (
	"errors"
	"sync"

	"golang.org/x/crypto/bcrypt"

	"github.com/seabfh/redis-multi-tenant-proxy/internal/config"
)

// Manager handles tenant identification and prefix management
type Manager struct {
	tenants        map[string]config.TenantConfig
	authRequired   bool
	connectionAuth map[string]string // Maps connection ID to tenant
	mu             sync.RWMutex
}

// NewManager creates a new tenant manager
func NewManager(tenants map[string]config.TenantConfig, authRequired bool) *Manager {
	if tenants == nil {
		tenants = make(map[string]config.TenantConfig)
	}

	return &Manager{
		tenants:        tenants,
		authRequired:   authRequired,
		connectionAuth: make(map[string]string),
	}
}

// Authenticate authenticates a connection with tenant credentials
func (m *Manager) Authenticate(connID, username, password string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, exists := m.tenants[username]
	if !exists {
		return false, errors.New("invalid username")
	}

	// Validate password
	if tenant.PasswordHash != "" {
		// Using bcrypt hash (more secure)
		if err := bcrypt.CompareHashAndPassword([]byte(tenant.PasswordHash), []byte(password)); err != nil {
			return false, errors.New("invalid password")
		}
	} else if tenant.Password != "" {
		// Using plain password (less secure)
		if tenant.Password != password {
			return false, errors.New("invalid password")
		}
	} else {
		// No password set for this tenant
		return false, errors.New("tenant has no password configured")
	}

	// Store the tenant username for this connection
	m.connectionAuth[connID] = username
	return true, nil
}

// GetPrefix returns the tenant prefix for a connection
func (m *Manager) GetPrefix(connID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	username, exists := m.connectionAuth[connID]
	if !exists {
		if m.authRequired {
			return "", errors.New("authentication required")
		}
		// If auth not required, return empty prefix (no isolation)
		return "", nil
	}

	// Return the tenant's prefix
	return m.tenants[username].Prefix, nil
}

// ConnectionClosed handles connection closure
func (m *Manager) ConnectionClosed(connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.connectionAuth, connID)
}

// IsAuthRequired returns whether authentication is required
func (m *Manager) IsAuthRequired() bool {
	return m.authRequired
}
