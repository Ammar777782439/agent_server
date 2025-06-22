package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "agent_server/agent_server/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	serverAddr = "localhost:50051"
	agentID    = "agent-007"
)

func main() {
	log.Printf("Agent %s starting...", agentID)

	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewAgentServiceClient(conn)
	log.Printf("Connected to server at %s", serverAddr)

	for {
		log.Println("Attempting to start task channel...")
		startTaskChannel(client)
		log.Printf("Channel ended. Reconnecting in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
}

// الدالة تستخدم الآن العقد الصحيح TaskChannel
func startTaskChannel(client pb.AgentServiceClient) {
	stream, err := client.TaskChannel(context.Background())
	if err != nil {
		log.Printf("Failed to create task channel: %v", err)
		return
	}

	// Goroutine للاستماع للمهام القادمة من السيرفر
	waitc := make(chan struct{})
	go func() {
		for {
			// الآن نستقبل TaskRequest بشكل صحيح
			task, err := stream.Recv()
			if err == io.EOF {
				log.Println("Server closed the stream (EOF)")
				close(waitc)
				return
			}
			if err != nil {
				log.Printf("Error receiving task: %v", err)
				close(waitc)
				return
			}

			log.Printf("✅ Received new task from server [ID: %s, Command: %s]", task.TaskId, task.CommandName)
			go executeTaskAndSendResult(stream, task)
		}
	}()

	<-waitc
}

// الدالة الآن ترسل النتيجة (TaskResult) عبر القناة
func executeTaskAndSendResult(stream pb.AgentService_TaskChannelClient, task *pb.TaskRequest) {
	log.Printf("Executing task: %s", task.CommandName)
	time.Sleep(2 * time.Second) 

	result := &pb.TaskResult{
		TaskId:      task.TaskId,
		Success:     true,
		Payload:     `{"status": "executed", "details": "This is a real result for the correct design"}`,
	}
	
	// الآن نرسل TaskResult بشكل صحيح
	if err := stream.Send(result); err != nil {
		log.Printf("Failed to send result for task %s: %v", task.TaskId, err)
	}

	log.Printf("✔️ Sent result for task [ID: %s]", task.TaskId)
}