package haconfig

import (
	"os"
	"testing"
	"time"
)

type TestConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    *RedisConfig   `yaml:"redis,omitempty"`
	Features FeatureFlags   `yaml:"features"`
	Timeout  time.Duration  `yaml:"timeout"`
}

type ServerConfig struct {
	Host string `yaml:"host" required:"true"`
	Port int    `yaml:"port" required:"true"`
	TLS  bool   `yaml:"tls"`
}

type DatabaseConfig struct {
	URL         string        `yaml:"url" required:"true"`
	MaxConns    int           `yaml:"max_conns"`
	Timeout     time.Duration `yaml:"timeout"`
	Credentials *Credentials  `yaml:"credentials,omitempty"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type Credentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type FeatureFlags struct {
	EnableMetrics bool     `yaml:"enable_metrics"`
	EnableTracing bool     `yaml:"enable_tracing"`
	AllowedIPs    []string `yaml:"allowed_ips"`
}

func TestNew(t *testing.T) {
	config := New()
	if config == nil {
		t.Error("New() should return non-nil config")
	}

	config = New(WithEnvPrefix("TEST"))
	if config.envPrefix != "TEST" {
		t.Error("WithEnvPrefix option not applied")
	}

	config = New(WithYAMLFile("test.yaml"))
	if config.yamlFile != "test.yaml" {
		t.Error("WithYAMLFile option not applied")
	}
}

func TestLoadFromEnvOnly(t *testing.T) {
	// Set up test environment variables
	envVars := map[string]string{
		"SERVER_HOST":             "localhost",
		"SERVER_PORT":             "8080",
		"SERVER_TLS":              "true",
		"DATABASE_URL":            "postgres://localhost/test",
		"DATABASE_MAX_CONNS":      "10",
		"DATABASE_TIMEOUT":        "30s",
		"FEATURES_ENABLE_METRICS": "true",
		"FEATURES_ENABLE_TRACING": "false",
		"FEATURES_ALLOWED_I_PS":   "192.168.1.1,10.0.0.1,127.0.0.1",
		"TIMEOUT":                 "5m",
	}

	// Set environment variables
	for key, value := range envVars {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	config := New()
	var cfg TestConfig

	err := config.Load(&cfg)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Validate loaded values
	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Server.Port)
	}
	if !cfg.Server.TLS {
		t.Error("Expected TLS to be true")
	}
	if cfg.Database.URL != "postgres://localhost/test" {
		t.Errorf("Expected database URL 'postgres://localhost/test', got '%s'", cfg.Database.URL)
	}
	if cfg.Database.MaxConns != 10 {
		t.Errorf("Expected max_conns 10, got %d", cfg.Database.MaxConns)
	}
	if cfg.Database.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", cfg.Database.Timeout)
	}
	if !cfg.Features.EnableMetrics {
		t.Error("Expected EnableMetrics to be true")
	}
	if cfg.Features.EnableTracing {
		t.Error("Expected EnableTracing to be false")
	}
	if len(cfg.Features.AllowedIPs) != 3 {
		t.Errorf("Expected 3 allowed IPs, got %d", len(cfg.Features.AllowedIPs))
	}
	if cfg.Timeout != 5*time.Minute {
		t.Errorf("Expected timeout 5m, got %v", cfg.Timeout)
	}
}

func TestLoadFromEnvWithPrefix(t *testing.T) {
	envVars := map[string]string{
		"MYAPP_SERVER_HOST":  "example.com",
		"MYAPP_SERVER_PORT":  "9000",
		"MYAPP_DATABASE_URL": "mysql://localhost/test",
	}

	for key, value := range envVars {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	config := New(WithEnvPrefix("MYAPP"))
	var cfg TestConfig

	err := config.Load(&cfg)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Host != "example.com" {
		t.Errorf("Expected host 'example.com', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", cfg.Server.Port)
	}
	if cfg.Database.URL != "mysql://localhost/test" {
		t.Errorf("Expected database URL 'mysql://localhost/test', got '%s'", cfg.Database.URL)
	}
}

func TestLoadFromYAML(t *testing.T) {
	// Create temporary YAML file
	yamlContent := `
server:
  host: yaml-host
  port: 3000
  tls: false
database:
  url: postgres://yaml-db/test
  max_conns: 5
  timeout: 15s
  credentials:
    username: yaml-user
    password: yaml-pass
features:
  enable_metrics: false
  enable_tracing: true
  allowed_ips:
    - 192.168.1.0
    - 10.0.0.0
timeout: 2m
`

	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write YAML content: %v", err)
	}
	tmpFile.Close()

	config := New(WithYAMLFile(tmpFile.Name()))
	var cfg TestConfig

	err = config.Load(&cfg)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Validate YAML values
	if cfg.Server.Host != "yaml-host" {
		t.Errorf("Expected host 'yaml-host', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 3000 {
		t.Errorf("Expected port 3000, got %d", cfg.Server.Port)
	}
	if cfg.Server.TLS {
		t.Error("Expected TLS to be false")
	}
	if cfg.Database.Credentials == nil {
		t.Error("Expected credentials to be set")
	} else {
		if cfg.Database.Credentials.Username != "yaml-user" {
			t.Errorf("Expected username 'yaml-user', got '%s'", cfg.Database.Credentials.Username)
		}
		if cfg.Database.Credentials.Password != "yaml-pass" {
			t.Errorf("Expected password 'yaml-pass', got '%s'", cfg.Database.Credentials.Password)
		}
	}
}

func TestEnvOverridesYAML(t *testing.T) {
	// Create temporary YAML file
	yamlContent := `
server:
  host: yaml-host
  port: 3000
database:
  url: postgres://yaml-db/test
`

	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write YAML content: %v", err)
	}
	tmpFile.Close()

	// Set environment variables that should override YAML
	os.Setenv("SERVER_HOST", "env-host")
	os.Setenv("SERVER_PORT", "8080")
	defer os.Unsetenv("SERVER_HOST")
	defer os.Unsetenv("SERVER_PORT")

	config := New(WithYAMLFile(tmpFile.Name()))
	var cfg TestConfig

	err = config.Load(&cfg)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Environment variables should override YAML values
	if cfg.Server.Host != "env-host" {
		t.Errorf("Expected env override host 'env-host', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected env override port 8080, got %d", cfg.Server.Port)
	}
	// YAML value should remain where no env override
	if cfg.Database.URL != "postgres://yaml-db/test" {
		t.Errorf("Expected YAML value for database URL, got '%s'", cfg.Database.URL)
	}
}

func TestPointerFields(t *testing.T) {
	os.Setenv("REDIS_HOST", "redis-server")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_DB", "2")
	defer os.Unsetenv("REDIS_HOST")
	defer os.Unsetenv("REDIS_PORT")
	defer os.Unsetenv("REDIS_DB")

	config := New()
	var cfg TestConfig

	err := config.Load(&cfg)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Redis == nil {
		t.Error("Expected Redis config to be initialized")
	} else {
		if cfg.Redis.Host != "redis-server" {
			t.Errorf("Expected Redis host 'redis-server', got '%s'", cfg.Redis.Host)
		}
		if cfg.Redis.Port != 6379 {
			t.Errorf("Expected Redis port 6379, got %d", cfg.Redis.Port)
		}
		if cfg.Redis.DB != 2 {
			t.Errorf("Expected Redis DB 2, got %d", cfg.Redis.DB)
		}
	}
}

func TestCustomEnvMapping(t *testing.T) {
	os.Setenv("CUSTOM_HOST", "custom-host")
	os.Setenv("CUSTOM_PORT", "9090")
	defer os.Unsetenv("CUSTOM_HOST")
	defer os.Unsetenv("CUSTOM_PORT")

	mapping := map[string]string{
		"Host": "CUSTOM_HOST",
		"Port": "CUSTOM_PORT",
	}

	config := New(WithEnvMapping(mapping))
	var cfg TestConfig

	err := config.Load(&cfg)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Host != "custom-host" {
		t.Errorf("Expected custom mapped host 'custom-host', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Expected custom mapped port 9090, got %d", cfg.Server.Port)
	}
}

func TestValidation(t *testing.T) {
	config := New()

	// Test with missing required fields
	var cfg TestConfig
	cfg.Server.Host = ""                           // Required field empty
	cfg.Server.Port = 0                            // Required field empty
	cfg.Database.URL = "postgres://localhost/test" // Required field set

	err := config.Validate(&cfg)
	if err == nil {
		t.Error("Expected validation error for missing required field")
	}

	// Test with all required fields
	cfg.Server.Host = "localhost"
	cfg.Server.Port = 8080
	cfg.Database.URL = "postgres://localhost/test"
	err = config.Validate(&cfg)
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}
}

func TestInvalidValues(t *testing.T) {
	config := New()

	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "invalid bool",
			envVars: map[string]string{
				"SERVER_TLS": "invalid-bool",
			},
			wantErr: true,
		},
		{
			name: "invalid int",
			envVars: map[string]string{
				"SERVER_PORT": "not-a-number",
			},
			wantErr: true,
		},
		{
			name: "invalid duration",
			envVars: map[string]string{
				"TIMEOUT": "invalid-duration",
			},
			wantErr: true,
		},
		{
			name: "valid values",
			envVars: map[string]string{
				"SERVER_HOST": "localhost",
				"SERVER_PORT": "8080",
				"SERVER_TLS":  "true",
				"TIMEOUT":     "30s",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up env vars
			for key := range tt.envVars {
				defer os.Unsetenv(key)
			}

			// Set test env vars
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			var cfg TestConfig
			err := config.Load(&cfg)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	config := New()

	tests := []struct {
		input    string
		expected string
	}{
		{"CamelCase", "camel_case"},
		{"XMLHttpRequest", "xml_http_request"},
		{"ID", "id"},
		{"UserID", "user_id"},
		{"HTTPServer", "http_server"},
		{"simple", "simple"},
		{"", ""},
		{"TLS", "tls"},
		{"URL", "url"},
		{"AllowedIPs", "allowed_i_ps"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := config.toSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMustLoad(t *testing.T) {
	config := New()

	// Test successful load (should not panic)
	os.Setenv("SERVER_HOST", "localhost")
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	defer os.Unsetenv("SERVER_HOST")
	defer os.Unsetenv("DATABASE_URL")

	var cfg TestConfig
	config.MustLoad(&cfg)

	// Test failed load (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustLoad should panic on invalid config")
		}
	}()

	var invalidCfg string // Not a struct
	config.MustLoad(&invalidCfg)
}

func TestLoadFromFile(t *testing.T) {
	yamlContent := `
server:
  host: file-host
  port: 4000
database:
  url: postgres://file-db/test
`

	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write YAML content: %v", err)
	}
	tmpFile.Close()

	var cfg TestConfig
	err = LoadFromFile(tmpFile.Name(), &cfg)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if cfg.Server.Host != "file-host" {
		t.Errorf("Expected host 'file-host', got '%s'", cfg.Server.Host)
	}
}

func TestLoadFromEnvFunction(t *testing.T) {
	os.Setenv("TEST_SERVER_HOST", "env-only-host")
	os.Setenv("TEST_SERVER_PORT", "7000")
	defer os.Unsetenv("TEST_SERVER_HOST")
	defer os.Unsetenv("TEST_SERVER_PORT")

	var cfg TestConfig
	err := LoadFromEnv(&cfg, WithEnvPrefix("TEST"))
	if err != nil {
		t.Fatalf("LoadFromEnv failed: %v", err)
	}

	if cfg.Server.Host != "env-only-host" {
		t.Errorf("Expected host 'env-only-host', got '%s'", cfg.Server.Host)
	}
	if cfg.Server.Port != 7000 {
		t.Errorf("Expected port 7000, got %d", cfg.Server.Port)
	}
}

// Benchmark tests
func BenchmarkLoadFromEnv(b *testing.B) {
	os.Setenv("SERVER_HOST", "localhost")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	defer os.Unsetenv("SERVER_HOST")
	defer os.Unsetenv("SERVER_PORT")
	defer os.Unsetenv("DATABASE_URL")

	config := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg TestConfig
		config.Load(&cfg)
	}
}

func BenchmarkToSnakeCase(b *testing.B) {
	config := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.toSnakeCase("HTTPServerConfiguration")
	}
}
