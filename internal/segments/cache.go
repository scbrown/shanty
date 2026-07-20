package segments

import (
	"os/exec"
	"strings"
	"sync"
	"time"
)

// cache stores segment results for a duration to avoid hammering external commands.
var cache = &segmentCache{
	entries: make(map[string]cacheEntry),
	ttl:     30 * time.Second,
}

type cacheEntry struct {
	value   string
	expires time.Time
}

type segmentCache struct {
	mu      sync.Mutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

func (c *segmentCache) get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok || time.Now().After(entry.expires) {
		return "", false
	}
	return entry.value, true
}

func (c *segmentCache) set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{
		value:   value,
		expires: time.Now().Add(c.ttl),
	}
}

// stAvailable checks if the st binary is on PATH.
func stAvailable() bool {
	_, err := exec.LookPath("st")
	return err == nil
}

// runST executes an st command and returns trimmed stdout, or empty on error.
// A non-zero exit yields an empty string, so a segment backed by a command that
// fails or does not exist simply hides itself instead of printing garbage.
func runST(args ...string) string {
	out, err := exec.Command("st", args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
