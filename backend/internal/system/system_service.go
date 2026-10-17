// Package system exposes operational health diagnostics for the admin back office.
// It probes each backing dependency (PostgreSQL, Redis, Kafka, MinIO) with a real
// round-trip and reports live connection status, latency, and—where meaningful—
// queue lag, so operators see the actual heartbeat of the platform rather than
// static labels.
package system

import (
	"context"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

// Status values reported per dependency and for the aggregated report.
const (
	// StatusHealthy indicates the dependency answered within the probe budget.
	StatusHealthy = "healthy"
	// StatusDegraded indicates the dependency is reachable but not fully ready.
	StatusDegraded = "degraded"
	// StatusUnhealthy indicates the dependency could not be reached.
	StatusUnhealthy = "unhealthy"
)

const (
	// probeTimeout bounds each individual dependency check so a single slow
	// backend cannot stall the whole report.
	probeTimeout = 3 * time.Second

	// Default ports used when a configured endpoint cannot be parsed.
	defaultPostgresPort = 5432
	defaultKafkaPort    = 9092
	defaultMinioPort    = 9000
)

// Config carries the dependency targets and display metadata the health service
// probes. Endpoints are parsed lazily so callers can pass raw config values.
type Config struct {
	// KafkaBrokers is the list of seed broker host:port addresses.
	KafkaBrokers []string
	// KafkaGroupID is the notification worker consumer group used for lag.
	KafkaGroupID string
	// NotificationsTopic is the topic inspected for connectivity and lag.
	NotificationsTopic string
	// Bucket is the object storage bucket whose existence is verified.
	Bucket string
	// PostgresPort is the configured database port (string form).
	PostgresPort string
	// RedisPort is the configured Redis port.
	RedisPort int
	// MinioEndpoint is the configured object storage endpoint (host:port or URL).
	MinioEndpoint string
}

// ServiceHealth describes the observed state of a single backing service.
type ServiceHealth struct {
	// Key is the stable identifier ("postgres", "redis", "kafka", "minio") that
	// the client uses to select an icon and localized labels.
	Key string `json:"key"`
	// Status is one of StatusHealthy, StatusDegraded, or StatusUnhealthy.
	Status string `json:"status"`
	// Port is the network port the service listens on.
	Port int `json:"port"`
	// LatencyMs is the measured round-trip time in milliseconds.
	LatencyMs float64 `json:"latencyMs"`
	// Lag is the total consumer lag for the notifications topic, when known.
	Lag *int64 `json:"lag,omitempty"`
	// Error carries the failure reason when the service is not healthy.
	Error string `json:"error,omitempty"`
}

// HealthReport is the aggregated health snapshot returned to the admin client.
type HealthReport struct {
	// Status is the worst-case rollup across all services.
	Status string `json:"status"`
	// CheckedAt is the UTC timestamp at which the probes completed.
	CheckedAt time.Time `json:"checkedAt"`
	// Services holds one entry per probed dependency, in a stable order.
	Services []ServiceHealth `json:"services"`
}

// Service owns the dependency clients used to produce health reports.
type Service struct {
	db        *database.DB
	redis     *redis.Client
	minio     *minio.Client
	kafkaPort int
	minioPort int
	pgPort    int
	cfg       Config
}

// NewService constructs a health service over the supplied dependency clients.
func NewService(db *database.DB, redisClient *redis.Client, minioClient *minio.Client, cfg Config) *Service {
	return &Service{
		db:        db,
		redis:     redisClient,
		minio:     minioClient,
		kafkaPort: firstBrokerPort(cfg.KafkaBrokers, defaultKafkaPort),
		minioPort: portFromEndpoint(cfg.MinioEndpoint, defaultMinioPort),
		pgPort:    atoiDefault(cfg.PostgresPort, defaultPostgresPort),
		cfg:       cfg,
	}
}

// Check probes every dependency concurrently and returns the aggregated report.
// A failing dependency never fails the request: its state is captured per-service
// so operators can see exactly which component is down.
func (s *Service) Check(ctx context.Context) *HealthReport {
	services := make([]ServiceHealth, 4)
	var wg sync.WaitGroup

	probes := []func(context.Context) ServiceHealth{
		s.checkPostgres,
		s.checkRedis,
		s.checkKafka,
		s.checkMinio,
	}

	for i, probe := range probes {
		wg.Add(1)
		go func(idx int, fn func(context.Context) ServiceHealth) {
			defer wg.Done()
			services[idx] = fn(ctx)
		}(i, probe)
	}
	wg.Wait()

	return &HealthReport{
		Status:    rollupStatus(services),
		CheckedAt: time.Now().UTC(),
		Services:  services,
	}
}

// checkPostgres verifies the database answers a ping.
func (s *Service) checkPostgres(ctx context.Context) ServiceHealth {
	health := ServiceHealth{Key: "postgres", Status: StatusHealthy, Port: s.pgPort}
	if s.db == nil {
		health.Status = StatusUnhealthy
		health.Error = "database client not configured"
		return health
	}

	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	start := time.Now()
	err := s.db.Ping(ctx)
	health.LatencyMs = elapsedMs(start)

	if err != nil {
		health.Status = StatusUnhealthy
		health.Error = err.Error()
	}
	return health
}

// checkRedis verifies the cache/session store answers a ping.
func (s *Service) checkRedis(ctx context.Context) ServiceHealth {
	health := ServiceHealth{Key: "redis", Status: StatusHealthy, Port: s.cfg.RedisPort}
	if s.redis == nil {
		health.Status = StatusUnhealthy
		health.Error = "redis client not configured"
		return health
	}

	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	start := time.Now()
	err := s.redis.Ping(ctx).Err()
	health.LatencyMs = elapsedMs(start)

	if err != nil {
		health.Status = StatusUnhealthy
		health.Error = err.Error()
	}
	return health
}

// checkKafka verifies broker/topic reachability and, when possible, reports the
// consumer lag of the notifications topic.
func (s *Service) checkKafka(ctx context.Context) ServiceHealth {
	health := ServiceHealth{Key: "kafka", Status: StatusHealthy, Port: s.kafkaPort}
	if len(s.cfg.KafkaBrokers) == 0 {
		health.Status = StatusUnhealthy
		health.Error = "no kafka brokers configured"
		return health
	}

	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	// Dial the broker directly and read topic metadata over that fresh connection.
	// This deliberately bypasses kafka-go's shared transport metadata cache, which
	// would otherwise report a stale "healthy" cluster for the cache TTL after a
	// broker goes away.
	start := time.Now()
	conn, err := kafka.DialContext(ctx, "tcp", s.cfg.KafkaBrokers[0])
	if err != nil {
		health.Status = StatusUnhealthy
		health.Error = err.Error()
		return health
	}
	partitions, err := conn.ReadPartitions(s.cfg.NotificationsTopic)
	_ = conn.Close()
	health.LatencyMs = elapsedMs(start)

	if err != nil {
		health.Status = StatusUnhealthy
		health.Error = err.Error()
		return health
	}

	if len(partitions) == 0 {
		health.Status = StatusDegraded
		health.Error = "notifications topic not found"
		return health
	}

	if lag, ok := s.kafkaLag(ctx, partitions); ok {
		health.Lag = &lag
	}
	return health
}

// kafkaLag sums (high watermark - committed offset) across the supplied topic
// partitions for the configured consumer group. Lag is best-effort: if any part
// of the calculation fails the caller simply omits the figure.
func (s *Service) kafkaLag(ctx context.Context, partitions []kafka.Partition) (int64, bool) {
	if len(partitions) == 0 || s.cfg.KafkaGroupID == "" {
		return 0, false
	}

	ids := make([]int, 0, len(partitions))
	lastOffsets := make(map[int]int64, len(partitions))

	for _, partition := range partitions {
		conn, err := kafka.DialLeader(ctx, "tcp", s.cfg.KafkaBrokers[0], s.cfg.NotificationsTopic, partition.ID)
		if err != nil {
			continue
		}
		last, err := conn.ReadLastOffset()
		_ = conn.Close()
		if err != nil {
			continue
		}
		ids = append(ids, partition.ID)
		lastOffsets[partition.ID] = last
	}

	if len(ids) == 0 {
		return 0, false
	}

	// A dedicated transport keeps the offset lookup off the shared client cache.
	client := &kafka.Client{
		Addr:      kafka.TCP(s.cfg.KafkaBrokers...),
		Timeout:   probeTimeout,
		Transport: &kafka.Transport{DialTimeout: probeTimeout},
	}
	resp, err := client.OffsetFetch(ctx, &kafka.OffsetFetchRequest{
		GroupID: s.cfg.KafkaGroupID,
		Topics:  map[string][]int{s.cfg.NotificationsTopic: ids},
	})
	if err != nil || resp.Error != nil {
		return 0, false
	}

	var lag int64
	for _, offset := range resp.Topics[s.cfg.NotificationsTopic] {
		last, ok := lastOffsets[offset.Partition]
		if !ok {
			continue
		}
		committed := offset.CommittedOffset
		if committed < 0 {
			// No committed offset yet: the whole partition backlog is pending.
			committed = 0
		}
		if last > committed {
			lag += last - committed
		}
	}
	return lag, true
}

// checkMinio verifies the object storage endpoint is reachable and the media
// bucket exists.
func (s *Service) checkMinio(ctx context.Context) ServiceHealth {
	health := ServiceHealth{Key: "minio", Status: StatusHealthy, Port: s.minioPort}
	if s.minio == nil {
		health.Status = StatusUnhealthy
		health.Error = "object storage client not configured"
		return health
	}

	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	start := time.Now()
	exists, err := s.minio.BucketExists(ctx, s.cfg.Bucket)
	health.LatencyMs = elapsedMs(start)

	if err != nil {
		health.Status = StatusUnhealthy
		health.Error = err.Error()
		return health
	}
	if !exists {
		health.Status = StatusDegraded
		health.Error = "media bucket not found"
	}
	return health
}

// rollupStatus reduces the per-service states into a single overall status,
// where any unhealthy service dominates and any degraded service downgrades an
// otherwise-healthy report.
func rollupStatus(services []ServiceHealth) string {
	status := StatusHealthy
	for _, svc := range services {
		if svc.Status == StatusUnhealthy {
			return StatusUnhealthy
		}
		if svc.Status == StatusDegraded {
			status = StatusDegraded
		}
	}
	return status
}

// elapsedMs returns the time since start in milliseconds, rounded to 2 decimals.
func elapsedMs(start time.Time) float64 {
	ms := float64(time.Since(start).Microseconds()) / 1000.0
	return math.Round(ms*100) / 100
}

// firstBrokerPort extracts the port from the first broker address, falling back
// to the supplied default when it cannot be parsed.
func firstBrokerPort(brokers []string, fallback int) int {
	if len(brokers) == 0 {
		return fallback
	}
	return portFromEndpoint(brokers[0], fallback)
}

// portFromEndpoint extracts a numeric port from a "host:port" or "scheme://host:port"
// endpoint, falling back to the supplied default when it cannot be parsed.
func portFromEndpoint(endpoint string, fallback int) int {
	endpoint = strings.TrimSpace(endpoint)
	if i := strings.Index(endpoint, "://"); i >= 0 {
		endpoint = endpoint[i+3:]
	}
	if i := strings.IndexByte(endpoint, '/'); i >= 0 {
		endpoint = endpoint[:i]
	}

	_, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		return fallback
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 {
		return fallback
	}
	return port
}

// atoiDefault parses an integer string, returning the fallback on failure.
func atoiDefault(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
