package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/school-api/school-api-v1/internal/config"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/repository"
)

// LogRetentionService sweeps relay call logs older than the configured
// retention window (customer requirement #7: logs kept 30 days).
type LogRetentionService struct {
	repo     repository.CallLogRepository
	days     int
	interval time.Duration
	now      func() time.Time // injectable for tests
}

// NewLogRetentionService builds the service from retention config.
func NewLogRetentionService(repo repository.CallLogRepository, cfg config.RetentionConfig) *LogRetentionService {
	return &LogRetentionService{
		repo:     repo,
		days:     cfg.CallLogDays,
		interval: cfg.SweepInterval,
		now:      time.Now,
	}
}

// Enabled reports whether sweeping is configured (call_log_days > 0).
func (s *LogRetentionService) Enabled() bool {
	return s.days > 0
}

// SweepOnce deletes logs older than the retention window and returns how many
// rows were removed.
func (s *LogRetentionService) SweepOnce(ctx context.Context) (int, error) {
	if !s.Enabled() {
		return 0, nil
	}
	cutoff := s.now().AddDate(0, 0, -s.days)
	n, err := s.repo.DeleteOlderThan(ctx, cutoff)
	if err != nil {
		return 0, domain.WrapInternal(err)
	}
	return n, nil
}

// Start sweeps immediately once and then on every interval until ctx is
// cancelled. Intended to run in a goroutine; it never returns an error, so
// failures are logged rather than propagated.
func (s *LogRetentionService) Start(ctx context.Context, logger *slog.Logger) {
	if !s.Enabled() {
		logger.Info("call-log retention disabled", "days", s.days)
		return
	}
	s.run(ctx, logger)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.run(ctx, logger)
		}
	}
}

func (s *LogRetentionService) run(ctx context.Context, logger *slog.Logger) {
	n, err := s.SweepOnce(ctx)
	if err != nil {
		logger.Error("call-log sweep failed", "error", err)
		return
	}
	if n > 0 {
		logger.Info("call-log sweep removed rows", "deleted", n, "days", s.days)
	}
}
