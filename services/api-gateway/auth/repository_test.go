package auth

import (
	"strings"
	"testing"
)

func TestRefreshTokenQueriesUseHashAndRevocation(t *testing.T) {
	expectedStoreParts := []string{
		"INSERT INTO refresh_tokens",
		"(user_id, token_hash, expires_at)",
		"VALUES ($1, $2, $3)",
	}

	for _, part := range expectedStoreParts {
		if !strings.Contains(storeRefreshTokenQuery, part) {
			t.Fatalf("expected store query to contain %q, got %s", part, storeRefreshTokenQuery)
		}
	}

	expectedFindParts := []string{
		"FROM refresh_tokens",
		"WHERE token_hash = $1",
		"revoked_at",
		"expires_at",
	}

	for _, part := range expectedFindParts {
		if !strings.Contains(findRefreshTokenQuery, part) {
			t.Fatalf("expected find query to contain %q, got %s", part, findRefreshTokenQuery)
		}
	}

	expectedRevokeParts := []string{
		"UPDATE refresh_tokens",
		"SET revoked_at = NOW()",
		"WHERE token_hash = $1",
		"AND revoked_at IS NULL",
	}

	for _, part := range expectedRevokeParts {
		if !strings.Contains(revokeRefreshTokenQuery, part) {
			t.Fatalf("expected revoke query to contain %q, got %s", part, revokeRefreshTokenQuery)
		}
	}
}
