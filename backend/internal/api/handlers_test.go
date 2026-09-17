package api

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestListIPOs(t *testing.T) {
	app := fiber.New()
	handler := NewIPOHandler(nil, nil, nil, nil)

	app.Get("/api/ipos", handler.ListIPOs)

	req := httptest.NewRequest("GET", "/api/ipos", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestManualSyncIPOs(t *testing.T) {
	app := fiber.New()
	handler := NewIPOHandler(nil, nil, nil, nil)

	app.Post("/api/ipos", handler.ManualSyncIPOs)

	req := httptest.NewRequest("POST", "/api/ipos", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
