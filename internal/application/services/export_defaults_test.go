package services

import (
	"encoding/json"
	"strings"
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
	if cfg.Safety.Mode != domain.ExportAdaptivityAutopilot {
		t.Fatalf("expected autopilot safety mode, got %q", cfg.Safety.Mode)
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
	if cfg.Safety.MaxStepSeconds != MaxAutopilotMetricStepSeconds {
		t.Fatalf("expected max step to default to 5m, got %d", cfg.Safety.MaxStepSeconds)
	}
	if cfg.MetricStepSeconds != MinBatchIntervalSeconds {
		t.Fatalf("expected autopilot to start at max precision step, got %d", cfg.MetricStepSeconds)
	}
	if len(cfg.Obfuscation.DropLabels) != 2 {
		t.Fatalf("expected drop labels to be preserved, got %v", cfg.Obfuscation.DropLabels)
	}
}

func TestApplyExportDefaults_EnablesSafetyWithNegativeSplitSettings(t *testing.T) {
	cfg := domain.ExportConfig{
		TimeRange: domain.TimeRange{
			Start: time.Now().Add(-time.Hour),
			End:   time.Now(),
		},
		Safety: domain.ExportSafetyConfig{
			MinWindowSeconds: -10,
			MaxSplitDepth:    -1,
		},
	}

	ApplyExportDefaults(&cfg)

	if !cfg.Safety.AutoSplit || !cfg.Safety.SplitByJob {
		t.Fatalf("expected invalid negative split settings to bootstrap safe defaults, got %+v", cfg.Safety)
	}
	if cfg.Safety.MinWindowSeconds != 5 || cfg.Safety.MaxSplitDepth != 8 {
		t.Fatalf("expected negative split settings to be normalized, got %+v", cfg.Safety)
	}
}

func TestExportConfigSafetyJSONIsExplicit(t *testing.T) {
	data, err := json.Marshal(domain.ExportConfig{})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if !strings.Contains(string(data), `"safety"`) {
		t.Fatalf("expected safety field to be explicit in JSON, got %s", string(data))
	}
}
