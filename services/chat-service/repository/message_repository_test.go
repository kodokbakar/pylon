package repository

import "testing"

func TestNewMessageRepositoryRequiresPostgresPool(t *testing.T) {
	_, err := NewMessageRepository(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
