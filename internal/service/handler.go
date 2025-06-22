// internal/service/handler.go

package service

import (
	pb "agent_server/agent_server/proto"
	"agent_server/internal/model"
	"agent_server/internal/usecase"
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

const (
	reportIntervalSeconds = 300 // 5 minutes
)

// AgentServer now depends on the use case layer, not the repository.
type AgentServer struct {
	pb.UnimplementedAgentServiceServer
	agentLogic usecase.AgentUseCase
	db         *gorm.DB
}

// NewAgentServer creates a new AgentServer with the injected business logic layer.
func NewAgentServer(logic usecase.AgentUseCase, db *gorm.DB) *AgentServer {
	
		return &AgentServer{agentLogic: logic, db: db}
	}
	
// RegisterAgent is now a thin layer that validates, maps, and calls the logic layer.
func (s *AgentServer) RegisterAgent(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {

	agentDetailsProto := req.GetAgentDetails()
	if agentDetailsProto == nil || agentDetailsProto.GetAgentId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Agent details and Agent ID are required")
	}

	log.Printf("Registration request for agent: %s", agentDetailsProto.GetAgentId())

	agentModel := mapProtoToModelAgent(agentDetailsProto)

	_, err := s.agentLogic.RegisterAgent(agentModel)
	if err != nil {
		log.Printf("Failed to process agent registration for %s: %v", agentModel.AgentID, err)
		return nil, status.Errorf(codes.Internal, "failed to register agent: %v", err)
	}

	log.Printf("Successfully registered agent: %s", agentModel.AgentID)
	return &pb.RegisterResponse{Success: true, Message: "Agent registered successfully", ReportIntervalSeconds: reportIntervalSeconds}, nil
}

func (s *AgentServer) SendHeartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if req.GetAgentId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Agent ID is required")
	}

	rowsAffected, err := s.agentLogic.ProcessHeartbeat(req.GetAgentId(), req.GetCurrentIp())
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
	rulesModel := mapProtoToModelFirewallRules(req.GetRules())

	err := s.agentLogic.StoreFirewallRules(req.GetAgentId(), rulesModel)
	if err != nil {
		log.Printf("Failed to save firewall rules for agent %s: %v", req.GetAgentId(), err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Agent not found")
		}
		return nil, status.Errorf(codes.Internal, "Could not save firewall rules")
	}

	return &pb.FirewallStatusResponse{Success: true, Message: "Firewall status received"}, nil
}

func (s *AgentServer) ReportInstalledApps(ctx context.Context, req *pb.InstalledAppsRequest) (*pb.InstalledAppsResponse, error) {
	appsModel := mapProtoToModelInstalledApps(req.GetApps())

	err := s.agentLogic.StoreInstalledApps(req.GetAgentId(), appsModel)
	if err != nil {
		log.Printf("Failed to save installed apps for agent %s: %v", req.GetAgentId(), err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Agent not found")
		}
		return nil, status.Errorf(codes.Internal, "Could not save installed apps")
	}

	return &pb.InstalledAppsResponse{Success: true, Message: "Installed apps received"}, nil
}

func (s *AgentServer) FindAgent(ctx context.Context, req *pb.FindAgentRequest) (*pb.FindAgentResponse, error) {
	agentModel, err := s.agentLogic.GetAgentByID(req.GetAgentId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &pb.FindAgentResponse{Found: false}, nil
		}
		return nil, status.Errorf(codes.Internal, "Database error")
	}

	agentProto := mapModelToProtoAgent(agentModel)

	return &pb.FindAgentResponse{Found: true, Agent: agentProto}, nil
}

// TaskChannel هي الدالة التي تنفذ الواجهة الجديدة لـ gRPC
// وتتعامل مع قناة الاتصال المفتوحة مع العميل.
func (s *AgentServer) TaskChannel(stream pb.AgentService_TaskChannelServer) error {
	log.Println("✅ Agent connected to the Task Channel!")

	// 🔴🔴🔴 المنطق الجديد: قراءة مهمة من قاعدة البيانات وإرسالها مرة واحدة 🔴🔴🔴
	var query model.Query
	// ابحث عن أول استعلام متاح في قاعدة البيانات
	if err := s.db.First(&query).Error; err != nil {
		log.Printf("Could not find any query in the database: %v", err)
		// يمكنك إغلاق الاتصال أو الانتظار
	} else {
		// تم العثور على استعلام، قم ببناء المهمة
		task := &pb.TaskRequest{
			TaskId:      fmt.Sprintf("task-%d", query.ID), // استخدام معرف الاستعلام كمعرف للمهمة
			CommandName: query.CommandName,
			// يمكنك إضافة المتغيرات هنا إذا لزم الأمر
		}
		log.Printf("✔️ Found query [ID: %d]. Sending task '%s' to agent.", query.ID, task.CommandName)
		if err := stream.Send(task); err != nil {
			log.Printf("Error sending initial task to agent: %v", err)
		}
	}
	// 🔴🔴🔴 نهاية المنطق الجديد 🔴🔴🔴

	// حلقة للاستماع للنتائج القادمة من العميل
	for {
		result, err := stream.Recv()
		if err == io.EOF {
			log.Println("Client closed the stream.")
			return nil
		}
		if err != nil {
			log.Printf("Error receiving result from agent: %v", err)
			return err
		}

		// TODO: في المستقبل، سنقوم بتخزين هذه النتيجة في قاعدة البيانات
		log.Printf("Received result from agent for Task [ID: %s], Success: %v", result.TaskId, result.Success)
	}
}