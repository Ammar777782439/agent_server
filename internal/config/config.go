package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config هو الهيكل الرئيسي الذي يمثل ملف الإعدادات بأكمله
type Config struct {
	Database DBConfig `yaml:"db"`
}

// DBConfig يحتوي على إعدادات الاتصال بقاعدة البيانات
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
	TimeZone string `yaml:"timezone"`
}

// LoadConfig يقرأ ملف الإعدادات من المسار المحدد ويقوم بتحليله
func LoadConfig(path string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
