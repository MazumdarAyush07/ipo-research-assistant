package worker

import (
	"log"

	"github.com/hibiken/asynq"
)

const (
	TaskSyncIPOs = "ipo:sync"
)

// InitCronScheduler sets up the daily cron job using asynq
func InitCronScheduler(redisAddr string) (*asynq.Scheduler, error) {
	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: redisAddr},
		&asynq.SchedulerOpts{
			LogLevel: asynq.InfoLevel,
		},
	)

	task := asynq.NewTask(TaskSyncIPOs, nil)

	// Run every day at 7:00 AM IST (1:30 AM UTC)
	// Example cron expression: "30 1 * * *"
	entryID, err := scheduler.Register("30 1 * * *", task)
	if err != nil {
		return nil, err
	}

	log.Printf("Registered Cron Job for Syncing IPOs. EntryID: %s", entryID)

	return scheduler, nil
}
