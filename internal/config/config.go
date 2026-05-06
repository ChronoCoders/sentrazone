package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DBPath        string
	JWTSecret     string
	WGInterface   string
	Port          string
	ControlURL    string
	AuthToken     string
	ServerID      string
	Insecure      bool
	DisableAgent  bool
	HostIP        string
	AdminEmail    string
	AdminPassword string
	// Email alerting
	SMTPHost   string
	SMTPPort   int
	SMTPUser   string
	SMTPPass   string
	AlertEmail      string
	WGEasyPassword string
}

func Load() *Config {
	return &Config{
		DBPath:        getEnv("SENTRA_DB", "sentra.db"),
		JWTSecret:     getEnv("SENTRA_JWT_SECRET", "dev-secret"),
		WGInterface:   getEnv("SENTRA_WG_INTERFACE", "wg0"),
		Port:          getEnv("PORT", "8080"),
		ControlURL:    getEnv("SENTRA_CONTROL_URL", "http://localhost:8080"),
		AuthToken:     getEnv("SENTRA_AUTH_TOKEN", ""),
		ServerID:      getEnv("SENTRA_SERVER_ID", "local"),
		Insecure:      getEnv("SENTRA_INSECURE_SKIP_VERIFY", "false") == "true",
		DisableAgent:  getEnv("SENTRA_DISABLE_AGENT", "false") == "true",
		HostIP:        getEnv("SENTRA_HOST_IP", ""),
		AdminEmail:    getEnv("SENTRA_ADMIN_EMAIL", ""),
		AdminPassword: getEnv("SENTRA_ADMIN_PASSWORD", ""),
		SMTPHost:      getEnv("SENTRA_SMTP_HOST", ""),
		SMTPPort:      getEnvInt("SENTRA_SMTP_PORT", 587),
		SMTPUser:      getEnv("SENTRA_SMTP_USER", ""),
		SMTPPass:      getEnv("SENTRA_SMTP_PASS", ""),
		AlertEmail:    getEnv("SENTRA_ALERT_EMAIL", ""),
		WGEasyPassword: getEnv("SENTRA_WG_EASY_PASSWORD", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.TrimSpace(value)
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if n, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
			return n
		}
	}
	return fallback
}
