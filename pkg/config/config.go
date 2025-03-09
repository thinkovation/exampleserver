package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port string

	// Auth
	JWTSecret []byte
	APIKeys   []string

	// Logging
	LogDir        string
	LogFile       string
	LogMaxSize    int
	LogMaxAge     int
	LogMaxBackups int
	LogCompress   bool

	// Datadog
	DatadogEnabled bool
	DatadogService string
	DatadogEnv     string

	// Stats
	StatsInterval time.Duration
}

func Load() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	// Determine default log directory based on OS
	defaultLogDir := "logs"
	if runtime.GOOS == "linux" {
		defaultLogDir = "/var/log/app"
	}

	// Get log directory from env or use default
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = defaultLogDir
	}

	// Ensure log directory is absolute
	logDir, err := filepath.Abs(logDir)
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:      getEnvDefault("PORT", "8080"),
		JWTSecret: []byte(getEnvDefault("JWT_SECRET", "your-secret-key")),
		APIKeys:   getAPIKeys(),

		// Logging
		LogDir:        logDir,
		LogFile:       filepath.Join(logDir, "app.log"),
		LogMaxSize:    getEnvIntDefault("LOG_MAX_SIZE", 10),    // 10 MB
		LogMaxAge:     getEnvIntDefault("LOG_MAX_AGE", 30),     // 30 days
		LogMaxBackups: getEnvIntDefault("LOG_MAX_BACKUPS", 5),  // 5 backups
		LogCompress:   getEnvBoolDefault("LOG_COMPRESS", true), // compress by default

		// Datadog
		DatadogEnabled: getEnvBoolDefault("DD_ENABLED", false),
		DatadogService: getEnvDefault("DD_SERVICE", "example-server"),
		DatadogEnv:     getEnvDefault("DD_ENV", "development"),

		// Stats
		StatsInterval: time.Duration(getEnvIntDefault("STATS_INTERVAL", 60)) * time.Second,
	}, nil
}

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBoolDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getAPIKeys() []string {
	apiKeys := os.Getenv("API_KEYS")
	if apiKeys == "" {
		return []string{"default-dev-key"}
	}
	return filepath.SplitList(apiKeys)
}
