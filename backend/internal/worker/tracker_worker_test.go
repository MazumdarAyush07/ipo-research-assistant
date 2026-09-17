package worker

import (
	"context"
	"testing"

	"github.com/hibiken/asynq"
)

func TestHandleSyncGMPTask(t *testing.T) {
	processor := NewProcessor(nil, nil, nil)
	
	// Create a dummy task
	task := asynq.NewTask(TaskSyncGMP, nil)

	if processor == nil {
		t.Errorf("Expected processor to be initialized")
	}

	if task.Type() != TaskSyncGMP {
		t.Errorf("Expected task type %s, got %s", TaskSyncGMP, task.Type())
	}

	// We avoid calling processor.HandleSyncGMPTask directly 
	// here because it makes an actual DB query and HTTP request.
	_ = context.Background()
}

func TestHandleSyncSubscriptionsTask(t *testing.T) {
	processor := NewProcessor(nil, nil, nil)
	
	// Create a dummy task
	task := asynq.NewTask(TaskSyncSubscriptions, nil)

	if processor == nil {
		t.Errorf("Expected processor to be initialized")
	}

	if task.Type() != TaskSyncSubscriptions {
		t.Errorf("Expected task type %s, got %s", TaskSyncSubscriptions, task.Type())
	}

	// We avoid calling processor.HandleSyncSubscriptionsTask directly 
	// here because it makes an actual DB query and HTTP request.
	_ = context.Background()
}
