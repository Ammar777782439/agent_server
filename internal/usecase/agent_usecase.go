// internal/usecase/agent_usecase.go

package usecase

import (
	"agent_server/internal/model"
	"agent_server/internal/repository"
	"errors"
	"time"

	"gorm.io/gorm"
)

// AgentUseCase defines the contract for agent business logic.
type AgentUseCase interface {
	RegisterAgent(agent *model.Agent) (*model.Agent, error)
	GetAgentByID(agentID string) (*model.Agent, error)
	ProcessHeartbeat(agentID, ip string) (int64, error)
	StoreFirewallRules(agentID string, rules []model.FirewallRule) error
	StoreInstalledApps(agentID string, apps []model.InstalledApplication) error
}

type agentUseCase struct {
	repo repository.AgentRepository
}

// NewAgentUseCase creates a new instance of the agent use case layer.
func NewAgentUseCase(repo repository.AgentRepository) AgentUseCase {
	return &agentUseCase{repo: repo}
}

// RegisterAgent handles the core logic of registering an agent.
// It checks if the agent exists, then updates or creates it.
func (uc *agentUseCase) RegisterAgent(agent *model.Agent) (*model.Agent, error) {
	existingAgent, err := uc.repo.FindAgentByID(agent.AgentID)

	if err == nil {
		// Agent exists, update it with new static info and set status.
		existingAgent.Hostname = agent.Hostname
		existingAgent.OSName = agent.OSName
		existingAgent.OSVersion = agent.OSVersion
		existingAgent.KernelVersion = agent.KernelVersion
		existingAgent.CPUCores = agent.CPUCores
		existingAgent.MemoryGB = agent.MemoryGB
		existingAgent.DiskSpaceGB = agent.DiskSpaceGB
		existingAgent.Status = "ONLINE"
		existingAgent.LastSeen = time.Now()
		
		err = uc.repo.UpdateAgent(existingAgent)
		return existingAgent, err
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// This is a real database error.
		return nil, err
	}

	// Agent does not exist, create a new one.
	agent.Status = "ONLINE"
	agent.LastSeen = time.Now()
	err = uc.repo.CreateAgent(agent)
	return agent, err
}

// GetAgentByID retrieves a single agent.
func (uc *agentUseCase) GetAgentByID(agentID string) (*model.Agent, error) {
	return uc.repo.FindAgentByID(agentID)
}

// ProcessHeartbeat updates the agent's status.
func (uc *agentUseCase) ProcessHeartbeat(agentID, ip string) (int64, error) {
	return uc.repo.UpdateHeartbeat(agentID, ip)
}

// StoreFirewallRules validates and stores firewall rules for an agent.
func (uc *agentUseCase) StoreFirewallRules(agentID string, rules []model.FirewallRule) error {
	// Business logic: ensure the agent exists before adding rules.
	agent, err := uc.repo.FindAgentByID(agentID)
	if err != nil {
		return err // Return error if agent not found or other DB issue.
	}

	// Assign the agent's primary key to each rule.
	for i := range rules {
		rules[i].AgentID = agent.ID
	}

	return uc.repo.CreateFirewallRules(rules)
}

// StoreInstalledApps validates and stores installed apps for an agent.
func (uc *agentUseCase) StoreInstalledApps(agentID string, apps []model.InstalledApplication) error {
	// Business logic: ensure the agent exists before adding apps.
	agent, err := uc.repo.FindAgentByID(agentID)
	if err != nil {
		return err
	}

	// Assign the agent's primary key to each app.
	for i := range apps {
		apps[i].AgentID = agent.ID
	}

	return uc.repo.CreateInstalledApps(apps)
}