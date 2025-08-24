package config

import (
	"os"
	"testing"
)

func TestNewServiceConfig(t *testing.T) {
	p := 8080
	config, err := NewServiceConfig(p)

	if err != nil {
		t.Errorf("NewServiceConfig failed: %v", err)
	}

	if config.Port != p {
		t.Errorf("NewServiceConfig failed: port mismatch")
	}
}

func TestNewServiceConfigFromEnv(t *testing.T) {
	oldPort := os.Getenv("STATISTICS_SERVICE_PORT")

	defer os.Setenv("STATISTICS_SERVICE_PORT", oldPort)

	os.Setenv("STATISTICS_SERVICE_PORT", "invalid port")

	if _, err := NewServiceConfigFromEnv(); err == nil {
		t.Errorf("NewServiceConfigFromEnv failed: %v", err)
	}

	os.Setenv("STATISTICS_SERVICE_PORT", "8082")

	conf, err := NewServiceConfigFromEnv()

	if err != nil {
		t.Errorf("NewServiceConfigFromEnv failed: %v", err)
	}

	if conf.Port != 8082 {
		t.Errorf("NewServiceConfigFromEnv failed: port mismatch")
	}

	os.Setenv("STATISTICS_SERVICE_PORT", "12345")

	conf, err = NewServiceConfigFromEnv()

	if err != nil {
		t.Errorf("NewServiceConfigFromEnv failed: %v", err)
	}

	if conf.Port != 12345 {
		t.Errorf("NewServiceConfigFromEnv failed: port mismatch")
	}
}
