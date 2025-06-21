package main

import (
	"context"
	"errors"
	"log"
	"net"
	"time"

	"agent_server/pkg/config"
	"agent_server/pkg/database"
	pb "agent_server/agent_server/proto" // <- تأكد أن المسار ده صح بناءً على الـ go_package في الـ proto

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
	reportIntervalSeconds = 300 // سيطلب السيرفر من العميل إرسال نبضة كل 5 دقائق (300 ثانية)
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

// toPBAgent يحول من GORM Agent إلى protobuf Agent
func toPBAgent(dbAgent *database.Agent) *pb.Agent {
	return &pb.Agent{
		AgentId:       dbAgent.AgentID,
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

func (s *agentServer) FindAgent(ctx context.Context, req *pb.FindAgentRequest) (*pb.FindAgentResponse, error) {
	agentID := req.GetAgentId()
	if agentID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Agent ID is required")
	}

	log.Printf("Received find request for agent: ID=%s", agentID)

	var dbAgent database.Agent
	if err := s.db.Where("agent_id = ?", agentID).First(&dbAgent).Error; err != nil {
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
		"status":      pb.AgentStatus_ONLINE.String(),
		"last_seen":   timestamppb.Now().AsTime(),
		"last_known_ip": req.GetCurrentIp(),
	}

	result := s.db.Model(&database.Agent{}).Where("agent_id = ?", agentID).Updates(updates) // استخدام agent_id هنا
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

func (s *agentServer) ReportFirewallStatus(ctx context.Context, req *pb.FirewallStatusRequest) (*pb.FirewallStatusResponse, error) {
	log.Printf("Received firewall status report from agent %s: %d rules", req.AgentId, len(req.Rules))

	var agent database.Agent
	if err := s.db.Where("agent_id = ?", req.AgentId).First(&agent).Error; err != nil { // استخدام agent_id هنا
		log.Printf("Agent %s not found for firewall report: %v", req.AgentId, err)
		return nil, status.Errorf(codes.NotFound, "Agent not found")
	}

	for _, ruleProto := range req.Rules {
		newRule := database.FirewallRule{
			AgentID:   agent.ID, // هنا هتستخدم ID بتاع الـ DB
			Name:      ruleProto.Name,
			Port:      ruleProto.Port,
			Protocol:  ruleProto.Protocol,
			Action:    ruleProto.Action.String(),
			Direction: ruleProto.Direction.String(),
			Enabled:   ruleProto.Enabled,
		}
		if err := s.db.Create(&newRule).Error; err != nil {
			log.Printf("Failed to save firewall rule for agent %s: %v", req.AgentId, err)
			return nil, status.Errorf(codes.Internal, "Could not save firewall rule")
		}
	}

	return &pb.FirewallStatusResponse{Success: true, Message: "Firewall status received and stored"}, nil
}

func (s *agentServer) ReportInstalledApps(ctx context.Context, req *pb.InstalledAppsRequest) (*pb.InstalledAppsResponse, error) {
	log.Printf("Received installed apps report from agent %s: %d apps", req.AgentId, len(req.Apps))

	var agent database.Agent
	if err := s.db.Where("agent_id = ?", req.AgentId).First(&agent).Error; err != nil { // استخدام agent_id هنا
		log.Printf("Agent %s not found for installed apps report: %v", req.AgentId, err)
		return nil, status.Errorf(codes.NotFound, "Agent not found")
	}

	for _, appProto := range req.Apps {
		newApp := database.InstalledApplication{
			AgentID:     agent.ID, // هنا هتستخدم ID بتاع الـ DB
			Name:        appProto.Name,
			Version:     appProto.Version,
			InstallDate: appProto.InstallDate.AsTime(),
			Publisher:   appProto.Publisher,
		}
		if err := s.db.Create(&newApp).Error; err != nil {
			log.Printf("Failed to save installed application for agent %s: %v", req.AgentId, err)
			return nil, status.Errorf(codes.Internal, "Could not save installed application")
		}
	}

	return &pb.InstalledAppsResponse{Success: true, Message: "Installed apps received and stored"}, nil
}

// -----------------------------------------------------------
// دالة التسجيل RegisterAgent
func (s *agentServer) RegisterAgent(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) { // <<< تم التعديل هنا
	agentDetails := req.GetAgentDetails() // الوصول لـ agent_details من الـ Request
	if agentDetails == nil || agentDetails.GetAgentId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Agent details and Agent ID are required for registration")
	}

	log.Printf("Received registration request for agent: %s", agentDetails.GetAgentId())

	// التحقق مما إذا كان العميل موجودًا بالفعل
	var existingAgent database.Agent
	// استخدام agent_id للبحث في قاعدة البيانات
	if err := s.db.Where("agent_id = ?", agentDetails.GetAgentId()).First(&existingAgent).Error; err == nil {
		// العميل موجود، قم بتحديث حالته
		log.Printf("Agent %s already exists. Updating status.", agentDetails.GetAgentId())
		existingAgent.Status = pb.AgentStatus_ONLINE.String() // تحويل ENUM إلى نص
		existingAgent.LastSeen = time.Now()
		// existingAgent.LastKnownIP = getRequesterIP(ctx) // getRequesterIP is not defined, commenting out for now
		if err := s.db.Save(&existingAgent).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update agent: %v", err)
		}
		// <<< تم التعديل هنا ليتناسب مع RegisterResponse الجديد
		return &pb.RegisterResponse{Success: true, Message: "Agent updated successfully", ReportIntervalSeconds: reportIntervalSeconds}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// خطأ آخر في قاعدة البيانات
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	// إنشاء عميل جديد
	newAgent := database.Agent{
		AgentID:       agentDetails.GetAgentId(),
		Hostname:      agentDetails.GetHostname(),
		OSName:        agentDetails.GetOsName(),
		OSVersion:     agentDetails.GetOsVersion(),
		KernelVersion: agentDetails.GetKernelVersion(),
		CPUCores:      agentDetails.GetCpuCores(),
		MemoryGB:      agentDetails.GetMemoryGb(),
		DiskSpaceGB:   agentDetails.GetDiskSpaceGb(),
		Status:        pb.AgentStatus_ONLINE.String(), // تحويل ENUM إلى نص
		LastSeen:      time.Now(),
		// LastKnownIP:   getRequesterIP(ctx), // getRequesterIP is not defined, commenting out for now
	}

	if err := s.db.Create(&newAgent).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create agent: %v", err)
	}

	log.Printf("Successfully registered new agent: %s with DB ID: %d", newAgent.AgentID, newAgent.ID)
	// <<< تم التعديل هنا ليتناسب مع RegisterResponse الجديد
	return &pb.RegisterResponse{Success: true, Message: "Agent registered successfully", ReportIntervalSeconds: reportIntervalSeconds}, nil
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
	reflection.Register(grpcServer)                          // <-- 2. تسجيل خدمة الانعكاس
	log.Printf("gRPC server listening on %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}