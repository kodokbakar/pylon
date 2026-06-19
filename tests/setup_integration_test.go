//go:build integration

package tests

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/kodokbakar/pylon/internal/config"
	apigateway "github.com/kodokbakar/pylon/services/api-gateway"
	chatserver "github.com/kodokbakar/pylon/services/chat-service"
	notificationserver "github.com/kodokbakar/pylon/services/notification-service"
	presenceserver "github.com/kodokbakar/pylon/services/presence-service"
	roomserver "github.com/kodokbakar/pylon/services/room-service"
)

const (
	integrationGatewayPort      = "18080"
	integrationChatPort         = "19001"
	integrationPresencePort     = "19002"
	integrationRoomPort         = "19003"
	integrationNotificationPort = "19004"

	integrationJWTSecret = "integration-test-secret"
)

type integrationSuite struct {
	GatewayURL      string
	ChatURL         string
	PresenceURL     string
	RoomURL         string
	NotificationURL string

	HTTPClient *http.Client
}

func setupIntegrationSuite(t *testing.T) *integrationSuite {
	t.Helper()

	baseDatabaseURL := getenv("PYLON_TEST_DATABASE_URL", "postgres://pylon:pylon_dev@localhost:5433/pylon?sslmode=disable")
	redisURL := getenv("PYLON_TEST_REDIS_URL", "redis://localhost:6380")
	kafkaBroker := getenv("PYLON_TEST_KAFKA_BROKER", "localhost:9092")

	requirePostgres(t, baseDatabaseURL)
	requireRedis(t, redisURL)
	requireKafka(t, kafkaBroker)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	schema := fmt.Sprintf("pylon_it_%d", time.Now().UnixNano())
	databaseURL := setupTestSchema(t, ctx, baseDatabaseURL, schema)

	baseCfg := newIntegrationConfig(databaseURL, redisURL, kafkaBroker)

	chatCfg := cloneConfig(baseCfg)
	chatCfg.GRPC.Port = integrationChatPort
	chatSrv, err := chatserver.New(ctx, chatCfg)
	if err != nil {
		t.Fatalf("create chat service: %v", err)
	}
	startManagedServer(t, "chat-service", "http://127.0.0.1:"+integrationChatPort+"/health", chatSrv.Start, chatSrv.Shutdown)

	roomCfg := cloneConfig(baseCfg)
	roomCfg.GRPC.Port = integrationRoomPort
	roomSrv, err := roomserver.New(ctx, roomCfg)
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}
	startManagedServer(t, "room-service", "http://127.0.0.1:"+integrationRoomPort+"/health", roomSrv.Start, roomSrv.Shutdown)

	presenceCfg := cloneConfig(baseCfg)
	presenceCfg.GRPC.Port = integrationPresencePort
	presenceSrv, err := presenceserver.New(ctx, presenceCfg)
	if err != nil {
		t.Fatalf("create presence service: %v", err)
	}
	startManagedServer(t, "presence-service", "http://127.0.0.1:"+integrationPresencePort+"/health", presenceSrv.Start, presenceSrv.Shutdown)

	notificationCfg := cloneConfig(baseCfg)
	notificationCfg.GRPC.Port = integrationNotificationPort
	notificationSrv, err := notificationserver.New(ctx, notificationCfg)
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}
	startManagedServer(t, "notification-service", "http://127.0.0.1:"+integrationNotificationPort+"/health", notificationSrv.Start, notificationSrv.Shutdown)

	gatewayCfg := cloneConfig(baseCfg)
	gatewayCfg.App.Port = integrationGatewayPort
	gatewaySrv, err := apigateway.New(gatewayCfg)
	if err != nil {
		t.Fatalf("create api gateway: %v", err)
	}
	startManagedServer(t, "api-gateway", "http://127.0.0.1:"+integrationGatewayPort+"/health", gatewaySrv.Start, gatewaySrv.Shutdown)

	return &integrationSuite{
		GatewayURL:      "http://127.0.0.1:" + integrationGatewayPort,
		ChatURL:         "http://127.0.0.1:" + integrationChatPort,
		PresenceURL:     "http://127.0.0.1:" + integrationPresencePort,
		RoomURL:         "http://127.0.0.1:" + integrationRoomPort,
		NotificationURL: "http://127.0.0.1:" + integrationNotificationPort,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func newIntegrationConfig(databaseURL, redisURL, kafkaBroker string) *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Env:      "test",
			Port:     integrationGatewayPort,
			LogLevel: "debug",
			CORSOrigins: []string{
				"http://localhost:5173",
				"http://127.0.0.1:" + integrationGatewayPort,
			},
		},
		Database: config.DatabaseConfig{
			URL:             databaseURL,
			MaxOpenConns:    10,
			MaxIdleConns:    2,
			ConnMaxLifetime: time.Minute,
		},
		Redis: config.RedisConfig{
			URL: redisURL,
		},
		Kafka: config.KafkaConfig{
			Brokers:  []string{kafkaBroker},
			ClientID: "pylon-integration-test",
		},
		GRPC: config.GRPCConfig{
			Port: integrationChatPort,
		},
		WebSocket: config.WebSocketConfig{
			MaxConnections:     100,
			ReadBufferSize:     4096,
			WriteBufferSize:    4096,
			InsecureSkipVerify: true,
		},
		JWT: config.JWTConfig{
			Secret:        integrationJWTSecret,
			Expiry:        time.Hour,
			RefreshExpiry: 24 * time.Hour,
		},
		Services: config.ServicesConfig{
			ChatURL:         "http://127.0.0.1:" + integrationChatPort,
			PresenceURL:     "http://127.0.0.1:" + integrationPresencePort,
			RoomURL:         "http://127.0.0.1:" + integrationRoomPort,
			NotificationURL: "http://127.0.0.1:" + integrationNotificationPort,
		},
	}
}

