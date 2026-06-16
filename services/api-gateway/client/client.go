package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/kodokbakar/pylon/internal/config"
)

type Clients struct {
	Chat         *ServiceClient
	Presence     *ServiceClient
	Room         *ServiceClient
	Notification *ServiceClient
}

type ServiceClient struct {
	Name       string
	BaseURL    string
	HTTPClient *http.Client
}

func NewClients(cfg config.ServicesConfig) (*Clients, error) {
	chatClient, err := NewServiceClient("chat-service", cfg.ChatURL)
	if err != nil {
		return nil, err
	}

	presenceClient, err := NewServiceClient("presence-service", cfg.PresenceURL)
	if err != nil {
		return nil, err
	}

	roomClient, err := NewServiceClient("room-service", cfg.RoomURL)
	if err != nil {
		return nil, err
	}

	notificationClient, err := NewServiceClient("notification-service", cfg.NotificationURL)
	if err != nil {
		return nil, err
	}

	return &Clients{
		Chat:         chatClient,
		Presence:     presenceClient,
		Room:         roomClient,
		Notification: notificationClient,
	}, nil
}

func NewServiceClient(name, rawURL string) (*ServiceClient, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("service name is required")
	}

	baseURL, err := normalizeBaseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("normalize %s url: %w", name, err)
	}

	return &ServiceClient{
		Name:    name,
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (c *ServiceClient) UnavailableError(err error) error {
	if err == nil {
		err = fmt.Errorf("service unavailable")
	}

	return connect.NewError(connect.CodeUnavailable, fmt.Errorf("%s unavailable: %w", c.Name, err))
}

func normalizeBaseURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", fmt.Errorf("url is required")
	}

	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", fmt.Errorf("url must include scheme and host")
	}

	return strings.TrimRight(parsedURL.String(), "/"), nil
}
