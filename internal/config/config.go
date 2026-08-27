package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultServerAddress = ":8080"
	defaultDBHost        = "localhost:3306"
	defaultDBUser        = "root"
	defaultDBName        = "sharing_vision_test"
)

// Config contains runtime settings for the service.
type Config struct {
	ServerAddress     string
	DBHost            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	QueryTimeout      time.Duration
	ShutdownTimeout   time.Duration
}

// Load reads .env, environment variables, and applies local defaults.
func Load() Config {
	// Support running from the project root, cmd/, or an internal package.
	// Missing files are ignored so deployed environments can inject variables.
	for _, filename := range []string{".env", "../.env", "../../.env"} {
		if err := godotenv.Load(filename); err == nil {
			break
		}
	}

	return Config{
		ServerAddress:     getEnv("SERVER_ADDRESS", defaultServerAddress),
		DBHost:            getEnv("DB_HOST", defaultDBHost),
		DBUser:            getEnv("DB_USER", defaultDBUser),
		DBPassword:        getEnv("DB_PASSWORD", ""),
		DBName:            getEnv("DB_NAME", defaultDBName),
		DBMaxOpenConns:    getPositiveIntEnv("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getPositiveIntEnv("DB_MAX_IDLE_CONNS", 25),
		DBConnMaxLifetime: 5 * time.Minute,
		QueryTimeout:      3 * time.Second,
		ShutdownTimeout:   10 * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getPositiveIntEnv(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
