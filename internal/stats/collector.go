package stats

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"exampleserver/pkg/logger"
)

type Stats struct {
	Timestamp    time.Time
	NumGoroutine int
	MemStats     runtime.MemStats
}

type StatsService struct {
	interval time.Duration
	stats    chan Stats
	logger   logger.LoggerInterface
}

func NewStatsService(interval time.Duration, logger logger.LoggerInterface) *StatsService {
	return &StatsService{
		interval: interval,
		stats:    make(chan Stats, 100),
		logger:   logger,
	}
}

func (s *StatsService) Start(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(s.stats)
			return ctx.Err()
		case <-ticker.C:
			stats := Stats{
				Timestamp:    time.Now(),
				NumGoroutine: runtime.NumGoroutine(),
			}
			runtime.ReadMemStats(&stats.MemStats)

			// Log the stats
			s.logStats(stats)

			// Try to send stats, but don't block if channel is full
			select {
			case s.stats <- stats:
			default:
				s.logger.Error("stats channel full, dropping metrics")
			}
		}
	}
}

func (s *StatsService) logStats(stats Stats) {
	memStats := stats.MemStats
	s.logger.Info(
		"[Stats] Time: %s, Goroutines: %d, Memory: {Alloc: %s, TotalAlloc: %s, Sys: %s, NumGC: %d}",
		stats.Timestamp.Format(time.RFC3339),
		stats.NumGoroutine,
		s.formatBytes(memStats.Alloc),
		s.formatBytes(memStats.TotalAlloc),
		s.formatBytes(memStats.Sys),
		memStats.NumGC,
	)
}

func (s *StatsService) formatBytes(bytes uint64) string {
	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2fGB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2fMB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2fKB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
