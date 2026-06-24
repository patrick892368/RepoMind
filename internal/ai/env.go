package ai

import (
	"os"
	"strings"
)

func lookupAPIKey(envPath string, names ...string) string {
	return lookupEnvValue(envPath, names...)
}

func lookupEnvValue(envPath string, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	values := readDotEnv(envPath)
	for _, name := range names {
		if value := strings.TrimSpace(values[name]); value != "" {
			return value
		}
	}
	return ""
}

func readDotEnv(path string) map[string]string {
	result := map[string]string{}
	if path == "" {
		return result
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return result
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key != "" {
			result[key] = value
		}
	}
	return result
}
