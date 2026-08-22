package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// loader reads values from the process environment and accumulates the
// problems it finds, so a misconfigured deployment reports every missing or
// malformed variable at once instead of one per restart.
type loader struct {
	problems []string
}

func (l *loader) fail(key, reason string) {
	l.problems = append(l.problems, fmt.Sprintf("%s: %s", key, reason))
}

func (l *loader) err() error {
	if len(l.problems) == 0 {
		return nil
	}
	return fmt.Errorf("invalid configuration:\n  - %s", strings.Join(l.problems, "\n  - "))
}

func (l *loader) str(key, fallback string) string {
	if v, ok := lookup(key); ok {
		return v
	}
	return fallback
}

// required returns the value of key, recording a problem when it is absent.
// Secrets use this so that an empty deployment fails loudly at startup.
func (l *loader) required(key string) string {
	v, ok := lookup(key)
	if !ok {
		l.fail(key, "is required")
		return ""
	}
	return v
}

func (l *loader) int(key string, fallback int) int {
	v, ok := lookup(key)
	if !ok {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		l.fail(key, fmt.Sprintf("must be an integer, got %q", v))
		return fallback
	}
	return parsed
}

func (l *loader) int64(key string, fallback int64) int64 {
	v, ok := lookup(key)
	if !ok {
		return fallback
	}
	parsed, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		l.fail(key, fmt.Sprintf("must be an integer, got %q", v))
		return fallback
	}
	return parsed
}

func (l *loader) duration(key string, fallback time.Duration) time.Duration {
	v, ok := lookup(key)
	if !ok {
		return fallback
	}
	parsed, err := time.ParseDuration(v)
	if err != nil {
		l.fail(key, fmt.Sprintf("must be a duration such as 30s or 15m, got %q", v))
		return fallback
	}
	return parsed
}

func (l *loader) durations(key string, fallback []time.Duration) []time.Duration {
	v, ok := lookup(key)
	if !ok {
		return fallback
	}
	parts := splitList(v)
	if len(parts) == 0 {
		return fallback
	}
	out := make([]time.Duration, 0, len(parts))
	for _, part := range parts {
		parsed, err := time.ParseDuration(part)
		if err != nil {
			l.fail(key, fmt.Sprintf("must be a comma-separated duration list, got %q", v))
			return fallback
		}
		out = append(out, parsed)
	}
	return out
}

func (l *loader) list(key string, fallback []string) []string {
	v, ok := lookup(key)
	if !ok {
		return fallback
	}
	parts := splitList(v)
	if len(parts) == 0 {
		return fallback
	}
	return parts
}

func lookup(key string) (string, bool) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return "", false
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return "", false
	}
	return v, true
}

func splitList(v string) []string {
	raw := strings.Split(v, ",")
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

// LoadDotEnv reads KEY=VALUE pairs from path into the environment without
// overwriting variables that are already set, so real environment variables
// always win over the development file. A missing file is not an error.
func LoadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}
