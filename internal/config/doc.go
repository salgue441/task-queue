// Package config provides a centralized configuration management system for
// the application. It supports configuration loading from multiple sources
// including YAML files and environment variables, with sensible defaults and
// strong typing.
//
// The package uses viper for configuration management and provides structured
// configuration types for all application components.
//
// Configuration Hierarchy:
//   - Default values are set for all configurable parameters
//   - Values can be overridden by configuration file
//   - Environment variables (with TQ_ prefix) take highest precedence
//
// Configuration Structure:
//   - Server: HTTP server settings including timeouts and TLS
//   - Database: PostgreSQL connection parameters and pool settings
//   - Redis: Redis connection and pooling configuration
//   - Queue: Job queue parameters including visibility and retention
//   - Worker: Concurrency and processing settings
//   - Metrics: Prometheus metrics endpoint configuration
//   - Tracing: Distributed tracing setup (e.g., Jaeger)
//   - Log: Logging format and level configuration
//
// Usage:
//
//	cfg, err := config.Load("config.yaml")
//	if err != nil {
//	    log.Fatal("Failed to load config:", err)
//	}
//	fmt.Println("Server port:", cfg.Server.Port)
//
// Environment Variables:
// All configuration can be overridden with environment variables using the pattern:
// TQ_SERVER_PORT=8081 would override the server port setting.
package config
