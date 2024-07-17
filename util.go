package goprivateproxy

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// GetConfig fetches configuration details from config.yaml.
func GetConfig() Config {
	// Define the config file path
	var configFile string
	flag.StringVar(&configFile, "config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	// Read the configuration file
	configData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	// Parse the configuration file
	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	return config
}

// isRobot checks if the User-Agent belongs to a bot and if it's permitted.
func isRobot(userAgent string, botBlockList, permittedBots []string) (bool, bool) {
	isBot := false
	isPermittedBot := false

	userAgentLower := strings.ToLower(userAgent)

	for _, bot := range permittedBots {
		if strings.Contains(userAgentLower, bot) {
			isBot = true
			isPermittedBot = true
			break
		}
	}

	if !isPermittedBot {
		for _, signature := range botBlockList {
			if strings.Contains(userAgentLower, signature) {
				isBot = true
				break
			}
		}
	}

	return isBot, isPermittedBot
}

// getClientIP extracts the client's IP address from the request headers.
func getClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		split := strings.Split(xForwardedFor, ",")
		if len(split) > 0 {
			return strings.TrimSpace(split[0])
		}
		return xForwardedFor
	}

	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	return r.RemoteAddr
}
