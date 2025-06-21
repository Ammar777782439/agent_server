package repository

import (
	"agent_server/internal/model"
	"time"

	"gorm.io/gorm"
)

// AgentRepository defines the interface for data operations.
// Using an interface allows us to easily mock the database for testing.	
 type AgentRepository interface {
	FindAgentByID(agentID string) (*model.Agent, error)
	CreateAgent(agent *model.Agent) error
	UpdateAgent(agent *model.Agent) error
	UpdateHeartbeat(agentID, ip string) (int64, error)
	CreateFirewallRules(rules []model.FirewallRule) error
	CreateInstalledApps(apps []model.InstalledApplication) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewAgentRepository creates a new repository with a GORM connection.
func NewAgentRepository(db *gorm.DB) AgentRepository {
	return &gormRepository{db: db}
}

func (r *gormRepository) FindAgentByID(agentID string) (*model.Agent, error) {
	var agent model.Agent
	if err := r.db.Where("agent_id = ?", agentID).First(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *gormRepository) CreateAgent(agent *model.Agent) error {
	return r.db.Create(agent).Error
}

func (r *gormRepository) UpdateAgent(agent *model.Agent) error {
	return r.db.Save(agent).Error
}

func (r *gormRepository) UpdateHeartbeat(agentID, ip string) (int64, error) {
	updates := map[string]interface{}{
		"status":        "ONLINE",
		"last_seen":     time.Now(),
		"last_known_ip": ip,
	}
	result := r.db.Model(&model.Agent{}).Where("agent_id = ?", agentID).Updates(updates)
	return result.RowsAffected, result.Error
}

func (r *gormRepository) CreateFirewallRules(rules []model.FirewallRule) error {
	return r.db.Create(&rules).Error
}

func (r *gormRepository) CreateInstalledApps(apps []model.InstalledApplication) error {
	return r.db.Create(&apps).Error
}
