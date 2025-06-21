// agent_server/pkg/config/config.go

package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// DBConfig يمثل إعدادات قاعدة البيانات
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
	TimeZone string `yaml:"timezone"`
}

// Config يمثل الهيكل الكامل لملف الإعدادات
type Config struct {
	Database DBConfig `yaml:"db"`
}

// LoadConfig يقرأ ملف الإعدادات من المسار المحدد ويحمّله
func LoadConfig(path string) (*Config, error) {
	configFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
