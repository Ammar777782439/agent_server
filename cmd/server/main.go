// agent_server/cmd/server/main.go

package main

import (
	"context"
	"log"
	"net"

	"agent_server/pkg/config"
	"agent_server/pkg/database"
	pb "agent_server/agent_server/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

const (
	port = ":50051"
	// سيطلب السيرفر من العميل إرسال نبضة كل 5 دقائق (300 ثانية)
	reportIntervalSeconds = 300
)

// agentServer هو تطبيقنا للخدمة، ويحتوي الآن على اتصال بقاعدة البيانات
type agentServer struct {
	pb.UnimplementedAgentServiceServer
	db *gorm.DB
}

// newServer ينشئ نسخة جديدة من السيرفر مع اتصال قاعدة البيانات
func newServer(db *gorm.DB) *agentServer {
	return &agentServer{db: db}
}

// --- دوال التحويل بين Protobuf و GORM ---

// toDBAgent يحول من protobuf Agent إلى GORM Agent
func toDBAgent(pbAgent *pb.Agent) *database.Agent {
	return &database.Agent{
		ID:            pbAgent.GetAgentId(),
		Hostname:      pbAgent.GetHostname(),
		OSName:        pbAgent.GetOsName(),
		OSVersion:     pbAgent.GetOsVersion(),
		KernelVersion: pbAgent.GetKernelVersion(),
		CPUCores:      pbAgent.GetCpuCores(),
		MemoryGB:      pbAgent.GetMemoryGb(),
		DiskSpaceGB:   pbAgent.GetDiskSpaceGb(),
		Status:        pbAgent.GetStatus().String(), // تحويل الـ enum إلى نص
		LastSeen:      pbAgent.GetLastSeen().AsTime(),
		LastKnownIP:   pbAgent.GetLastKnownIp(),
	}
}

// toPBAgent يحول من GORM Agent إلى protobuf Agent
func toPBAgent(dbAgent *database.Agent) *pb.Agent {
	return &pb.Agent{
		AgentId:       dbAgent.ID,
		Hostname:      dbAgent.Hostname,
		OsName:        dbAgent.OSName,
		OsVersion:     dbAgent.OSVersion,
		KernelVersion: dbAgent.KernelVersion,
		CpuCores:      dbAgent.CPUCores,
		MemoryGb:      dbAgent.MemoryGB,
		DiskSpaceGb:   dbAgent.DiskSpaceGB,
		Status:        pb.AgentStatus(pb.AgentStatus_value[dbAgent.Status]), // تحويل النص إلى enum
		LastSeen:      timestamppb.New(dbAgent.LastSeen),
		LastKnownIp:   dbAgent.LastKnownIP,
	}
}

// --- دوال RPC المنفذة ---

func (s *agentServer) RegisterAgent(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	details := req.GetAgentDetails()
	if details == nil || details.GetAgentId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Agent details and ID are required")
	}

	p, _ := peer.FromContext(ctx)
	clientIP := p.Addr.String()

	log.Printf("Received registration request from Agent ID: %s at IP: %s", details.GetAgentId(), clientIP)

	// تحديث تفاصيل العميل قبل الحفظ
	details.LastKnownIp = clientIP
	details.Status = pb.AgentStatus_ONLINE
	details.LastSeen = timestamppb.Now()

	dbAgent := toDBAgent(details)

	// استخدام GORM لحفظ البيانات (Create or Update)
	if err := s.db.Save(dbAgent).Error; err != nil {
		log.Printf("Failed to save agent to database: %v", err)
		return nil, status.Errorf(codes.Internal, "Could not save agent data")
	}

	log.Printf("Agent %s registered/updated successfully.", dbAgent.ID)

	return &pb.RegisterResponse{
		Success:              true,
		Message:              "Agent registered successfully.",
		ReportIntervalSeconds: reportIntervalSeconds,
	}, nil
}

func (s *agentServer) FindAgent(ctx context.Context, req *pb.FindAgentRequest) (*pb.FindAgentResponse, error) {
	agentID := req.GetAgentId()
	if agentID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Agent ID is required")
	}

	log.Printf("Received find request for agent: ID=%s", agentID)

	var dbAgent database.Agent
	if err := s.db.First(&dbAgent, "id = ?", agentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.FindAgentResponse{Found: false}, nil
		}
		log.Printf("Failed to find agent in database: %v", err)
		return nil, status.Errorf(codes.Internal, "Database error")
	}

	return &pb.FindAgentResponse{Found: true, Agent: toPBAgent(&dbAgent)}, nil
}

func (s *agentServer) SendHeartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	agentID := req.GetAgentId()
	if agentID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Agent ID is required")
	}

	// تحديث الحقول المحددة فقط باستخدام GORM
	updates := map[string]interface{}{
		"status":        pb.AgentStatus_ONLINE.String(),
		"last_seen":     timestamppb.Now().AsTime(),
		"last_known_ip": req.GetCurrentIp(),
	}

	result := s.db.Model(&database.Agent{}).Where("id = ?", agentID).Updates(updates)
	if result.Error != nil {
		log.Printf("Failed to update heartbeat: %v", result.Error)
		return nil, status.Errorf(codes.Internal, "Could not update agent status")
	}

	if result.RowsAffected == 0 {
		log.Printf("Heartbeat from unknown agent: %s", agentID)
		return nil, status.Errorf(codes.NotFound, "Agent not registered")
	}

	log.Printf("Heartbeat received from agent: %s", agentID)
	return &pb.HeartbeatResponse{Acknowledged: true}, nil
}

func (s *agentServer) ReportFirewallStatus(ctx context.Context, req *pb.FirewallReportRequest) (*pb.FirewallReportResponse, error) {
	agentID := req.GetAgentId()
	log.Printf("Received firewall report from agent: %s. Contains %d rules.", agentID, len(req.GetRules()))
	// ملاحظة: في تطبيق حقيقي، ستقوم بإنشاء نموذج GORM لـ FirewallRule
	// وتحفظ هذه البيانات في جدول منفصل مرتبط بالـ Agent.
	return &pb.FirewallReportResponse{Received: true}, nil
}

func (s *agentServer) ReportInstalledApps(ctx context.Context, req *pb.ReportAppsRequest) (*pb.ReportAppsResponse, error) {
	agentID := req.GetAgentId()
	log.Printf("Received installed apps report from agent: %s. Contains %d apps.", agentID, len(req.GetApplications()))
	// ملاحظة: في تطبيق حقيقي، ستقوم بإنشاء نموذج GORM لـ ApplicationInfo
	// وتحفظ هذه البيانات.
	return &pb.ReportAppsResponse{Received: true}, nil
}

func main() {
	// 1. تحميل الإعدادات
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. الاتصال بقاعدة البيانات
	db, err := database.ConnectDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection successful.")

	// 3. بدء السيرفر
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAgentServiceServer(grpcServer, newServer(db)) // تمرير اتصال DB

	log.Printf("gRPC server listening on %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
