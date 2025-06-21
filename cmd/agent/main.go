// agent_server/cmd/agent/main.go

package main

import (
	"context"
	"log"
	"time"

	pb "agent_server/agent_server/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// الاتصال بالسيرفر على المنفذ 50051
	// نستخدم insecure.NewCredentials() لأننا لا نستخدم SSL/TLS في هذا المثال
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	// إنشاء عميل جديد لخدمة AgentService
	c := pb.NewAgentServiceClient(conn)

	// إنشاء سياق (context) مع مهلة زمنية (timeout)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// --- 1. تسجيل العميل ---
	log.Println("--- Registering Agent ---")
	registerResp, err := c.RegisterAgent(ctx, &pb.RegisterRequest{
		AgentId:   "agent-123",
		Hostname:  "prod-server-01",
		IpAddress: "192.168.1.100",
	})
	if err != nil {
		log.Fatalf("Could not register agent: %v", err)
	}
	log.Printf("Registration Response: Success=%t, Message='%s'", registerResp.GetSuccess(), registerResp.GetMessage())

	// --- 2. إرسال تقرير جدار الحماية ---
	log.Println("\n--- Reporting Firewall Status ---")
	firewallRules := []*pb.FirewallRule{
		{Name: "Allow SSH", Port: "22", Protocol: "TCP", Action: "ALLOW"},
		{Name: "Allow HTTP", Port: "80", Protocol: "TCP", Action: "ALLOW"},
		{Name: "Block All UDP", Port: "*", Protocol: "UDP", Action: "DENY"},
	}

	reportResp, err := c.ReportFirewallStatus(ctx, &pb.FirewallReportRequest{
		AgentId: "agent-123",
		Rules:   firewallRules,
	})
	if err != nil {
		log.Fatalf("Could not report firewall status: %v", err)
	}
	log.Printf("Firewall Report Response: Received=%t", reportResp.GetReceived())

	// --- 3. البحث عن العميل ---
	log.Println("\n--- Finding Agent ---")
	findResp, err := c.FindAgent(ctx, &pb.FindAgentRequest{AgentId: "agent-123"})
	if err != nil {
		log.Fatalf("Could not find agent: %v", err)
	}

	if findResp.GetFound() {
		log.Printf("Agent Found: ID=%s, Hostname=%s, IP=%s, LastSeen=%s",
			findResp.GetAgentId(), findResp.GetHostname(), findResp.GetIpAddress(), findResp.GetLastSeen())
	} else {
		log.Println("Agent with ID 'agent-123' was not found.")
	}
}
