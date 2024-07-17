// Package privatelinkproxy is a Go-based reverse proxy and AWS PrivateLink replacement
// running as a systemd service on EC2 instances. Includes rate limiting, bot blocking,
// and dynamic routing based on subdomains and paths.
package goprivateproxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"golang.org/x/time/rate"
)

// Server represents the reverse proxy server.
type Server struct {
	Config Config
}

// Config holds the server configuration.
type Config struct {
	Port          int           `yaml:"port"`
	Limiter       LimiterConfig `yaml:"limiter"`
	BotBlockList  []string      `yaml:"botBlockList"`
	PermittedBots []string      `yaml:"permittedBots"`
	OtherRoutes   []Route       `yaml:"otherRoutes"`
	DefaultRoute  Route         `yaml:"defaultRoute"`
}

// LimiterConfig represents the rate limiter configuration.
type LimiterConfig struct {
	RPS   int `yaml:"rps"`
	Burst int `yaml:"burst"`
}

// Route represents the target IP, port, and path for a route.
type Route struct {
	IP   string `yaml:"ip"`
	Port int    `yaml:"port"`
	Path string `yaml:"path"`
}

// NewServer initializes a new Server instance.
func NewServer(config Config) *Server {
	return &Server{
		Config: config,
	}
}

// Start runs the HTTP server.
func (s *Server) Start() {
	http.Handle("/", s.rateLimiter(s.botChecker(http.HandlerFunc(s.handler))))
	http.HandleFunc("/health-check", s.healthCheckHandler)

	log.Printf("Starting EC2 Reverse Proxy + Private Link on port %d", s.Config.Port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.Config.Port), nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// healthCheckHandler responds with a 200 OK status for health checks.
func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Health check")
	w.WriteHeader(http.StatusOK)
}

// handler is the main reverse proxy handler that checks rate limits and blocks bots.
func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	clientIPs := getClientIP(r)
	targetURL := s.getTargetURL(r)
	if targetURL == "" {
		http.Error(w, "No target URL configured", http.StatusNotFound)
		return
	}

	url, _ := url.Parse(targetURL)
	proxy := httputil.NewSingleHostReverseProxy(url)

	log.Printf("Forwarding request to targetURL: %s; requestURL: [%s]; ip-address: %s", targetURL, r.URL, clientIPs)
	proxy.ServeHTTP(w, r)
}

// getTargetURL returns the target URL based on the request's subdomain or path.
func (s *Server) getTargetURL(r *http.Request) string {
	host := r.Host
	path := r.URL.Path

	if len(s.Config.OtherRoutes) > 0 {
		subdomain := strings.Split(host, ".")[0]
		for _, route := range s.Config.OtherRoutes {
			if subdomain == route.Path || strings.HasPrefix(path, route.Path) {
				return fmt.Sprintf("http://%s:%d", route.IP, route.Port)
			}
		}
	}

	return fmt.Sprintf("http://%s:%d", s.Config.DefaultRoute.IP, s.Config.DefaultRoute.Port)
}

// botChecker checks if the User-Agent belongs to a bot and blocks if necessary.
func (s *Server) botChecker(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.UserAgent()
		if isBot, _ := isRobot(userAgent, s.Config.BotBlockList, s.Config.PermittedBots); isBot {
			http.Error(w, "Access Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// rateLimiter applies rate limiting to requests.
func (s *Server) rateLimiter(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(s.Config.Limiter.RPS), s.Config.Limiter.Burst)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Calm down, too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