func cloneConfig(cfg *config.Config) *config.Config {
	copyCfg := *cfg
	copyCfg.App.CORSOrigins = append([]string(nil), cfg.App.CORSOrigins...)
	copyCfg.Kafka.Brokers = append([]string(nil), cfg.Kafka.Brokers...)
	return &copyCfg
}

func setupTestSchema(t *testing.T, ctx context.Context, baseDatabaseURL, schema string) string {
	t.Helper()

	adminPool, err := pgxpool.New(ctx, baseDatabaseURL)
	if err != nil {
		t.Fatalf("connect postgres admin pool: %v", err)
	}

	if err := adminPool.Ping(ctx); err != nil {
		t.Fatalf("ping postgres admin pool: %v", err)
	}

	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+quoteIdent(schema)); err != nil {
		t.Fatalf("create test schema: %v", err)
	}

	t.Cleanup(func() {
		defer adminPool.Close()

		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err := adminPool.Exec(cleanupCtx, "DROP SCHEMA IF EXISTS "+quoteIdent(schema)+" CASCADE"); err != nil {
			t.Logf("drop test schema %s: %v", schema, err)
		}
	})

	testDatabaseURL := databaseURLWithSearchPath(t, baseDatabaseURL, schema)

	testPool, err := pgxpool.New(ctx, testDatabaseURL)
	if err != nil {
		t.Fatalf("connect postgres test pool: %v", err)
	}
	defer testPool.Close()

	runMigrations(t, ctx, testPool)

	return testDatabaseURL
}

func runMigrations(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	migrationDir := filepath.Join(projectRoot(t), "migrations")

	files, err := filepath.Glob(filepath.Join(migrationDir, "*.up.sql"))
	if err != nil {
		t.Fatalf("list migrations: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("no migration files found in %s", migrationDir)
	}

	sort.Strings(files)

	for _, file := range files {
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read migration %s: %v", file, err)
		}

		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			t.Fatalf("run migration %s: %v", filepath.Base(file), err)
		}
	}
}

func startManagedServer(
	t *testing.T,
	name string,
	healthURL string,
	start func() error,
	shutdown func(context.Context) error,
) {
	t.Helper()

	errCh := make(chan error, 1)
	go func() {
		errCh <- start()
	}()

	waitForHealth(t, name, healthURL, errCh)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := shutdown(ctx); err != nil {
			t.Errorf("shutdown %s: %v", name, err)
		}

		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("%s stopped with error: %v", name, err)
			}
		case <-time.After(5 * time.Second):
			t.Errorf("timeout waiting for %s to stop", name)
		}
	})
}

func waitForHealth(t *testing.T, name, healthURL string, errCh <-chan error) {
	t.Helper()

	deadline := time.Now().Add(30 * time.Second)
	client := &http.Client{Timeout: time.Second}

	for time.Now().Before(deadline) {
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("%s failed before becoming healthy: %v", name, err)
			}
			t.Fatalf("%s stopped before becoming healthy", name)
		default:
		}

		resp, err := client.Get(healthURL)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("%s did not become healthy at %s", name, healthURL)
}

func requirePostgres(t *testing.T, databaseURL string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Skipf("postgres unavailable, run `make dev` first: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("postgres unavailable, run `make dev` first: %v", err)
	}
}

func requireRedis(t *testing.T, redisURL string) {
	t.Helper()

	options, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("parse redis url: %v", err)
	}

	client := redis.NewClient(options)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("close redis client: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("redis unavailable, run `make dev` first: %v", err)
	}
}

func requireKafka(t *testing.T, broker string) {
	t.Helper()

	conn, err := net.DialTimeout("tcp", broker, 3*time.Second)
	if err != nil {
		t.Skipf("kafka unavailable, run `make dev` first: %v", err)
	}

	if err := conn.Close(); err != nil {
		t.Logf("close kafka readiness connection: %v", err)
	}
}

func databaseURLWithSearchPath(t *testing.T, rawURL, schema string) string {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse database url: %v", err)
	}

	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()

	return parsed.String()
}

func quoteIdent(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func projectRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	if filepath.Base(wd) == "tests" {
		return filepath.Dir(wd)
	}

	return wd
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}
