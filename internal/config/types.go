package config

import "time"

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Queue    QueueConfig    `mapstructure:"queue"`
	Worker   WorkerConfig   `mapstructure:"worker"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Tracing  TracingConfig  `mapstructure:"tracing"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	TLSEnabled      bool          `mapstructure:"tls_enabled"`
	TLSCertFile     string        `mapstructure:"tls_cert_file"`
	TLSKeyFile      string        `mapstructure:"tls_key_file"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxConnections  int           `mapstructure:"max_connections"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// QueueConfig holds queue-specific configuration
type QueueConfig struct {
	MaxQueueSize         int           `mapstructure:"max_queue_size"`
	PollInterval         time.Duration `mapstructure:"poll_interval"`
	VisibilityTimeout    time.Duration `mapstructure:"visibility_timeout"`
	RetentionPeriod      time.Duration `mapstructure:"retention_period"`
	DeadLetterMaxRetries int           `mapstructure:"dead_letter_max_retries"`
}

// WorkerConfig holds worker-specific configuration
type WorkerConfig struct {
	Concurrency       int           `mapstructure:"concurrency"`
	BatchSize         int           `mapstructure:"batch_size"`
	ProcessTimeout    time.Duration `mapstructure:"process_timeout"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
	Port    int    `mapstructure:"port"`
}

// TracingConfig holds tracing configuration
type TracingConfig struct {
	Enabled      bool    `mapstructure:"enabled"`
	ServiceName  string  `mapstructure:"service_name"`
	CollectorURL string  `mapstructure:"collector_url"`
	SampleRate   float64 `mapstructure:"sample_rate"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
}
