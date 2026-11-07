// Package system exposes operational health diagnostics for the admin back office.
// It probes each backing dependency (PostgreSQL, Redis, Kafka, MinIO) with a real
// round-trip and reports live connection status, latency, and the resource usage
// each service exposes over its own protocol, so operators see the actual
// heartbeat and footprint of the platform rather than static labels.
package system

import (
	"context"
	"errors"
	"fmt"
	"math"
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

	// maxListObjects caps the MinIO usage scan so a large media bucket cannot
	// stall the health report. The object count is suffixed with "+" when the
	// cap is hit.
	maxListObjects = 1000
)

// Config carries the dependency targets the health service probes.
type Config struct {
	// KafkaBrokers is the list of seed broker host:port addresses.
	KafkaBrokers []string
	// KafkaGroupID is the notification worker consumer group used for lag.
	KafkaGroupID string
	// NotificationsTopic is the topic inspected for connectivity and lag.
	NotificationsTopic string
	// Bucket is the object storage bucket whose usage is reported.
	Bucket string
}

// ResourceMetric is a single labeled resource reading for a service.
type ResourceMetric struct {
	// Key is the stable identifier the client maps to a localized label
	// (e.g. "memory", "connections", "lag").
	Key string `json:"key"`
	// Value is the formatted, already-human-readable reading (e.g. "1.2 MB").
	Value string `json:"value"`
}

// ServiceHealth describes the observed state of a single backing service.
type ServiceHealth struct {
	// Key is the stable identifier ("postgres", "redis", "kafka", "minio") that
	// the client uses to select an icon and localized labels.
	Key string `json:"key"`
	// Status is one of StatusHealthy, StatusDegraded, or StatusUnhealthy.
	Status string `json:"status"`
	// LatencyMs is the measured round-trip time in milliseconds.
	LatencyMs float64 `json:"latencyMs"`
	// Resources holds the service's own resource readings, when available.
	Resources []ResourceMetric `json:"resources,omitempty"`
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
	db    *database.DB
	redis *redis.Client
	minio *minio.Client
	cfg   Config
}

