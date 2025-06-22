// internal/service/mapper.go

package service

import (
	"agent_server/internal/model"
	agentpb "agent_server/gen/agent/v1"
	commonpb "agent_server/gen/common/v1"
)

// mapProtoToModelAgent converts a Protobuf AgentInfo to a GORM model Agent.
func mapProtoToModelAgent(p *agentpb.AgentInfo) *model.Agent {
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
		LastKnownIP:   "", // AgentInfo لا يحتوي LastKnownIP مباشرة
	}
}

// mapModelToProtoAgent converts a GORM model Agent to a Protobuf AgentInfo.
func mapModelToProtoAgent(m *model.Agent) *agentpb.AgentInfo {
	if m == nil {
		return nil
	}
	return &agentpb.AgentInfo{
		AgentId:       m.AgentID,
		Hostname:      m.Hostname,
		OsName:        m.OSName,
		OsVersion:     m.OSVersion,
		KernelVersion: m.KernelVersion,
		CpuCores:      m.CPUCores,
		MemoryGb:      m.MemoryGB,
		DiskSpaceGb:   m.DiskSpaceGB,
		// لا يوجد LastKnownIp أو Status أو LastSeen في AgentInfo حسب البروتو الحالي
	}
}

// mapProtoToModelFirewallRules converts a slice of Protobuf FirewallRules to GORM models.
func mapProtoToModelFirewallRules(protoRules []*commonpb.FirewallRule) []model.FirewallRule {
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
func mapProtoToModelInstalledApps(protoApps []*commonpb.ApplicationInfo) []model.InstalledApplication {
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

