package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Load loads configuration from file and environment
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	setDefaults()

	viper.AutomaticEnv()
	viper.SetEnvPrefix("TQ")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.shutdown_timeout", "10s")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_connections", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "5m")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)

	// Queue defaults
	viper.SetDefault("queue.max_queue_size", 10000)
	viper.SetDefault("queue.poll_interval", "1s")
	viper.SetDefault("queue.visibility_timeout", "30m")
	viper.SetDefault("queue.retention_period", "7d")

	// Worker defaults
	viper.SetDefault("worker.concurrency", 10)
	viper.SetDefault("worker.batch_size", 10)
	viper.SetDefault("worker.process_timeout", "5m")

	// Metrics defaults
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.path", "/metrics")
	viper.SetDefault("metrics.port", 9090)

	// Log defaults
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
}
