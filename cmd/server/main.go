// cmd/server/main.go

package main

import (
	"log"
	"net"

	pb "agent_server/agent_server/proto"
	"agent_server/internal/config"
	"agent_server/internal/model"

	"agent_server/internal/repository"
	"agent_server/internal/service"
	"agent_server/internal/usecase"
	"agent_server/internal/worker"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port                  = ":50051"
	configPath            = "config.yaml"
	
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

	
	// إضافة استعلام تجريبي لضمان وجود شيء لإرساله
	seedQuery := model.Query{
		Name:        "Get Basic System Information",
		Description: "Collects OS, CPU, and Memory info from an agent.",
		CommandName: "get_system_info",
	}
	// FirstOrCreate: يقوم بالإنشاء فقط إذا لم يجد استعلامًا بنفس الاسم
	db.Where("command_name = ?", "get_system_info").FirstOrCreate(&seedQuery)
	log.Println("✅ Seed query ensured in the database.")
	

	// 3. إنشاء المستودع (Repository)
	agentRepo := repository.NewAgentRepository(db)

	// 4. جديد: إنشاء طبقة منطق العمل (Use Case)
	agentLogic := usecase.NewAgentUseCase(agentRepo)

	// 5. إنشاء الخادم (Handler) مع حقن طبقة منطق العمل
	agentServer := service.NewAgentServer(agentLogic, db)
	monitor := worker.NewMonitor(agentLogic)

	go monitor.Start()
	// 6. بدء خادم gRPC
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAgentServiceServer(grpcServer, agentServer)
	
	reflection.Register(grpcServer)

	log.Printf("gRPC server listening on %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
