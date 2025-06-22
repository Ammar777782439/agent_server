// internal/service/handler.go

package service

import (
	"agent_server/internal/usecase"

	pb "agent_server/gen/service/v1"
	agentpb "agent_server/gen/agent/v1"
	"log"
)

const (
	reportIntervalSeconds = 300 // 5 minutes
)

// AgentServer now depends on the use case layer, not the repository.
type AgentServer struct {
	pb.UnimplementedAgentServiceServer
	agentLogic usecase.AgentUseCase
}

// NewAgentServer creates a new AgentServer with the injected business logic layer.
func NewAgentServer(logic usecase.AgentUseCase) *AgentServer {
	return &AgentServer{agentLogic: logic}
}

// CommandStream implements the bidirectional streaming RPC for agent-server communication.
func (s *AgentServer) CommandStream(stream pb.AgentService_CommandStreamServer) error {
	log.Println("CommandStream started")
	for {
		in, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				log.Println("Stream closed by client")
				return nil
			}
			log.Printf("Error receiving from stream: %v", err)
			return err
		}

		// Switch on message type (مطابق لتعريف AgentToServer في البروتو)
		switch payload := in.Payload.(type) {
		case *agentpb.AgentToServer_AgentInfo:
			log.Printf("[AgentInfo] AgentID: %s Hostname: %s", payload.AgentInfo.AgentId, payload.AgentInfo.Hostname)
			// هنا من الممكن معالجة التسجيل أو تحديث المعلومات
		case *agentpb.AgentToServer_Heartbeat:
			log.Printf("[Heartbeat] from agent stream: message_id=%s", in.MessageId)
			// هنا تعالج نبضات القلب
		case *agentpb.AgentToServer_FirewallRulesReport:
			log.Printf("[FirewallRulesReport] received %d rules", len(payload.FirewallRulesReport.Rules))
			// معالجة قواعد جدار الحماية
		case *agentpb.AgentToServer_InstalledAppsReport:
			log.Printf("[InstalledAppsReport] received %d apps", len(payload.InstalledAppsReport.Apps))
			// معالجة التطبيقات المثبتة
		case *agentpb.AgentToServer_ScanReport:
			log.Printf("[ScanReport] scan for request_id=%s, success=%v", payload.ScanReport.OriginalRequestId, payload.ScanReport.Success)
			// معالجة نتائج الفحص
		default:
			log.Printf("Unknown or unhandled message type from agent")
		}

		// Example: send a command to agent (server-initiated)
		// out := &serverpb.ServerToAgent{...}
		// err = stream.Send(out)
		// if err != nil {
		// 	log.Printf("Error sending to agent: %v", err)
		// }
	}
}