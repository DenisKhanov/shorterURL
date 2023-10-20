package config

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		envServAddr string
		envBaseURL  string
		args        []string
		expected    *ENVConfig
	}{
		{
			name:     "test config not environment & not flags",
			args:     []string{"cmd"},
			expected: &ENVConfig{EnvServAdr: "localhost:8080", EnvBaseURL: "http://localhost:8080"},
		},
		{
			name:     "test config not environment",
			args:     []string{"cmd", "-a", "localhost:9090", "-b", "http://flags"},
			expected: &ENVConfig{EnvServAdr: "localhost:9090", EnvBaseURL: "http://flags"},
		},
		{
			name:     "test config flag -a, not environment",
			args:     []string{"cmd", "-a", "localhost:9090"},
			expected: &ENVConfig{EnvServAdr: "localhost:9090", EnvBaseURL: "http://localhost:8080"},
		},
		{
			name:        "test config environment & flags",
			envServAddr: "localhost:9090",
			envBaseURL:  "http://enviroment",
			args:        []string{"cmd", "-a", "localhost:7070", "-b", "http://flags"},
			expected:    &ENVConfig{EnvServAdr: "localhost:9090", EnvBaseURL: "http://enviroment"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envServAddr != "" {
				os.Setenv("SERVER_ADDRESS", tt.envServAddr)
			}
			if tt.envBaseURL != "" {
				os.Setenv("BASE_URL", tt.envBaseURL)
			}
			//if tt.envLogLevel != "" {
			//	os.Setenv("LOG_LEVEl", tt.envLogLevel)
			//}

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // Сбрасываем значение флагов перед каждым тестом
			os.Args = tt.args
			cfg := NewConfig()
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
