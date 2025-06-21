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
	gorm.Model
	AgentID        string    `gorm:"uniqueIndex;size:100"` // المعرّف النصي الفريد للعميل
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
}

// FirewallRule يمثل قاعدة جدار حماية واحدة يبلغ عنها العميل
type FirewallRule struct {
	gorm.Model
	AgentID   uint   // مفتاح خارجي لنموذج Agent
	Name      string `gorm:"size:255"`
	Port      string `gorm:"size:50"`
	Protocol  string `gorm:"size:50"`
	Action    string `gorm:"size:50"` // مثال: "ALLOW", "DENY"
	Direction string `gorm:"size:50"` // مثال: "IN", "OUT"
	Enabled   bool
}

// InstalledApplication يمثل تطبيقًا واحدًا مثبتًا على نظام العميل
type InstalledApplication struct {
	gorm.Model
	AgentID     uint      // مفتاح خارجي لنموذج Agent
	Name        string    `gorm:"size:255"`
	Version     string    `gorm:"size:100"`
	InstallDate time.Time
	Publisher   string    `gorm:"size:255"`
}

// ConnectDB يتصل بقاعدة البيانات ويقوم بترحيل النموذج تلقائيًات DSN
func ConnectDB(cfg *config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode, cfg.TimeZone)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// AutoMigrate سيقوم بإنشاء جدول 'agents' إذا لم يكن موجودًا
	err = db.AutoMigrate(&Agent{}, &FirewallRule{}, &InstalledApplication{})
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database: %w", err)
	}

	return db, nil
}
