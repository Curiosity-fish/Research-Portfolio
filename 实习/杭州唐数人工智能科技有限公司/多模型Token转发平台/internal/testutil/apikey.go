package testutil

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/group"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/internal/auth"
)

// CreateTestAPIKey creates a group, a user in that group, and an enabled API
// token for that user. It returns the created entities and the plaintext key.
func CreateTestAPIKey(t *testing.T, ctx context.Context, client *ent.Client) (*ent.User, *ent.Group, *ent.UserToken, string) {
	t.Helper()

	g, err := client.Group.Create().
		SetName("测试分组").
		SetCode("test-group-" + uuid.NewString()[:8]).
		SetStatus(group.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	u, err := client.User.Create().
		SetUsername("test-user-" + uuid.NewString()[:8]).
		SetPasswordHash("hash").
		SetName("测试用户").
		SetRole(user.RoleStudent).
		SetStatus(user.StatusActive).
		SetGroupID(g.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	plaintext, hash, err := auth.GenerateAPIKey()
	if err != nil {
		t.Fatalf("generate api key: %v", err)
	}

	tok, err := client.UserToken.Create().
		SetUserID(u.ID).
		SetName("default").
		SetTokenHash(hash).
		SetTokenLast4(plaintext[len(plaintext)-4:]).
		SetIsEnabled(true).
		Save(ctx)
	if err != nil {
		t.Fatalf("create user token: %v", err)
	}

	return u, g, tok, plaintext
}
