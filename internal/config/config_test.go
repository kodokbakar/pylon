package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("APP_PORT", "")
	t.Setenv("APP_LOG_LEVEL", "")
	t.Setenv("CORS_ORIGINS", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "")
	t.Setenv("DATABASE_MAX_IDLE_CONNS", "")
	t.Setenv("DATABASE_CONN_MAX_LIFETIME", "")
	t.Setenv("REDIS_URL", "")
	t.Setenv("REDIS_DB", "")
	t.Setenv("KAFKA_BROKERS", "")
	t.Setenv("KAFKA_CLIENT_ID", "")
	t.Setenv("GRPC_PORT", "")
	t.Setenv("WS_MAX_CONNECTIONS", "")
	t.Setenv("WS_READ_BUFFER_SIZE", "")
	t.Setenv("WS_WRITE_BUFFER_SIZE", "")
	t.Setenv("WS_INSECURE_SKIP_VERIFY", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("JWT_EXPIRY", "")
	t.Setenv("JWT_REFRESH_EXPIRY", "")
	t.Setenv("CHAT_SERVICE_URL", "")
	t.Setenv("PRESENCE_SERVICE_URL", "")
	t.Setenv("ROOM_SERVICE_URL", "")
	t.Setenv("NOTIFICATION_SERVICE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.App.Env != "development" {
		t.Fatalf("expected default app env, got %q", cfg.App.Env)
	}

	if cfg.App.Port != "8080" {
		t.Fatalf("expected default app port, got %q", cfg.App.Port)
	}

	if len(cfg.App.CORSOrigins) != 1 || cfg.App.CORSOrigins[0] != "*" {
		t.Fatalf("expected default cors origins wildcard, got %#v", cfg.App.CORSOrigins)
	}

	if cfg.Database.MaxOpenConns != 25 {
		t.Fatalf("expected default max open conns 25, got %d", cfg.Database.MaxOpenConns)
	}

	if cfg.Database.ConnMaxLifetime != 5*time.Minute {
		t.Fatalf("expected default conn max lifetime 5m, got %s", cfg.Database.ConnMaxLifetime)
	}

	if len(cfg.Kafka.Brokers) != 1 || cfg.Kafka.Brokers[0] != "localhost:9092" {
		t.Fatalf("expected default kafka broker, got %#v", cfg.Kafka.Brokers)
	}

	if cfg.WebSocket.InsecureSkipVerify {
		t.Fatal("expected websocket insecure skip verify false by default")
	}

	if cfg.JWT.Secret != "dev-secret-change-me" {
		t.Fatalf("expected default jwt secret, got %q", cfg.JWT.Secret)
	}

	if cfg.JWT.Expiry != 24*time.Hour {
		t.Fatalf("expected default jwt expiry 24h, got %s", cfg.JWT.Expiry)
	}

	if cfg.Services.ChatURL != "http://localhost:9001" {
		t.Fatalf("expected default chat service url, got %q", cfg.Services.ChatURL)
	}
}

