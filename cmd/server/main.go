package main

import (
	"log"
	"net"

	"agent_server/internal/config"
	"agent_server/internal/repository"
	"agent_server/internal/service"
	pb "agent_server/agent_server/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
	configPath = "config.yaml"
)

func main() {
	// 1. تحميل الإعدادات
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", configPath, err)
	}

	// 2. الاتصال بقاعدة البيانات
	db, err := repository.ConnectDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection successful.")

	// 3. إنشاء المستودع (Repository)
	agentRepo := repository.NewAgentRepository(db)

	// 4. إنشاء الخادم (Server/Handler) مع حقن المستودع
	agentServer := service.NewAgentServer(agentRepo)

	// 5. بدء خادم gRPC
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAgentServiceServer(grpcServer, agentServer)
	reflection.Register(grpcServer) // تفعيل الانعكاس لتسهيل الاختبار

	log.Printf("gRPC server listening on %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}