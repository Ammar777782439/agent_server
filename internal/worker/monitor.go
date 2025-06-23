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

	// هذا التيكر هو الساعة التي ينظر إليها المفتش ليبدأ جولته
	ticker := time.NewTicker(1 * time.Minute) // افحص كل دقيقة
	defer ticker.Stop()

	// حلقة لا نهائية تمثل جولات المفتش الدورية
	for range ticker.C {
		log.Println("Inspector at work: Checking for offline agents...")

		//  لا يقوم بالفعل بنفسه، بل يطلب منطق
		err := m.agentLogic.MarkOfflineAgents()
		if err != nil {
			log.Printf("❌ Error during offline agent check: %v", err)
		}
	}
}