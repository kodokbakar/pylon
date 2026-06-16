package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	Kafka     KafkaConfig
	GRPC      GRPCConfig
	WebSocket WebSocketConfig
}

type AppConfig struct {
	Env         string
	Port        string
	LogLevel    string
	CORSOrigins []string
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

type KafkaConfig struct {
	Brokers  []string
	ClientID string
}

type GRPCConfig struct {
	Port string
}

type WebSocketConfig struct {
	MaxConnections  int
	ReadBufferSize  int
	WriteBufferSize int
}

func Load() (*Config, error) {
	_ = loadDotEnv(".env")

	databaseConnMaxLifetime, err := getDuration("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute)
	if err != nil {
		return nil, err
	}

	databaseMaxOpenConns, err := getInt("DATABASE_MAX_OPEN_CONNS", 25)
	if err != nil {
		return nil, err
	}

	databaseMaxIdleConns, err := getInt("DATABASE_MAX_IDLE_CONNS", 5)
	if err != nil {
		return nil, err
	}

	redisDB, err := getInt("REDIS_DB", 0)
	if err != nil {
		return nil, err
	}

	wsMaxConnections, err := getInt("WS_MAX_CONNECTIONS", 1000)
	if err != nil {
		return nil, err
	}

	wsReadBufferSize, err := getInt("WS_READ_BUFFER_SIZE", 4096)
	if err != nil {
		return nil, err
	}

	wsWriteBufferSize, err := getInt("WS_WRITE_BUFFER_SIZE", 4096)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		App: AppConfig{
			Env:         getString("APP_ENV", "development"),
			Port:        getString("APP_PORT", "8080"),
			LogLevel:    getString("APP_LOG_LEVEL", "info"),
			CORSOrigins: getStringSlice("CORS_ORIGINS", []string{"*"}),
		},
		Database: DatabaseConfig{
			URL:             getString("DATABASE_URL", ""),
			MaxOpenConns:    databaseMaxOpenConns,
			MaxIdleConns:    databaseMaxIdleConns,
			ConnMaxLifetime: databaseConnMaxLifetime,
		},
		Redis: RedisConfig{
			URL:      getString("REDIS_URL", "redis://localhost:6379"),
			Password: getString("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		Kafka: KafkaConfig{
			Brokers:  getStringSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			ClientID: getString("KAFKA_CLIENT_ID", "pylon"),
		},
		GRPC: GRPCConfig{
			Port: getString("GRPC_PORT", "9000"),
		},
		WebSocket: WebSocketConfig{
			MaxConnections:  wsMaxConnections,
			ReadBufferSize:  wsReadBufferSize,
			WriteBufferSize: wsWriteBufferSize,
		},
	}

	return cfg, nil
}

func getString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getInt(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s as int: %w", key, err)
	}

	return parsed, nil
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s as duration: %w", key, err)
	}

	return parsed, nil
}

func getStringSlice(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			result = append(result, item)
		}
	}

	if len(result) == 0 {
		return fallback
	}

	return result
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open dotenv file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)

		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set dotenv value for %s: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan dotenv file: %w", err)
	}

	return nil
}
