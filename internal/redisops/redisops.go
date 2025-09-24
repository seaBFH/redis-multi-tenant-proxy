package redisops

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/tidwall/redcon"
)

var (
	// Pattern to match MOVED response
	movedPattern = regexp.MustCompile(`MOVED\s+(\d+)\s+([^:]+):(\d+)`)

	// Pattern to match ASK response
	askPattern = regexp.MustCompile(`ASK\s+(\d+)\s+([^:]+):(\d+)`)
)

// CommandHandler handles Redis commands
type CommandHandler struct {
	cluster *redis.ClusterClient
	ctx     context.Context
}

// NewCommandHandler creates a new Redis command handler
func NewCommandHandler(clusterNodes []string) (*CommandHandler, error) {
	if len(clusterNodes) == 0 {
		return nil, errors.New("no Redis cluster nodes provided")
	}

	cluster := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: clusterNodes,
	})

	// Test connection
	ctx := context.Background()
	if err := cluster.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis cluster: %w", err)
	}

	return &CommandHandler{
		cluster: cluster,
		ctx:     ctx,
	}, nil
}

// ProcessCommand processes a Redis command with tenant prefix
func (h *CommandHandler) ProcessCommand(cmd redcon.Command, tenantPrefix string) (interface{}, error) {
	// Get command name
	cmdName := strings.ToUpper(string(cmd.Args[0]))

	// Clone the command for modification
	modifiedCmd := make([]interface{}, len(cmd.Args))
	for i, arg := range cmd.Args {
		modifiedCmd[i] = string(arg)
	}

	// Add tenant prefix to keys for key-based commands
	if isKeyCommand(cmdName) && len(cmd.Args) > 1 {
		keyPos := getKeyPositions(cmdName)
		for _, pos := range keyPos {
			if pos < len(modifiedCmd) {
				key := modifiedCmd[pos].(string)
				modifiedCmd[pos] = tenantPrefix + key
			}
		}
	}

	// Handle special commands
	switch cmdName {
	case "PING":
		return "PONG", nil

	case "AUTH":
		// Auth is handled by the proxy, not forwarded
		return nil, errors.New("AUTH command handled by proxy")

	case "INFO":
		// We could customize the INFO response or forward it
		return h.cluster.Info(h.ctx).Result()

	default:
		// Forward the modified command to Redis
		return h.executeCommand(modifiedCmd)
	}
}

// executeCommand executes a Redis command and handles redirections
func (h *CommandHandler) executeCommand(args []interface{}) (interface{}, error) {
	result, err := h.cluster.Do(h.ctx, args...).Result()

	// Check for MOVED/ASK errors and handle them
	if err != nil {
		errStr := err.Error()

		// Handle MOVED redirection
		if movedMatch := movedPattern.FindStringSubmatch(errStr); movedMatch != nil {
			slot, _ := strconv.Atoi(movedMatch[1])
			host := movedMatch[2]
			port, _ := strconv.Atoi(movedMatch[3])

			// debug log for redirection
			fmt.Printf("Redirecting command to %s:%d for slot %d\n", host, port, slot)

			// Create a new client for the target node
			targetClient := redis.NewClient(&redis.Options{
				Addr: fmt.Sprintf("%s:%d", host, port),
			})
			defer targetClient.Close()

			// Retry the command on the new node
			return targetClient.Do(h.ctx, args...).Result()
		}

		// Handle ASK redirection
		if askMatch := askPattern.FindStringSubmatch(errStr); askMatch != nil {
			slot, _ := strconv.Atoi(askMatch[1])
			host := askMatch[2]
			port, _ := strconv.Atoi(askMatch[3])

			// debug log for redirection
			fmt.Printf("Redirecting command to %s:%d for slot %d\n", host, port, slot)

			// Create a new client for the target node
			targetClient := redis.NewClient(&redis.Options{
				Addr: fmt.Sprintf("%s:%d", host, port),
			})
			defer targetClient.Close()

			// For ASK, we need to run ASKING command first
			targetClient.Do(h.ctx, "ASKING")

			// Retry the command on the new node
			return targetClient.Do(h.ctx, args...).Result()
		}
	}

	return result, err
}

// Shutdown closes the Redis client
func (h *CommandHandler) Shutdown() {
	if h.cluster != nil {
		h.cluster.Close()
	}
}

// isKeyCommand checks if a command operates on keys
func isKeyCommand(cmdName string) bool {
	// Common key commands
	keyCommands := map[string]bool{
		"GET": true, "SET": true, "DEL": true, "EXISTS": true,
		"EXPIRE": true, "TTL": true, "INCR": true, "DECR": true,
		"HGET": true, "HSET": true, "HMGET": true, "HMSET": true,
		"LPUSH": true, "RPUSH": true, "LPOP": true, "RPOP": true,
		"LRANGE": true, "SADD": true, "SREM": true, "SMEMBERS": true,
		"ZADD": true, "ZREM": true, "ZRANGE": true, "ZRANK": true,
		"MGET": true, "MSET": true,
		// Add more commands as needed
	}
	return keyCommands[cmdName]
}

// getKeyPositions returns the positions of keys in a command
func getKeyPositions(cmdName string) []int {
	// Define position of keys for different commands
	switch cmdName {
	case "GET", "EXISTS", "TTL", "INCR", "DECR", "DEL", "TYPE":
		return []int{1}

	case "SET", "EXPIRE", "SETEX":
		return []int{1}

	case "HGET", "HDEL":
		return []int{1}

	case "HSET":
		return []int{1}

	case "LPUSH", "RPUSH", "SADD", "SREM":
		return []int{1}

	case "MGET":
		// All args after position 1 are keys
		return makeRange(1, 99) // Arbitrary upper limit

	case "MSET":
		// Every odd-numbered arg is a key
		positions := []int{}
		for i := 1; i < 100; i += 2 {
			positions = append(positions, i)
		}
		return positions

	default:
		// Default case - assume the first arg is a key
		return []int{1}
	}
}

// makeRange creates an array of integers from min to max
func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}
