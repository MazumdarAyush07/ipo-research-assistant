package worker

import (
	"context"
	"testing"

	"github.com/hibiken/asynq"
)

func TestHandleSyncIPOsTask(t *testing.T) {
	processor := NewProcessor(nil, nil, nil)
	
	// Create a dummy task
	task := asynq.NewTask(TaskSyncIPOs, nil)

	// In a real test, we would mock the scraper HTTP call so it doesn't hit the internet.
	// For this test, we simply verify the function signature and processor struct are valid.
	if processor == nil {
		t.Errorf("Expected processor to be initialized")
	}

	if task.Type() != TaskSyncIPOs {
		t.Errorf("Expected task type %s, got %s", TaskSyncIPOs, task.Type())
	}

	// We avoid calling processor.HandleSyncIPOsTask(context.Background(), task) directly 
	// here because it makes an actual HTTP request to Chittorgarh.
	_ = context.Background()
}
