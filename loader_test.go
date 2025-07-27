package kongtoml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/kong"
)

type SimpleCLI struct {
	Host    string `help:"Server host"`
	Port    int    `help:"Server port"`
	Debug   bool   `help:"Enable debug mode"`
	Timeout int    `help:"Request timeout"`
}

type NestedCLI struct {
	Name    string `help:"Application name"`
	Version string `help:"Application version"`

	Server struct {
		Host    string `help:"Server host"`
		Port    int    `help:"Server port"`
		Timeout int    `help:"Server timeout"`

		TLS struct {
			Enabled bool   `help:"Enable TLS"`
			Cert    string `help:"Certificate file"`
			Key     string `help:"Key file"`
		} `embed:"" prefix:"tls-"`
	} `embed:"" prefix:"server-"`

	Database struct {
		Driver string `help:"Database driver"`
		Host   string `help:"Database host"`
		Port   int    `help:"Database port"`

		Connection struct {
			MaxIdle     int    `help:"Max idle connections" name:"max-idle"`
			MaxOpen     int    `help:"Max open connections" name:"max-open"`
			MaxLifetime string `help:"Max connection lifetime" name:"max-lifetime"`
		} `embed:"" prefix:"connection-"`

		Migrations struct {
			Enabled bool   `help:"Enable migrations"`
			Path    string `help:"Migration path"`
		} `embed:"" prefix:"migrations-"`
	} `embed:"" prefix:"database-"`

	Logging struct {
		Level  string `help:"Log level"`
		Format string `help:"Log format"`

		Outputs struct {
			Console bool   `help:"Enable console output"`
			File    string `help:"Log file"`
			Syslog  bool   `help:"Enable syslog"`
		} `embed:"" prefix:"outputs-"`
	} `embed:"" prefix:"logging-"`

	Features struct {
		Auth    bool `help:"Enable authentication"`
		Metrics bool `help:"Enable metrics"`
		Tracing bool `help:"Enable tracing"`

		Cache struct {
			Enabled bool   `help:"Enable cache"`
			TTL     string `help:"Cache TTL"`
			Size    int    `help:"Cache size"`
		} `embed:"" prefix:"cache-"`
	} `embed:"" prefix:"features-"`
}

type HyphenatedCLI struct {
	MaxConnections int    `help:"Maximum connections" name:"max-connections"`
	ReadTimeout    string `help:"Read timeout" name:"read-timeout"`
	WriteTimeout   string `help:"Write timeout" name:"write-timeout"`
	IdleTimeout    string `help:"Idle timeout" name:"idle-timeout"`
	KeepAlive      bool   `help:"Keep alive" name:"keep-alive"`

	HTTPServer struct {
		BindAddress   string `help:"Bind address" name:"bind-address"`
		ListenPort    int    `help:"Listen port" name:"listen-port"`
		MaxHeaderSize int    `help:"Max header size" name:"max-header-size"`
		RequestTimeout string `help:"Request timeout" name:"request-timeout"`

		RateLimiting struct {
			Enabled           bool `help:"Enable rate limiting"`
			RequestsPerMinute int  `help:"Requests per minute" name:"requests-per-minute"`
			BurstSize         int  `help:"Burst size" name:"burst-size"`
		} `embed:"" prefix:"rate-limiting-"`
	} `embed:"" prefix:"http-server-"`

	DatabasePool struct {
		MaxOpenConnections     int    `help:"Max open connections" name:"max-open-connections"`
		MaxIdleConnections     int    `help:"Max idle connections" name:"max-idle-connections"`
		ConnectionMaxLifetime  string `help:"Connection max lifetime" name:"connection-max-lifetime"`
		ConnectionMaxIdleTime  string `help:"Connection max idle time" name:"connection-max-idle-time"`

		HealthCheck struct {
			Enabled       bool   `help:"Enable health check"`
			CheckInterval string `help:"Check interval" name:"check-interval"`
			Timeout       string `help:"Timeout"`
		} `embed:"" prefix:"health-check-"`
	} `embed:"" prefix:"database-pool-"`

	FeatureFlags struct {
		EnableAuth      bool `help:"Enable auth" name:"enable-auth"`
		EnableMetrics   bool `help:"Enable metrics" name:"enable-metrics"`
		EnableDebugMode bool `help:"Enable debug mode" name:"enable-debug-mode"`
		AutoMigrate     bool `help:"Auto migrate" name:"auto-migrate"`
	} `embed:"" prefix:"feature-flags-"`

	SSLConfig struct {
		CertFile     string `help:"Certificate file" name:"cert-file"`
		KeyFile      string `help:"Key file" name:"key-file"`
		CAFile       string `help:"CA file" name:"ca-file"`
		VerifyClient bool   `help:"Verify client" name:"verify-client"`
	} `embed:"" prefix:"ssl-config-"`
}

