package config

import (
	"os"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	for _, k := range []string{"HTTP_PORT", "DEFAULT_RESULT_LIMIT"} {
		os.Unsetenv(k)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPPort != 8082 {
		t.Errorf("port: %d", cfg.HTTPPort)
	}
	if cfg.Timezone != "Asia/Shanghai" {
		t.Errorf("tz: %s", cfg.Timezone)
	}
}

func TestLoadConfigOverride(t *testing.T) {
	os.Setenv("HTTP_PORT", "9090")
	defer os.Unsetenv("HTTP_PORT")
	cfg, _ := LoadConfig()
	if cfg.HTTPPort != 9090 {
		t.Errorf("port: %d", cfg.HTTPPort)
	}
}
