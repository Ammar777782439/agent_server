package model

import (
	"time"

	"gorm.io/gorm"
)

// Agent هو نموذج GORM الذي يمثل جدول العملاء في قاعدة البيانات
type Agent struct {
	ID            uint `gorm:"primaryKey;autoIncrement"`
	AgentID       string `gorm:"uniqueIndex"`
	Hostname      string
	OSName        string
	OSVersion     string
	KernelVersion string
	CPUCores      int32
	MemoryGB      float64
	DiskSpaceGB   float64
	Status        string
	LastSeen      time.Time
	LastKnownIP   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

// FirewallRule يمثل قاعدة جدار حماية واحدة يبلغ عنها العميل
type FirewallRule struct {
	gorm.Model
	AgentID   uint
	Name      string `gorm:"size:255"`
	Port      string `gorm:"size:50"`
	Protocol  string `gorm:"size:50"`
	Action    string `gorm:"size:50"`
	Direction string `gorm:"size:50"`
	Enabled   bool
}

// InstalledApplication يمثل تطبيقًا واحدًا مثبتًا على نظام العميل
type InstalledApplication struct {
	gorm.Model
	AgentID     uint
	Name        string    `gorm:"size:255"`
	Version     string    `gorm:"size:100"`
	InstallDate time.Time
	Publisher   string    `gorm:"size:255"`
}
