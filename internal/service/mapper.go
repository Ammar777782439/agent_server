// internal/service/mapper.go

package service

import (
	"agent_server/internal/model"
	pb "agent_server/agent_server/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// mapProtoToModelAgent converts a Protobuf Agent to a GORM model Agent.
func mapProtoToModelAgent(p *pb.Agent) *model.Agent {
	if p == nil {
		return nil
	}
	return &model.Agent{
		AgentID:       p.GetAgentId(),
		Hostname:      p.GetHostname(),
		OSName:        p.GetOsName(),
		OSVersion:     p.GetOsVersion(),
		KernelVersion: p.GetKernelVersion(),
		CPUCores:      p.GetCpuCores(),
		MemoryGB:      p.GetMemoryGb(),
		DiskSpaceGB:   p.GetDiskSpaceGb(),
		LastKnownIP:   p.GetLastKnownIp(),
	}
}

// mapModelToProtoAgent converts a GORM model Agent to a Protobuf Agent.
func mapModelToProtoAgent(m *model.Agent) *pb.Agent {
	if m == nil {
		return nil
	}
	return &pb.Agent{
		AgentId:       m.AgentID,
		Hostname:      m.Hostname,
		OsName:        m.OSName,
		OsVersion:     m.OSVersion,
		KernelVersion: m.KernelVersion,
		CpuCores:      m.CPUCores,
		MemoryGb:      m.MemoryGB,
		DiskSpaceGb:   m.DiskSpaceGB,
		Status:        pb.AgentStatus(pb.AgentStatus_value[m.Status]),
		LastSeen:      timestamppb.New(m.LastSeen),
		LastKnownIp:   m.LastKnownIP,
	}
}

// mapProtoToModelFirewallRules converts a slice of Protobuf FirewallRules to GORM models.
func mapProtoToModelFirewallRules(protoRules []*pb.FirewallRule) []model.FirewallRule {
	modelRules := make([]model.FirewallRule, 0, len(protoRules))
	for _, p := range protoRules {
		modelRules = append(modelRules, model.FirewallRule{
			Name:      p.GetName(),
			Port:      p.GetPort(),
			Protocol:  p.GetProtocol(),
			Action:    p.GetAction().String(),
			Direction: p.GetDirection().String(),
			Enabled:   p.GetEnabled(),
		})
	}
	return modelRules
}

// mapProtoToModelInstalledApps converts a slice of Protobuf Applications to GORM models.
func mapProtoToModelInstalledApps(protoApps []*pb.ApplicationInfo) []model.InstalledApplication {
	modelApps := make([]model.InstalledApplication, 0, len(protoApps))
	for _, p := range protoApps {
		modelApps = append(modelApps, model.InstalledApplication{
			Name:        p.GetName(),
			Version:     p.GetVersion(),
			InstallDate: p.GetInstallDate().AsTime(),
			Publisher:   p.GetPublisher(),
		})
	}
	return modelApps
}