func TestLoader(t *testing.T) {
	tests := []struct {
		Name     string
		CLI      interface{}
		TOMLFile string
		Expected interface{}
		WantErr  bool
	}{
		{
			Name:     "SimpleConfiguration",
			CLI:      &SimpleCLI{},
			TOMLFile: "simple.toml",
			Expected: &SimpleCLI{
				Host:    "localhost",
				Port:    8080,
				Debug:   true,
				Timeout: 30,
			},
		},
		{
			Name:     "NestedConfiguration",
			CLI:      &NestedCLI{},
			TOMLFile: "nested.toml",
			Expected: &NestedCLI{
				Name:    "myapp",
				Version: "1.0.0",
				Server: struct {
					Host    string `help:"Server host"`
					Port    int    `help:"Server port"`
					Timeout int    `help:"Server timeout"`
					TLS     struct {
						Enabled bool   `help:"Enable TLS"`
						Cert    string `help:"Certificate file"`
						Key     string `help:"Key file"`
					} `embed:"" prefix:"tls-"`
				}{
					Host:    "0.0.0.0",
					Port:    8080,
					Timeout: 30,
					TLS: struct {
						Enabled bool   `help:"Enable TLS"`
						Cert    string `help:"Certificate file"`
						Key     string `help:"Key file"`
					}{
						Enabled: true,
						Cert:    "/path/to/cert.pem",
						Key:     "/path/to/key.pem",
					},
				},
				Database: struct {
					Driver string `help:"Database driver"`
					Host   string `help:"Database host"`
					Port   int    `help:"Database port"`
					Connection struct {
						MaxIdle     int    `help:"Max idle connections" name:"max-idle"`
						MaxOpen     int    `help:"Max open connections" name:"max-open"`
						MaxLifetime string `help:"Max connection lifetime" name:"max-lifetime"`
					} `embed:"" prefix:"connection-"`
					Migrations struct {
						Enabled bool   `help:"Enable migrations"`
						Path    string `help:"Migration path"`
					} `embed:"" prefix:"migrations-"`
				}{
					Driver: "postgres",
					Host:   "localhost",
					Port:   5432,
					Connection: struct {
						MaxIdle     int    `help:"Max idle connections" name:"max-idle"`
						MaxOpen     int    `help:"Max open connections" name:"max-open"`
						MaxLifetime string `help:"Max connection lifetime" name:"max-lifetime"`
					}{
						MaxIdle:     10,
						MaxOpen:     100,
						MaxLifetime: "1h",
					},
					Migrations: struct {
						Enabled bool   `help:"Enable migrations"`
						Path    string `help:"Migration path"`
					}{
						Enabled: true,
						Path:    "./migrations",
					},
				},
				Logging: struct {
					Level  string `help:"Log level"`
					Format string `help:"Log format"`
					Outputs struct {
						Console bool   `help:"Enable console output"`
						File    string `help:"Log file"`
						Syslog  bool   `help:"Enable syslog"`
					} `embed:"" prefix:"outputs-"`
				}{
					Level:  "debug",
					Format: "json",
					Outputs: struct {
						Console bool   `help:"Enable console output"`
						File    string `help:"Log file"`
						Syslog  bool   `help:"Enable syslog"`
					}{
						Console: true,
						File:    "/var/log/app.log",
						Syslog:  false,
					},
				},
				Features: struct {
					Auth    bool `help:"Enable authentication"`
					Metrics bool `help:"Enable metrics"`
					Tracing bool `help:"Enable tracing"`
					Cache   struct {
						Enabled bool   `help:"Enable cache"`
						TTL     string `help:"Cache TTL"`
						Size    int    `help:"Cache size"`
					} `embed:"" prefix:"cache-"`
				}{
					Auth:    true,
					Metrics: true,
					Tracing: false,
					Cache: struct {
						Enabled bool   `help:"Enable cache"`
						TTL     string `help:"Cache TTL"`
						Size    int    `help:"Cache size"`
					}{
						Enabled: true,
						TTL:     "5m",
						Size:    1000,
					},
				},
			},
		},
		{
			Name:     "HyphenatedKeys",
			CLI:      &HyphenatedCLI{},
			TOMLFile: "hyphenated.toml",
			Expected: &HyphenatedCLI{
				MaxConnections: 100,
				ReadTimeout:    "30s",
				WriteTimeout:   "30s",
				IdleTimeout:    "2m",
				KeepAlive:      true,
				HTTPServer: struct {
					BindAddress    string `help:"Bind address" name:"bind-address"`
					ListenPort     int    `help:"Listen port" name:"listen-port"`
					MaxHeaderSize  int    `help:"Max header size" name:"max-header-size"`
					RequestTimeout string `help:"Request timeout" name:"request-timeout"`
					RateLimiting   struct {
						Enabled           bool `help:"Enable rate limiting"`
						RequestsPerMinute int  `help:"Requests per minute" name:"requests-per-minute"`
						BurstSize         int  `help:"Burst size" name:"burst-size"`
					} `embed:"" prefix:"rate-limiting-"`
				}{
					BindAddress:    "127.0.0.1",
					ListenPort:     8080,
					MaxHeaderSize:  8192,
					RequestTimeout: "10s",
					RateLimiting: struct {
						Enabled           bool `help:"Enable rate limiting"`
						RequestsPerMinute int  `help:"Requests per minute" name:"requests-per-minute"`
						BurstSize         int  `help:"Burst size" name:"burst-size"`
					}{
						Enabled:           true,
						RequestsPerMinute: 1000,
						BurstSize:         50,
					},
				},
				DatabasePool: struct {
					MaxOpenConnections    int    `help:"Max open connections" name:"max-open-connections"`
					MaxIdleConnections    int    `help:"Max idle connections" name:"max-idle-connections"`
					ConnectionMaxLifetime string `help:"Connection max lifetime" name:"connection-max-lifetime"`
					ConnectionMaxIdleTime string `help:"Connection max idle time" name:"connection-max-idle-time"`
					HealthCheck           struct {
						Enabled       bool   `help:"Enable health check"`
						CheckInterval string `help:"Check interval" name:"check-interval"`
						Timeout       string `help:"Timeout"`
					} `embed:"" prefix:"health-check-"`
				}{
					MaxOpenConnections:    25,
					MaxIdleConnections:    5,
					ConnectionMaxLifetime: "1h",
					ConnectionMaxIdleTime: "10m",
					HealthCheck: struct {
						Enabled       bool   `help:"Enable health check"`
						CheckInterval string `help:"Check interval" name:"check-interval"`
						Timeout       string `help:"Timeout"`
					}{
						Enabled:       true,
						CheckInterval: "30s",
						Timeout:       "5s",
					},
				},
				FeatureFlags: struct {
					EnableAuth      bool `help:"Enable auth" name:"enable-auth"`
					EnableMetrics   bool `help:"Enable metrics" name:"enable-metrics"`
					EnableDebugMode bool `help:"Enable debug mode" name:"enable-debug-mode"`
					AutoMigrate     bool `help:"Auto migrate" name:"auto-migrate"`
				}{
					EnableAuth:      true,
					EnableMetrics:   false,
					EnableDebugMode: true,
					AutoMigrate:     false,
				},
				SSLConfig: struct {
					CertFile     string `help:"Certificate file" name:"cert-file"`
					KeyFile      string `help:"Key file" name:"key-file"`
					CAFile       string `help:"CA file" name:"ca-file"`
					VerifyClient bool   `help:"Verify client" name:"verify-client"`
				}{
					CertFile:     "/etc/ssl/certs/server.crt",
					KeyFile:      "/etc/ssl/private/server.key",
					CAFile:       "/etc/ssl/certs/ca.crt",
					VerifyClient: false,
				},
			},
		},
		{
			Name:     "EmptyConfiguration",
			CLI:      &SimpleCLI{},
			TOMLFile: "empty.toml",
			Expected: &SimpleCLI{},
		},
		{
			Name:     "InvalidTOML",
			CLI:      &SimpleCLI{},
			TOMLFile: "invalid.toml",
			WantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Open the TOML file
			file, err := os.Open(filepath.Join("testdata", tt.TOMLFile))
			if err != nil {
				t.Fatalf("failed to open test file %s: %v", tt.TOMLFile, err)
			}
			defer file.Close()

			// Create the loader
			loader, err := Loader(file)
			if tt.WantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Parse with kong using the loader
			parser, err := kong.New(tt.CLI, kong.Resolvers(loader))
			assert.NoError(t, err)

			// Parse empty args to trigger configuration loading
			_, err = parser.Parse([]string{})
			if tt.WantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Compare the results
			assert.Equal(t, tt.Expected, tt.CLI)
		})
	}
}

