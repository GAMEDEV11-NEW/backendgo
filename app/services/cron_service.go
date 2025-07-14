package services

import (
	"log"
	"time"

	"sync"

	"github.com/gocql/gocql"
)

// CronService handles scheduled background tasks
type CronService struct {
	matchmakingService *MatchmakingService
	stopChan           chan bool
	isRunning          bool
	pendingRun         bool
	pendingRunMu       sync.Mutex
}

// NewCronService creates a new cron service instance
func NewCronService(cassandraSession *gocql.Session) *CronService {
	return &CronService{
		matchmakingService: NewMatchmakingService(cassandraSession),
		stopChan:           make(chan bool),
		isRunning:          false,
		pendingRun:         false,
		pendingRunMu:       sync.Mutex{},
	}
}

// StartMatchmakingCron starts the matchmaking cron job
func (c *CronService) StartMatchmakingCron(interval time.Duration) {
	if c.isRunning {
		log.Println("⚠️ Matchmaking cron is already running")
		return
	}

	c.isRunning = true
	log.Printf("🚀 Starting matchmaking cron job (interval: %v)", interval)

	go func() {
		for {
			// 1. Run matchmaking
			c.runMatchmaking()

			// 2. Check if another run was requested during the last run
			c.pendingRunMu.Lock()
			rerun := c.pendingRun
			c.pendingRun = false
			c.pendingRunMu.Unlock()

			if rerun {
				// Run again immediately (do-while style)
				continue
			}

			// 3. Otherwise, wait for the interval
			select {
			case <-c.stopChan:
				log.Println("🛑 Stopping matchmaking cron job")
				return
			case <-time.After(interval):
				// Loop continues
			}
		}
	}()
}

// StopMatchmakingCron stops the matchmaking cron job
func (c *CronService) StopMatchmakingCron() {
	if !c.isRunning {
		log.Println("⚠️ Matchmaking cron is not running")
		return
	}

	c.isRunning = false
	c.stopChan <- true
	log.Println("🛑 Matchmaking cron job stopped")
}

// runMatchmaking executes the matchmaking process
func (c *CronService) runMatchmaking() {
	startTime := time.Now()
	log.Println("🔄 Running scheduled matchmaking process...")

	// Process matchmaking
	err := c.matchmakingService.ProcessMatchmaking()
	if err != nil {
		log.Printf("❌ Matchmaking process failed: %v", err)
		return
	}

	// Get and log statistics
	stats, err := c.matchmakingService.GetMatchmakingStats()
	if err != nil {
		log.Printf("❌ Failed to get matchmaking stats: %v", err)
		return
	}

	duration := time.Since(startTime)
	log.Printf("✅ Matchmaking completed in %v", duration)
	log.Printf("📊 Stats: %+v", stats)
}

// RunCleanupCron starts the cleanup cron job
func (c *CronService) RunCleanupCron(interval time.Duration, maxAge time.Duration) {
	log.Printf("🧹 Starting cleanup cron job (interval: %v, max age: %v)", interval, maxAge)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.runCleanup(maxAge)
			case <-c.stopChan:
				log.Println("🛑 Stopping cleanup cron job")
				return
			}
		}
	}()
}

// runCleanup executes the cleanup process
func (c *CronService) runCleanup(maxAge time.Duration) {
	log.Println("🧹 Running scheduled cleanup process...")

	err := c.matchmakingService.CleanupExpiredMatches(maxAge)
	if err != nil {
		log.Printf("❌ Cleanup process failed: %v", err)
		return
	}

	log.Println("✅ Cleanup completed")
}

// IsRunning returns whether the cron service is currently running
func (c *CronService) IsRunning() bool {
	return c.isRunning
}

// GetMatchmakingStats returns current matchmaking statistics
func (c *CronService) GetMatchmakingStats() (map[string]interface{}, error) {
	return c.matchmakingService.GetMatchmakingStats()
}

// RequestMatchmakingRun sets the pendingRun flag to true, so the next loop will run again immediately
func (c *CronService) RequestMatchmakingRun() {
	c.pendingRunMu.Lock()
	c.pendingRun = true
	c.pendingRunMu.Unlock()
}
