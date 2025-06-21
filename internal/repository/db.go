package repository

import (
	"agent_server/internal/config"
	"agent_server/internal/model"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB يتصل بقاعدة البيانات ويقوم بترحيل النموذج تلقائيًا
func ConnectDB(cfg *config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode, cfg.TimeZone)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// AutoMigrate سيقوم بإنشاء الجداول إذا لم تكن موجودة
	err = db.AutoMigrate(&model.Agent{}, &model.FirewallRule{}, &model.InstalledApplication{})
	if err != nil {
		// إذا فشل، حاول إغلاق الاتصال قبل إرجاع الخطأ
		sqlDB, _ := db.DB()
		sqlDB.Close()
		return nil, fmt.Errorf("failed to auto-migrate database: %w", err)
	}

	return db, nil
}
