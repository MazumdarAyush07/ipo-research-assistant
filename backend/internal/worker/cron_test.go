package worker

import (
	"testing"
)

func TestInitCronScheduler(t *testing.T) {
	// Simple test to ensure the task constant is correct
	if TaskSyncIPOs != "ipo:sync" {
		t.Errorf("Expected TaskSyncIPOs to be 'ipo:sync', got %s", TaskSyncIPOs)
	}

	// Note: We avoid instantiating the full scheduler here without a mock redis server, 
	// as it connects to Redis immediately upon scheduling. In a full suite we'd use miniredis.
}
