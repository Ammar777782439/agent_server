// internal/worker/monitor.go

package worker

import (
	"agent_server/internal/usecase" 
	"log"
	"time"
)

// Monitor هو الهيكل الذي يمثل تغير حاله الوكيل 
type Monitor struct {
	agentLogic usecase.AgentUseCase // يعتمد على  (منطق العمل) ليقوم بالفعل
}

// NewMonitor هو المُصنِّع لتغير الحاله الجديد
func NewMonitor(logic usecase.AgentUseCase) *Monitor {
	return &Monitor{agentLogic: logic}
}

// Start يبدأ عملية المراقبة الدورية في الخلفية
func (m *Monitor) Start() {
	log.Println("✅ Starting offline agent monitor...")

	
	ticker := time.NewTicker(1 * time.Minute) 
	defer ticker.Stop()

	
	for range ticker.C {
		log.Println("Inspector at work: Checking for offline agents...")

		
		err := m.agentLogic.MarkOfflineAgents()
		if err != nil {
			log.Printf("❌ Error during offline agent check: %v", err)
		}
	}





}