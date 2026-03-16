package config

import "strings"

type RolloutConfig struct {
	Environment string
	Source      string
}

func Load(path string) RolloutConfig {
	env := "development"
	if strings.Contains(path, "production") {
		env = "production"
	}
	return RolloutConfig{
		Environment: env,
		Source:      path,
	}
}
