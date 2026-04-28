package services

import "github.com/VictoriaMetrics/vmgather/internal/domain"

// ApplyExportDefaults normalizes export configuration for CLI and server usage.
func ApplyExportDefaults(config *domain.ExportConfig) {
	settings := &config.Batching
	safety := &config.Safety
	if !settings.Enabled && settings.Strategy == "" && settings.CustomIntervalSecs == 0 {
		settings.Enabled = true
	}
	if settings.Strategy == "" {
		settings.Strategy = "auto"
	}
	if settings.CustomIntervalSecs < 0 {
		settings.CustomIntervalSecs = 0
	}
	minSeconds := MinBatchIntervalSeconds
	maxSeconds := MaxBatchIntervalSeconds
	if settings.CustomIntervalSecs > 0 && settings.CustomIntervalSecs < minSeconds {
		settings.CustomIntervalSecs = minSeconds
	}
	if settings.CustomIntervalSecs > maxSeconds {
		settings.CustomIntervalSecs = maxSeconds
	}
	if config.MetricStepSeconds <= 0 {
		config.MetricStepSeconds = RecommendedMetricStepSeconds(config.TimeRange)
	}
	if !safety.AutoSplit && !safety.SplitByJob && !safety.SplitByMetricName && safety.MinWindowSeconds == 0 && safety.MaxSplitDepth == 0 {
		safety.AutoSplit = true
		safety.SplitByJob = true
	}
	if safety.MinWindowSeconds <= 0 {
		safety.MinWindowSeconds = 5
	}
	if safety.MaxSplitDepth <= 0 {
		safety.MaxSplitDepth = 8
	}
	if !config.Obfuscation.Enabled {
		config.Obfuscation = domain.ObfuscationConfig{DropLabels: config.Obfuscation.DropLabels}
	}
}
