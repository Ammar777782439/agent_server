// agent_server/pkg/database/database.go

package database

import (
	"fmt"
	"time"

	"agent_server/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Agent هو نموذج GORM الذي يمثل جدول العملاء في قاعدة البيانات
type Agent struct {
	ID             string    `gorm:"primaryKey"`
	Hostname       string
	OSName         string
	OSVersion      string
	KernelVersion  string
	CPUCores       int32
	MemoryGB       float64
	DiskSpaceGB    float64
	Status         string // "ONLINE", "OFFLINE", etc.
	LastSeen       time.Time
	LastKnownIP    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ConnectDB يتصل بقاعدة البيانات باستخدام إعدادات DSN
func ConnectDB(cfg *config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode, cfg.TimeZone)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// AutoMigrate سيقوم بإنشاء جدول 'agents' إذا لم يكن موجودًا
	err = db.AutoMigrate(&Agent{})
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database: %w", err)
	}

	return db, nil
}
