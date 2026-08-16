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

	// 1. Sync IPOs (Daily at 1:30 AM UTC / 7:00 AM IST)
	taskSync := asynq.NewTask(TaskSyncIPOs, nil)
	entrySync, err := scheduler.Register("30 1 * * *", taskSync)
	if err != nil {
		return nil, err
	}
	log.Printf("Registered Cron Job for Syncing IPOs. EntryID: %s", entrySync)

	// 2. Sync GMP (Every 4 hours)
	taskGMP := asynq.NewTask(TaskSyncGMP, nil)
	entryGMP, err := scheduler.Register("0 */4 * * *", taskGMP)
	if err != nil {
		return nil, err
	}
	log.Printf("Registered Cron Job for Syncing GMP. EntryID: %s", entryGMP)

	// 3. Sync Subscriptions (Daily at 1:00 PM UTC / 6:30 PM IST)
	taskSub := asynq.NewTask(TaskSyncSubscriptions, nil)
	entrySub, err := scheduler.Register("0 13 * * *", taskSub)
	if err != nil {
		return nil, err
	}
	log.Printf("Registered Cron Job for Syncing Subscriptions. EntryID: %s", entrySub)

	return scheduler, nil
}