func TestLoadUsesEnvironmentValues(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_PORT", "8088")
	t.Setenv("APP_LOG_LEVEL", "warn")
	t.Setenv("CORS_ORIGINS", "https://app.example.com, https://admin.example.com")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pylon?sslmode=disable")
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "50")
	t.Setenv("DATABASE_MAX_IDLE_CONNS", "10")
	t.Setenv("DATABASE_CONN_MAX_LIFETIME", "10m")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("REDIS_PASSWORD", "secret")
	t.Setenv("REDIS_DB", "2")
	t.Setenv("KAFKA_BROKERS", "localhost:9092, localhost:9093")
	t.Setenv("KAFKA_CLIENT_ID", "pylon-test")
	t.Setenv("GRPC_PORT", "9001")
	t.Setenv("WS_MAX_CONNECTIONS", "2000")
	t.Setenv("WS_READ_BUFFER_SIZE", "8192")
	t.Setenv("WS_WRITE_BUFFER_SIZE", "8192")
	t.Setenv("WS_INSECURE_SKIP_VERIFY", "true")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_EXPIRY", "30m")
	t.Setenv("JWT_REFRESH_EXPIRY", "48h")
	t.Setenv("CHAT_SERVICE_URL", "http://chat:9001")
	t.Setenv("PRESENCE_SERVICE_URL", "http://presence:9002")
	t.Setenv("ROOM_SERVICE_URL", "http://room:9003")
	t.Setenv("NOTIFICATION_SERVICE_URL", "http://notification:9004")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.App.Env != "production" {
		t.Fatalf("expected production env, got %q", cfg.App.Env)
	}

	if len(cfg.App.CORSOrigins) != 2 {
		t.Fatalf("expected 2 cors origins, got %#v", cfg.App.CORSOrigins)
	}

	if cfg.App.CORSOrigins[0] != "https://app.example.com" {
		t.Fatalf("expected first cors origin https://app.example.com, got %q", cfg.App.CORSOrigins[0])
	}

	if cfg.Database.MaxOpenConns != 50 {
		t.Fatalf("expected max open conns 50, got %d", cfg.Database.MaxOpenConns)
	}

	if cfg.Database.ConnMaxLifetime != 10*time.Minute {
		t.Fatalf("expected conn max lifetime 10m, got %s", cfg.Database.ConnMaxLifetime)
	}

	if cfg.Redis.DB != 2 {
		t.Fatalf("expected redis db 2, got %d", cfg.Redis.DB)
	}

	if len(cfg.Kafka.Brokers) != 2 {
		t.Fatalf("expected 2 kafka brokers, got %#v", cfg.Kafka.Brokers)
	}

	if cfg.WebSocket.MaxConnections != 2000 {
		t.Fatalf("expected ws max connections 2000, got %d", cfg.WebSocket.MaxConnections)
	}

	if !cfg.WebSocket.InsecureSkipVerify {
		t.Fatal("expected websocket insecure skip verify true")
	}

	if cfg.JWT.Secret != "test-secret" {
		t.Fatalf("expected jwt secret test-secret, got %q", cfg.JWT.Secret)
	}

	if cfg.JWT.Expiry != 30*time.Minute {
		t.Fatalf("expected jwt expiry 30m, got %s", cfg.JWT.Expiry)
	}

	if cfg.JWT.RefreshExpiry != 48*time.Hour {
		t.Fatalf("expected jwt refresh expiry 48h, got %s", cfg.JWT.RefreshExpiry)
	}

	if cfg.Services.RoomURL != "http://room:9003" {
		t.Fatalf("expected room service url http://room:9003, got %q", cfg.Services.RoomURL)
	}
}

func TestLoadReturnsErrorForInvalidInt(t *testing.T) {
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadReturnsErrorForInvalidDuration(t *testing.T) {
	t.Setenv("DATABASE_CONN_MAX_LIFETIME", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadDotEnvSupportsExportPrefix(t *testing.T) {
	key := "PYLON_TEST_EXPORT_PREFIX"
	t.Setenv(key, "")

	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	content := "export PYLON_TEST_EXPORT_PREFIX=enabled\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write dotenv file: %v", err)
	}

	if err := loadDotEnv(path); err != nil {
		t.Fatalf("load dotenv file: %v", err)
	}

	if got := os.Getenv(key); got != "" {
		t.Fatalf("expected existing env value to be respected, got %q", got)
	}
}

func TestLoadDotEnvSetsExportValueWhenEnvDoesNotExist(t *testing.T) {
	key := "PYLON_TEST_EXPORT_VALUE"
	t.Setenv(key, "")
	os.Unsetenv(key)

	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	content := "export PYLON_TEST_EXPORT_VALUE=enabled\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write dotenv file: %v", err)
	}

	if err := loadDotEnv(path); err != nil {
		t.Fatalf("load dotenv file: %v", err)
	}

	if got := os.Getenv(key); got != "enabled" {
		t.Fatalf("expected dotenv value enabled, got %q", got)
	}
}

func TestLoadReturnsErrorForInvalidJWTExpiry(t *testing.T) {
	t.Setenv("JWT_EXPIRY", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadReturnsErrorForInvalidWebSocketInsecureSkipVerify(t *testing.T) {
	t.Setenv("WS_INSECURE_SKIP_VERIFY", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
