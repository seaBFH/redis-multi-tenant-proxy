package tenant

import (
	"errors"
	"sync"
)

// Manager handles tenant identification and prefix management
type Manager struct {
	tenantMap      map[string]string
	authRequired   bool
	connectionAuth map[string]string // Maps connection ID to tenant
	mu             sync.RWMutex
}

// NewManager creates a new tenant manager
func NewManager(tenantMap map[string]string, authRequired bool) *Manager {
	if tenantMap == nil {
		tenantMap = make(map[string]string)
	}

	return &Manager{
		tenantMap:      tenantMap,
		authRequired:   authRequired,
		connectionAuth: make(map[string]string),
	}
}

// Authenticate authenticates a connection with tenant credentials
func (m *Manager) Authenticate(connID, username, password string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// In a real implementation, validate credentials against a secure store
	// For this example, we just check if the tenant exists in our map
	prefix, exists := m.tenantMap[username]
	if !exists {
		return false, errors.New("invalid credentials")
	}

	// Store the tenant prefix for this connection
	m.connectionAuth[connID] = prefix
	return true, nil
}

// GetPrefix returns the tenant prefix for a connection
func (m *Manager) GetPrefix(connID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prefix, exists := m.connectionAuth[connID]
	if !exists {
		if m.authRequired {
			return "", errors.New("authentication required")
		}
		// If auth not required, return empty prefix (no isolation)
		return "", nil
	}

	return prefix, nil
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
