package client

import (
	"errors"
	"strings"
	"testing"

	"connectrpc.com/connect"

	"github.com/kodokbakar/pylon/internal/config"
)

func TestNewServiceClientNormalizesBaseURL(t *testing.T) {
	client, err := NewServiceClient("chat-service", "localhost:9001/")
	if err != nil {
		t.Fatalf("create service client: %v", err)
	}

	if client.Name != "chat-service" {
		t.Fatalf("expected chat-service, got %q", client.Name)
	}

	if client.BaseURL != "http://localhost:9001" {
		t.Fatalf("expected normalized url, got %q", client.BaseURL)
	}

	if client.HTTPClient == nil {
		t.Fatal("expected http client")
	}
}

func TestNewServiceClientRequiresName(t *testing.T) {
	_, err := NewServiceClient(" ", "http://localhost:9001")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "service name is required") {
		t.Fatalf("expected service name error, got %v", err)
	}
}

func TestNewServiceClientRequiresURL(t *testing.T) {
	_, err := NewServiceClient("chat-service", " ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "normalize chat-service url") {
		t.Fatalf("expected wrapped normalize error, got %v", err)
	}
}

func TestNewServiceClientRejectsInvalidURL(t *testing.T) {
	_, err := NewServiceClient("chat-service", "http://[::1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "parse url") {
		t.Fatalf("expected parse url error, got %v", err)
	}
}

func TestNewClientsBuildsAllServiceClients(t *testing.T) {
	clients, err := NewClients(config.ServicesConfig{
		ChatURL:         "chat:9001",
		PresenceURL:     "presence:9002",
		RoomURL:         "room:9003",
		NotificationURL: "notification:9004",
	})
	if err != nil {
		t.Fatalf("create clients: %v", err)
	}

	if clients.Chat.BaseURL != "http://chat:9001" {
		t.Fatalf("expected chat base url, got %q", clients.Chat.BaseURL)
	}

	if clients.Presence.BaseURL != "http://presence:9002" {
		t.Fatalf("expected presence base url, got %q", clients.Presence.BaseURL)
	}

	if clients.Room.BaseURL != "http://room:9003" {
		t.Fatalf("expected room base url, got %q", clients.Room.BaseURL)
	}

	if clients.Notification.BaseURL != "http://notification:9004" {
		t.Fatalf("expected notification base url, got %q", clients.Notification.BaseURL)
	}
}

func TestNewClientsReturnsFirstInvalidServiceURLError(t *testing.T) {
	_, err := NewClients(config.ServicesConfig{
		ChatURL:         "http://[::1",
		PresenceURL:     "presence:9002",
		RoomURL:         "room:9003",
		NotificationURL: "notification:9004",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "normalize chat-service url") {
		t.Fatalf("expected chat service url error, got %v", err)
	}
}

func TestUnavailableErrorWrapsCauseAsConnectUnavailable(t *testing.T) {
	cause := errors.New("dial tcp failed")
	serviceClient := &ServiceClient{Name: "room-service"}

	err := serviceClient.UnavailableError(cause)

	if connect.CodeOf(err) != connect.CodeUnavailable {
		t.Fatalf("expected unavailable code, got %v", connect.CodeOf(err))
	}

	if !strings.Contains(err.Error(), "room-service unavailable") {
		t.Fatalf("expected service name in error, got %v", err)
	}

	if !errors.Is(err, cause) {
		t.Fatalf("expected cause to be wrapped, got %v", err)
	}
}

func TestUnavailableErrorUsesDefaultCauseWhenNil(t *testing.T) {
	serviceClient := &ServiceClient{Name: "room-service"}

	err := serviceClient.UnavailableError(nil)

	if connect.CodeOf(err) != connect.CodeUnavailable {
		t.Fatalf("expected unavailable code, got %v", connect.CodeOf(err))
	}

	if !strings.Contains(err.Error(), "service unavailable") {
		t.Fatalf("expected default cause, got %v", err)
	}
}
