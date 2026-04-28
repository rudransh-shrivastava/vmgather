package services

import (
	"testing"
	"time"

	"github.com/VictoriaMetrics/vmgather/internal/domain"
)

func TestApplyExportDefaults_PreservesDropLabels(t *testing.T) {
	cfg := domain.ExportConfig{
		TimeRange: domain.TimeRange{
			Start: time.Now().Add(-time.Hour),
			End:   time.Now(),
		},
		Obfuscation: domain.ObfuscationConfig{
			Enabled:    false,
			DropLabels: []string{"env", "job"},
		},
	}

	ApplyExportDefaults(&cfg)

	if cfg.MetricStepSeconds == 0 {
		t.Fatalf("expected metric step to be set")
	}
	if !cfg.Safety.AutoSplit || !cfg.Safety.SplitByJob {
		t.Fatalf("expected safety defaults to be enabled, got %+v", cfg.Safety)
	}
	if cfg.Safety.MinWindowSeconds != 5 {
		t.Fatalf("expected min window to default to 5s, got %d", cfg.Safety.MinWindowSeconds)
	}
	if cfg.Safety.MaxSplitDepth != 8 {
		t.Fatalf("expected max split depth to default to 8, got %d", cfg.Safety.MaxSplitDepth)
	}
	if len(cfg.Obfuscation.DropLabels) != 2 {
		t.Fatalf("expected drop labels to be preserved, got %v", cfg.Obfuscation.DropLabels)
	}
}
