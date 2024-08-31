package routes

import (
	"log"
	"os"
	"strings"
	"sync"
)

var (
	allowedEndpoints  map[string]bool
	mu                sync.RWMutex
	inMaintenanceMode bool
)

func loadAllowedEndpoints(filePath string) error {
	mu.Lock()
	defer mu.Unlock()

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")

	if len(lines) > 0 && strings.HasPrefix(lines[0], "MAINTENANCE=") {
		inMaintenanceMode = strings.TrimPrefix(lines[0], "MAINTENANCE=") == "ON"
		lines = lines[1:]
	}

	allowed := make(map[string]bool)

	for _, line := range lines {
		if strings.HasPrefix(line, "//") || inMaintenanceMode {
			continue
		}

		line = strings.TrimSpace(line)

		parts := strings.SplitN(line, ",", 3)
		if len(parts) != 3 {
			continue
		}

		method, context, endpoint := parts[0], parts[1], parts[2]

		if method != "GET" && method != "POST" {
			log.Println("Invalid method in allowed endpoints:", method)
			continue
		}

		if context != "public" && context != "protected" {
			log.Println("Invalid context in allowed endpoints:", context)
			continue
		}

		key := method + "," + context + "," + endpoint
		allowed[key] = true
	}

	allowedEndpoints = allowed
	return nil
}
