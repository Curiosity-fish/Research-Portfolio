package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/internal/config"
	"github.com/school-api/school-api-v1/internal/repository"
)

// fakeCallLogRepo captures the cutoff passed to DeleteOlderThan and returns a
// canned count so the sweep logic can be verified without a database.
type fakeCallLogRepo struct {
	gotCutoff time.Time
	calls     int
	deleted   int
	err       error
}

func (f *fakeCallLogRepo) Create(context.Context, repository.CreateCallLogInput) (*ent.CallLog, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeCallLogRepo) DeleteOlderThan(_ context.Context, cutoff time.Time) (int, error) {
	f.gotCutoff = cutoff
	f.calls++
	return f.deleted, f.err
}

func TestLogRetentionService_SweepOnceComputesCutoff(t *testing.T) {
	repo := &fakeCallLogRepo{deleted: 7}
	svc := NewLogRetentionService(repo, config.RetentionConfig{CallLogDays: 30, SweepInterval: time.Hour})
	fixed := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return fixed }

	n, err := svc.SweepOnce(context.Background())
	if err != nil {
		t.Fatalf("sweep once: %v", err)
	}
	if n != 7 {
		t.Fatalf("expected 7 deleted, got %d", n)
	}
	want := fixed.AddDate(0, 0, -30)
	if !repo.gotCutoff.Equal(want) {
		t.Fatalf("expected cutoff %v, got %v", want, repo.gotCutoff)
	}
	if repo.calls != 1 {
		t.Fatalf("expected repo called once, got %d", repo.calls)
	}
}

func TestLogRetentionService_SweepDisabled(t *testing.T) {
	repo := &fakeCallLogRepo{}
	svc := NewLogRetentionService(repo, config.RetentionConfig{CallLogDays: 0})
	if svc.Enabled() {
		t.Fatal("expected disabled when days <= 0")
	}
	n, err := svc.SweepOnce(context.Background())
	if err != nil {
		t.Fatalf("sweep once: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 deleted when disabled, got %d", n)
	}
	if repo.calls != 0 {
		t.Fatalf("expected repo not called, got %d calls", repo.calls)
	}
}

func TestLogRetentionService_SweepErrorWrapped(t *testing.T) {
	repo := &fakeCallLogRepo{err: errors.New("db down")}
	svc := NewLogRetentionService(repo, config.RetentionConfig{CallLogDays: 30})

	_, err := svc.SweepOnce(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if repo.calls != 1 {
		t.Fatalf("expected repo called once, got %d", repo.calls)
	}
}
