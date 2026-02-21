package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	OAuth    OAuthConfig
	Security SecurityConfig
	Logging  LoggingConfig
	Metrics  MetricsConfig
	Tracing  TracingConfig
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

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

type JWTConfig struct {
	Secret                string
	AccessTokenExpiry     time.Duration
	RefreshTokenExpiry    time.Duration
	Issuer                string
	Audience              string
	AllowMultipleSessions bool
}

type OAuthConfig struct {
	Providers map[string]OAuthProvider
}

type OAuthProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	Enabled      bool
}

type SecurityConfig struct {
	BCryptCost         int
	MaxLoginAttempts   int
	LockoutDuration    time.Duration
	PasswordMinLength  int
	RequireSpecialChar bool
	RequireNumber      bool
	RequireUppercase   bool
}

type LoggingConfig struct {
	Level  string
	Format string
	Output string
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

	// Set config file
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Enable environment variable override
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables in config
	expandEnvVars(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate config
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func expandEnvVars(v *viper.Viper) {
	// Manually expand environment variables
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

	if cfg.JWT.Secret == "" && os.Getenv("JWT_SECRET") == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = os.Getenv("JWT_SECRET")
	}

	// Set defaults
	if cfg.Server.ShutdownTimeout == 0 {
		cfg.Server.ShutdownTimeout = 30 * time.Second
	}

	if cfg.JWT.AccessTokenExpiry == 0 {
		cfg.JWT.AccessTokenExpiry = 15 * time.Minute
	}

	if cfg.JWT.RefreshTokenExpiry == 0 {
		cfg.JWT.RefreshTokenExpiry = 7 * 24 * time.Hour
	}

	if cfg.Security.BCryptCost == 0 {
		cfg.Security.BCryptCost = 12
	}

	return nil
}
