package handler

import "fixture/benchmark/internal/config"

func LoadRolloutConfig() (config.RolloutConfig, string) {
	cfg := config.Load("config/rollout.production.yaml")
	return cfg, cfg.Source
}
