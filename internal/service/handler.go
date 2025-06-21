package service

import (
	"context"
	"errors"
	"log"
	"time"

	"agent_server/internal/model"
	"agent_server/internal/repository"
	pb "agent_server/agent_server/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

const (
	reportIntervalSeconds = 300 // 5 minutes
)

// AgentServer implements the gRPC server.
// It depends on the AgentRepository interface for data operations.
type AgentServer struct {
	pb.UnimplementedAgentServiceServer
	repo repository.AgentRepository
}

// NewAgentServer creates a new AgentServer.
func NewAgentServer(repo repository.AgentRepository) *AgentServer {
	return &AgentServer{repo: repo}
}

func (s *AgentServer) RegisterAgent(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	agentDetails := req.GetAgentDetails()
	if agentDetails == nil || agentDetails.GetAgentId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Agent details and Agent ID are required")
	}

	log.Printf("Registration request for agent: %s", agentDetails.GetAgentId())

	existingAgent, err := s.repo.FindAgentByID(agentDetails.GetAgentId())
	if err == nil {
		// Agent exists, update status
		log.Printf("Agent %s already exists. Updating status.", existingAgent.AgentID)
		existingAgent.Status = pb.AgentStatus_ONLINE.String()
		existingAgent.LastSeen = time.Now()
		if err := s.repo.UpdateAgent(existingAgent); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update agent: %v", err)
		}
		return &pb.RegisterResponse{Success: true, Message: "Agent updated successfully", ReportIntervalSeconds: reportIntervalSeconds}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// A different database error occurred
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	// Agent does not exist, create new one
	newAgent := &model.Agent{
		AgentID:       agentDetails.GetAgentId(),
		Hostname:      agentDetails.GetHostname(),
		OSName:        agentDetails.GetOsName(),
		OSVersion:     agentDetails.GetOsVersion(),
		KernelVersion: agentDetails.GetKernelVersion(),
		CPUCores:      agentDetails.GetCpuCores(),
		MemoryGB:      agentDetails.GetMemoryGb(),
		DiskSpaceGB:   agentDetails.GetDiskSpaceGb(),
		Status:        pb.AgentStatus_ONLINE.String(),
		LastSeen:      time.Now(),
	}

	if err := s.repo.CreateAgent(newAgent); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create agent: %v", err)
	}

	log.Printf("Successfully registered new agent: %s", newAgent.AgentID)
	return &pb.RegisterResponse{Success: true, Message: "Agent registered successfully", ReportIntervalSeconds: reportIntervalSeconds}, nil
}

func (s *AgentServer) SendHeartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if req.GetAgentId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Agent ID is required")
	}

	rowsAffected, err := s.repo.UpdateHeartbeat(req.GetAgentId(), req.GetCurrentIp())
	if err != nil {
		log.Printf("Failed to update heartbeat for agent %s: %v", req.GetAgentId(), err)
		return nil, status.Errorf(codes.Internal, "Could not update agent status")
	}

	if rowsAffected == 0 {
		log.Printf("Heartbeat from unknown agent: %s", req.GetAgentId())
		return nil, status.Errorf(codes.NotFound, "Agent not registered")
	}

	log.Printf("Heartbeat received from agent: %s", req.GetAgentId())
	return &pb.HeartbeatResponse{Acknowledged: true}, nil
}

func (s *AgentServer) ReportFirewallStatus(ctx context.Context, req *pb.FirewallStatusRequest) (*pb.FirewallStatusResponse, error) {
	agent, err := s.repo.FindAgentByID(req.AgentId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Agent not found")
	}

	var rules []model.FirewallRule
	for _, ruleProto := range req.Rules {
		rules = append(rules, model.FirewallRule{
			AgentID:   agent.ID,
			Name:      ruleProto.Name,
			Port:      ruleProto.Port,
			Protocol:  ruleProto.Protocol,
			Action:    ruleProto.Action.String(),
			Direction: ruleProto.Direction.String(),
			Enabled:   ruleProto.Enabled,
		})
	}

	if err := s.repo.CreateFirewallRules(rules); err != nil {
		log.Printf("Failed to save firewall rules for agent %s: %v", req.AgentId, err)
		return nil, status.Errorf(codes.Internal, "Could not save firewall rules")
	}

	return &pb.FirewallStatusResponse{Success: true, Message: "Firewall status received"}, nil
}

func (s *AgentServer) ReportInstalledApps(ctx context.Context, req *pb.InstalledAppsRequest) (*pb.InstalledAppsResponse, error) {
	agent, err := s.repo.FindAgentByID(req.AgentId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Agent not found")
	}

	var apps []model.InstalledApplication
	for _, appProto := range req.Apps {
		apps = append(apps, model.InstalledApplication{
			AgentID:     agent.ID,
			Name:        appProto.Name,
			Version:     appProto.Version,
			InstallDate: appProto.InstallDate.AsTime(),
			Publisher:   appProto.Publisher,
		})
	}

	if err := s.repo.CreateInstalledApps(apps); err != nil {
		log.Printf("Failed to save installed apps for agent %s: %v", req.AgentId, err)
		return nil, status.Errorf(codes.Internal, "Could not save installed apps")
	}

	return &pb.InstalledAppsResponse{Success: true, Message: "Installed apps received"}, nil
}

func (s *AgentServer) FindAgent(ctx context.Context, req *pb.FindAgentRequest) (*pb.FindAgentResponse, error) {
	agent, err := s.repo.FindAgentByID(req.GetAgentId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &pb.FindAgentResponse{Found: false}, nil
		}
		return nil, status.Errorf(codes.Internal, "Database error")
	}

	pbAgent := &pb.Agent{
		AgentId:       agent.AgentID,
		Hostname:      agent.Hostname,
		OsName:        agent.OSName,
		OsVersion:     agent.OSVersion,
		KernelVersion: agent.KernelVersion,
		CpuCores:      agent.CPUCores,
		MemoryGb:      agent.MemoryGB,
		DiskSpaceGb:   agent.DiskSpaceGB,
		Status:        pb.AgentStatus(pb.AgentStatus_value[agent.Status]),
		LastSeen:      timestamppb.New(agent.LastSeen),
		LastKnownIp:   agent.LastKnownIP,
	}

	return &pb.FindAgentResponse{Found: true, Agent: pbAgent}, nil
}