// NewService constructs a health service over the supplied dependency clients.
func NewService(db *database.DB, redisClient *redis.Client, minioClient *minio.Client, cfg Config) *Service {
	if db == nil {
		panic("system: database is required")
	}
	if redisClient == nil {
		panic("system: redis client is required")
	}
	if minioClient == nil {
		panic("system: minio client is required")
	}
	return &Service{
		db:    db,
		redis: redisClient,
		minio: minioClient,
		cfg:   cfg,
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

// checkPostgres verifies the database answers a ping and reports pool usage and
// on-disk size.
func (s *Service) checkPostgres(ctx context.Context) ServiceHealth {
	health := ServiceHealth{Key: "postgres", Status: StatusHealthy}

	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	start := time.Now()
	err := s.db.Ping(ctx)
	health.LatencyMs = elapsedMs(start)

	if err != nil {
		// The pool is shared with live traffic, so a busy pool can starve the
		// probe of a connection even though PostgreSQL itself is healthy. Report
		// that distinctly (with the pool gauge) instead of a false outage.
		if errors.Is(err, context.DeadlineExceeded) && s.poolSaturated() {
			health.Status = StatusDegraded
			health.Error = "connection pool saturated"
		} else {
			health.Status = StatusUnhealthy
			health.Error = err.Error()
		}
		health.Resources = s.postgresPoolResource()
		return health
	}

	health.Resources = s.postgresResources(ctx)
	return health
}

// checkRedis verifies the cache/session store answers a ping and reports memory,
// key count, and connected clients.
func (s *Service) checkRedis(ctx context.Context) ServiceHealth {
	health := ServiceHealth{Key: "redis", Status: StatusHealthy}

	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	start := time.Now()
	err := s.redis.Ping(ctx).Err()
	health.LatencyMs = elapsedMs(start)

	if err != nil {
		health.Status = StatusUnhealthy
		health.Error = err.Error()
		return health
	}

	health.Resources = s.redisResources(ctx)
	return health
}

// checkKafka verifies broker/topic reachability and reports the cluster's broker
// count, topic partitions, and consumer lag for the notifications topic.
func (s *Service) checkKafka(ctx context.Context) ServiceHealth {
	health := ServiceHealth{Key: "kafka", Status: StatusHealthy}
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
	brokers, _ := conn.Brokers()
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

	resources := []ResourceMetric{
		metric("brokers", strconv.Itoa(len(brokers))),
		metric("partitions", strconv.Itoa(len(partitions))),
	}
	if lag, total, ok := s.kafkaOffsets(ctx, partitions); ok {
		resources = append(resources,
			metric("messages", compactCount(total)),
			metric("lag", strconv.FormatInt(lag, 10)))
	}
	health.Resources = resources
	return health
}

// kafkaOffsets returns the topic's total produced message count (the sum of each
// partition's high watermark) and the configured consumer group's lag. Both are
// best-effort: if the calculation fails the caller simply omits the figures.
func (s *Service) kafkaOffsets(ctx context.Context, partitions []kafka.Partition) (lag int64, total int64, ok bool) {
	if len(partitions) == 0 || s.cfg.KafkaGroupID == "" {
		return 0, 0, false
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
		return 0, 0, false
	}

	for _, id := range ids {
		total += lastOffsets[id]
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
		return 0, total, true
	}

	for _, offset := range resp.Topics[s.cfg.NotificationsTopic] {
		last, exists := lastOffsets[offset.Partition]
		if !exists {
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
	return lag, total, true
}

// checkMinio verifies the object storage endpoint is reachable and reports the
// media bucket's object count and total size.
func (s *Service) checkMinio(ctx context.Context) ServiceHealth {
	health := ServiceHealth{Key: "minio", Status: StatusHealthy}

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
		return health
	}

	health.Resources = s.minioResources(ctx)
	return health
}

// postgresPoolResource reports the pool's open connections against its ceiling.
// It reads only in-memory pool statistics, so it never blocks on the database and
// is safe to return even when a live round-trip is not possible.
func (s *Service) postgresPoolResource() []ResourceMetric {
	if s.db == nil || s.db.Pool == nil {
		return nil
	}
	stat := s.db.Pool.Stat()
	// Report the pool's open connections (in-use + idle) against its ceiling,
	// which stays meaningful at rest, rather than only the checked-out count.
	return []ResourceMetric{metric("connections",
		fmt.Sprintf("%d / %d", stat.TotalConns(), stat.MaxConns()))}
}

// poolSaturated reports whether every pooled connection is currently checked out.
func (s *Service) poolSaturated() bool {
	if s.db == nil || s.db.Pool == nil {
		return false
	}
	stat := s.db.Pool.Stat()
	return stat.MaxConns() > 0 && stat.AcquiredConns() >= stat.MaxConns()
}

// postgresResources reports the connection pool utilisation and database size.
func (s *Service) postgresResources(ctx context.Context) []ResourceMetric {
	resources := s.postgresPoolResource()
	if resources == nil {
		return nil
	}

	var size string
	if err := s.db.QueryRowContext(ctx, "SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&size); err == nil && size != "" {
		resources = append(resources, metric("storage", size))
	}
	return resources
}

// redisResources reports cache memory, key count, and connected clients.
func (s *Service) redisResources(ctx context.Context) []ResourceMetric {
	resources := make([]ResourceMetric, 0, 4)
	if info, err := s.redis.Info(ctx, "memory").Result(); err == nil {
		if mem := parseInfoField(info, "used_memory_human"); mem != "" {
			resources = append(resources, metric("memory", mem))
		}
	}
	if size, err := s.redis.DBSize(ctx).Result(); err == nil {
		resources = append(resources, metric("keys", strconv.FormatInt(size, 10)))
	}
	if info, err := s.redis.Info(ctx, "clients").Result(); err == nil {
		if clients := parseInfoField(info, "connected_clients"); clients != "" {
			resources = append(resources, metric("clients", clients))
		}
	}
	// Throughput makes Redis load visible: it climbs with request volume rather
	// than reflecting only the steady-state key/memory gauges.
	if info, err := s.redis.Info(ctx, "stats").Result(); err == nil {
		if ops := parseInfoField(info, "instantaneous_ops_per_sec"); ops != "" {
			resources = append(resources, metric("ops", ops))
		}
	}
	return resources
}

// minioResources reports the media bucket's object count and total byte size.
func (s *Service) minioResources(ctx context.Context) []ResourceMetric {
	count, bytes, truncated := s.bucketUsage(ctx, s.cfg.Bucket)
	objects := strconv.FormatInt(count, 10)
	if truncated {
		objects += "+"
	}
	return []ResourceMetric{
		metric("objects", objects),
		metric("storage", humanBytes(bytes)),
	}
}

// bucketUsage counts objects and sums their sizes in a bucket, stopping at
// maxListObjects to keep the scan bounded.
func (s *Service) bucketUsage(ctx context.Context, bucket string) (count int64, bytes int64, truncated bool) {
	for object := range s.minio.ListObjects(ctx, bucket, minio.ListObjectsOptions{Recursive: true}) {
		if object.Err != nil {
			break
		}
		count++
		bytes += object.Size
		if count >= maxListObjects {
			return count, bytes, true
		}
	}
	return count, bytes, false
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

// metric is a small constructor for a labeled resource reading.
func metric(key, value string) ResourceMetric {
	return ResourceMetric{Key: key, Value: value}
}

// parseInfoField extracts a "field:value" line from a Redis INFO payload.
func parseInfoField(info, field string) string {
	for _, line := range strings.Split(info, "\n") {
		name, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if ok && name == field {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// elapsedMs returns the time since start in milliseconds, rounded to 2 decimals.
func elapsedMs(start time.Time) float64 {
	ms := float64(time.Since(start).Microseconds()) / 1000.0
	return math.Round(ms*100) / 100
}

// compactCount abbreviates a large count into a short human string (e.g. "12.3k").
func compactCount(n int64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return strconv.FormatInt(n, 10)
	}
}

// humanBytes formats a byte count into a compact IEC string (e.g. "1.2 MB").
func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for value := n / unit; value >= unit; value /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
