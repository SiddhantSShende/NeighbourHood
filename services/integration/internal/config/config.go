package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Providers map[string]ProviderConfig
	Retry     RetryConfig
	Circuit   CircuitConfig
	Logging   LoggingConfig
	Metrics   MetricsConfig
	Tracing   TracingConfig
}

type ServerConfig struct {
	Port            int
	GRPC            GRPCConfig
	ShutdownTimeout time.Duration
}

type GRPCConfig struct {
	MaxRecvMsgSize int
	MaxSendMsgSize int
	Keepalive      KeepaliveConfig
}

type KeepaliveConfig struct {
	Time    time.Duration
	Timeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	Enabled      bool
	Timeout      time.Duration
	RateLimit    int
	RatePeriod   time.Duration
}

type RetryConfig struct {
	MaxAttempts       int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
}

type CircuitConfig struct {
	MaxRequests uint32
	Interval    time.Duration
	Timeout     time.Duration
}

type LoggingConfig struct {
	Level  string
	Format string
}

type MetricsConfig struct {
	Enabled bool
	Port    int
	Path    string
}

type TracingConfig struct {
	Enabled     bool
	Endpoint    string
	ServiceName string
	SampleRate  float64
}

func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	expandEnvVars(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func expandEnvVars(v *viper.Viper) {
	cfg := v.AllSettings()
	expandMap(cfg)
	for k, val := range cfg {
		v.Set(k, val)
	}
}

func expandMap(m map[string]interface{}) {
	for k, v := range m {
		switch val := v.(type) {
		case string:
			m[k] = os.ExpandEnv(val)
		case map[string]interface{}:
			expandMap(val)
		case []interface{}:
			for i, item := range val {
				if str, ok := item.(string); ok {
					val[i] = os.ExpandEnv(str)
				} else if mapItem, ok := item.(map[string]interface{}); ok {
					expandMap(mapItem)
				}
			}
		}
	}
}

func validate(cfg *Config) error {
	if cfg.Server.Port == 0 {
		return fmt.Errorf("server port is required")
	}
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if len(cfg.Providers) == 0 {
		return fmt.Errorf("at least one provider must be configured")
	}

	// Set defaults
	if cfg.Server.ShutdownTimeout == 0 {
		cfg.Server.ShutdownTimeout = 30 * time.Second
	}
	if cfg.Retry.MaxAttempts == 0 {
		cfg.Retry.MaxAttempts = 3
	}

	return nil
}
