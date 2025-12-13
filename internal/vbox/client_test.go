package vbox

import (
	"errors"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	endpoint := "http://localhost:18083/"
	username := "testuser"
	password := "testpass"

	client := NewClient(endpoint, username, password)

	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.endpoint != endpoint {
		t.Errorf("expected endpoint %q, got %q", endpoint, client.endpoint)
	}
	if client.username != username {
		t.Errorf("expected username %q, got %q", username, client.username)
	}
	if client.password != password {
		t.Errorf("expected password %q, got %q", password, client.password)
	}
}

func TestIsNotFound_True(t *testing.T) {
	err := errNotFound
	if !IsNotFound(err) {
		t.Error("expected IsNotFound to return true for errNotFound")
	}
}

func TestIsNotFound_Wrapped(t *testing.T) {
	err := errors.New("some wrapper: " + errNotFound.Error())
	wrappedErr := errors.Join(errNotFound, err)
	if !IsNotFound(wrappedErr) {
		t.Error("expected IsNotFound to return true for wrapped errNotFound")
	}
}

func TestIsNotFound_False(t *testing.T) {
	err := errors.New("some other error")
	if IsNotFound(err) {
		t.Error("expected IsNotFound to return false for other errors")
	}
}

func TestIsNotFound_Nil(t *testing.T) {
	if IsNotFound(nil) {
		t.Error("expected IsNotFound to return false for nil error")
	}
}

func TestCloneRequest_Defaults(t *testing.T) {
	req := CloneRequest{
		Name:   "test-vm",
		Source: "source-vm",
	}

	// Verify the struct fields
	if req.Name != "test-vm" {
		t.Errorf("expected Name 'test-vm', got %q", req.Name)
	}
	if req.Source != "source-vm" {
		t.Errorf("expected Source 'source-vm', got %q", req.Source)
	}
	if req.CloneMode != "" {
		t.Errorf("expected empty CloneMode, got %q", req.CloneMode)
	}
	if req.CloneOptions != nil {
		t.Errorf("expected nil CloneOptions, got %v", req.CloneOptions)
	}
	if req.DesiredState != "" {
		t.Errorf("expected empty DesiredState, got %q", req.DesiredState)
	}
	if req.SessionType != "" {
		t.Errorf("expected empty SessionType, got %q", req.SessionType)
	}
	if req.Timeout != 0 {
		t.Errorf("expected zero Timeout, got %v", req.Timeout)
	}
}

func TestCloneRequest_WithOptions(t *testing.T) {
	req := CloneRequest{
		Name:         "test-vm",
		Source:       "source-vm",
		CloneMode:    "MachineState",
		CloneOptions: []string{"Link", "KeepAllMACs"},
		DesiredState: "started",
		SessionType:  "headless",
		Timeout:      30 * time.Minute,
	}

	if req.CloneMode != "MachineState" {
		t.Errorf("expected CloneMode 'MachineState', got %q", req.CloneMode)
	}
	if len(req.CloneOptions) != 2 {
		t.Errorf("expected 2 CloneOptions, got %d", len(req.CloneOptions))
	}
	if req.CloneOptions[0] != "Link" {
		t.Errorf("expected first CloneOption 'Link', got %q", req.CloneOptions[0])
	}
	if req.CloneOptions[1] != "KeepAllMACs" {
		t.Errorf("expected second CloneOption 'KeepAllMACs', got %q", req.CloneOptions[1])
	}
	if req.DesiredState != "started" {
		t.Errorf("expected DesiredState 'started', got %q", req.DesiredState)
	}
	if req.SessionType != "headless" {
		t.Errorf("expected SessionType 'headless', got %q", req.SessionType)
	}
	if req.Timeout != 30*time.Minute {
		t.Errorf("expected Timeout 30m, got %v", req.Timeout)
	}
}

// Integration test placeholder - requires a running VirtualBox webservice
// To run: go test -tags=integration ./...
// func TestClient_Integration(t *testing.T) {
//     if testing.Short() {
//         t.Skip("skipping integration test in short mode")
//     }
//
//     endpoint := os.Getenv("VBOX_ENDPOINT")
//     username := os.Getenv("VBOX_USERNAME")
//     password := os.Getenv("VBOX_PASSWORD")
//
//     if endpoint == "" || username == "" || password == "" {
//         t.Skip("VBOX_ENDPOINT, VBOX_USERNAME, and VBOX_PASSWORD must be set for integration tests")
//     }
//
//     client := NewClient(endpoint, username, password)
//     // ... integration tests
// }
