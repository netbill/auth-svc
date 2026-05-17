package rest_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	"go.yaml.in/yaml/v3"
)

type config struct {
	REST struct {
		BaseURL string `yaml:"base_url"`
	} `yaml:"rest"`

	Load struct {
		Accounts int `yaml:"accounts"`
	} `yaml:"load"`
}

var (
	testCfg    *config
	testClient *http.Client
)

func TestMain(m *testing.M) {
	cfg, err := loadConfig("test_config.yaml")
	if err != nil {
		panic("load test config: " + err.Error())
	}
	testCfg = cfg

	testClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	os.Exit(m.Run())
}

func loadConfig(path string) (*config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