func TestResolver_Resolve(t *testing.T) {
	// Simple integration test to verify resolver can be created
	loader, err := Loader(strings.NewReader(`host = "localhost"`))
	assert.NoError(t, err)

	resolver := loader.(*Resolver)
	assert.True(t, resolver != nil)
}

func TestResolver_Validate(t *testing.T) {
	loader, err := Loader(strings.NewReader(`host = "localhost"`))
	assert.NoError(t, err)

	resolver := loader.(*Resolver)
	app := &kong.Application{}

	// Currently Validate is a no-op, so it should always return nil
	err = resolver.Validate(app)
	assert.NoError(t, err)
}

func TestLoaderWithInvalidTOML(t *testing.T) {
	invalidTOML := `[section\nmissing-bracket = true`

	_, err := Loader(strings.NewReader(invalidTOML))
	assert.Error(t, err)
}

func TestLoaderWithNamedReader(t *testing.T) {
	file, err := os.Open(filepath.Join("testdata", "simple.toml"))
	assert.NoError(t, err)
	defer file.Close()

	loader, err := Loader(file)
	assert.NoError(t, err)

	resolver := loader.(*Resolver)
	assert.True(t, strings.Contains(resolver.filename, "simple.toml"))
}

func TestLoaderWithUnnamedReader(t *testing.T) {
	loader, err := Loader(strings.NewReader(`host = "localhost"`))
	assert.NoError(t, err)

	resolver := loader.(*Resolver)
	assert.Equal(t, "", resolver.filename)
}
