// agent_server/cmd/server/main.go

package main

import (
	"context"
	"log"
	"net"

	pb "agent_server/agent_server/proto" // استيراد الكود الذي تم توليده من مجلد proto

	"google.golang.org/grpc"
)

// تعريف هيكل السيرفر. يجب أن يتضمن الواجهة غير المنفذة (unimplemented) لضمان التوافق المستقبلي
type agentServer struct {
	pb.UnimplementedAgentServiceServer
}

// تنفيذ دالة تسجيل الـ Agent
func (s *agentServer) RegisterAgent(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.Printf("Received registration request from agent: ID=%s, Hostname=%s, IP=%s", req.GetAgentId(), req.GetHostname(), req.GetIpAddress())
	
	// هنا يمكنك كتابة منطق حفظ بيانات الـ Agent في قاعدة بيانات
	// على سبيل المثال: db.SaveAgent(req)

	return &pb.RegisterResponse{
		Success: true,
		Message: "Agent registered successfully",
	}, nil
}

// تنفيذ دالة استقبال تقرير جدار الحماية
func (s *agentServer) ReportFirewallStatus(ctx context.Context, req *pb.FirewallReportRequest) (*pb.FirewallReportResponse, error) {
	log.Printf("Received firewall report from agent: ID=%s. Report contains %d rules.", req.GetAgentId(), len(req.GetRules()))
	
	// هنا يمكنك تحليل القواعد وتخزينها أو إطلاق تنبيهات بناءً عليها
	for _, rule := range req.GetRules() {
		log.Printf("  - Rule: %s, Port: %s, Protocol: %s, Action: %s", rule.GetName(), rule.GetPort(), rule.GetProtocol(), rule.GetAction())
	}

	return &pb.FirewallReportResponse{
		Received: true,
	}, nil
}

// تنفيذ دالة البحث عن Agent
func (s *agentServer) FindAgent(ctx context.Context, req *pb.FindAgentRequest) (*pb.FindAgentResponse, error) {
	agentID := req.GetAgentId()
	log.Printf("Received find request for agent: ID=%s", agentID)

	// هنا يمكنك البحث عن الـ Agent في قاعدة البيانات
	// في هذا المثال، سنقوم بإرجاع بيانات وهمية إذا كان الـ ID موجوداً
	if agentID == "agent-123" {
		return &pb.FindAgentResponse{
			Found:      true,
			AgentId:    "agent-123",
			Hostname:   "prod-server-01",
			IpAddress:  "192.168.1.100",
			LastSeen:   "2025-06-21T15:00:00Z",
		}, nil
	}
	
	return &pb.FindAgentResponse{
		Found: false,
	}, nil
}

func main() {
	// تحديد المنفذ الذي سيعمل عليه السيرفر
	port := ":50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	// إنشاء سيرفر gRPC جديد
	s := grpc.NewServer()
	
	// تسجيل خدمتنا (agentServer) مع سيرفر gRPC
	// هذا يخبر سيرفر gRPC كيف يتعامل مع الطلبات الواردة
	pb.RegisterAgentServiceServer(s, &agentServer{})

	log.Printf("gRPC server listening on %s", port)

	// تشغيل السيرفر وانتظار الاتصالات
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
