package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/notification"
	"github.com/school-api/school-api-v1/internal/testutil"
)

func TestEntNotificationRepository_CreateAndList(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	repo := NewEntNotificationRepository(client)

	broadcast, err := repo.Create(ctx, CreateNotificationInput{
		Type:    string(notification.TypeAnnouncement),
		Title:   "系统公告",
		Content: "全体用户可见",
	})
	if err != nil {
		t.Fatalf("create broadcast: %v", err)
	}
	if broadcast.UserID != nil {
		t.Fatalf("expected broadcast user_id nil, got %v", broadcast.UserID)
	}

	personal, err := repo.Create(ctx, CreateNotificationInput{
		UserID:  &u.ID,
		Type:    string(notification.TypeSystem),
		Title:   "个人通知",
		Content: "仅该用户可见",
	})
	if err != nil {
		t.Fatalf("create personal: %v", err)
	}
	if personal.UserID == nil || *personal.UserID != u.ID {
		t.Fatalf("expected personal user_id %s, got %v", u.ID, personal.UserID)
	}

	list, total, err := repo.ListByUserID(ctx, u.ID, 0, 10)
	if err != nil {
		t.Fatalf("list by user: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}
}

func TestEntNotificationRepository_ListAdminFilter(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	repo := NewEntNotificationRepository(client)

	_, err := repo.Create(ctx, CreateNotificationInput{
		UserID:  &u.ID,
		Type:    string(notification.TypeSystem),
		Title:   "个人通知",
		Content: "仅该用户可见",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	list, total, err := repo.List(ctx, ListNotificationFilter{UserID: &u.ID, Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("list admin: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 item, got total=%d len=%d", total, len(list))
	}

	typ := string(notification.TypeAnnouncement)
	_, total, err = repo.List(ctx, ListNotificationFilter{Type: &typ, Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("list admin by type: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected 0 items, got %d", total)
	}
}

func TestEntNotificationRepository_MarkRead(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	other, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	repo := NewEntNotificationRepository(client)

	n, err := repo.Create(ctx, CreateNotificationInput{
		UserID:  &u.ID,
		Type:    string(notification.TypeSystem),
		Title:   "个人通知",
		Content: "仅该用户可见",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	ok, err := repo.MarkRead(ctx, n.ID, other.ID)
	if err != nil {
		t.Fatalf("mark read other: %v", err)
	}
	if ok {
		t.Fatal("expected mark read to fail for other user")
	}

	ok, err = repo.MarkRead(ctx, n.ID, u.ID)
	if err != nil {
		t.Fatalf("mark read owner: %v", err)
	}
	if !ok {
		t.Fatal("expected mark read to succeed for owner")
	}

	fresh, err := repo.GetByID(ctx, n.ID)
	if err != nil {
		t.Fatalf("get after mark: %v", err)
	}
	if !fresh.IsRead {
		t.Fatal("expected is_read true")
	}
}

func TestEntNotificationRepository_MarkReadBroadcast(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	other, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	repo := NewEntNotificationRepository(client)

	n, err := repo.Create(ctx, CreateNotificationInput{
		Type:    string(notification.TypeAnnouncement),
		Title:   "广播",
		Content: "所有人可读",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	ok, err := repo.MarkRead(ctx, n.ID, u.ID)
	if err != nil {
		t.Fatalf("mark read broadcast: %v", err)
	}
	if !ok {
		t.Fatal("expected mark read broadcast to succeed")
	}

	// The reader sees the broadcast as read.
	list, _, err := repo.ListByUserID(ctx, u.ID, 0, 10)
	if err != nil {
		t.Fatalf("list reader: %v", err)
	}
	if len(list) != 1 || !list[0].IsRead {
		t.Fatalf("expected reader to see broadcast as read, got %+v", list)
	}

	// Other users still see it as unread, and the notification row itself
	// must stay unread because read state is tracked per user.
	list, _, err = repo.ListByUserID(ctx, other.ID, 0, 10)
	if err != nil {
		t.Fatalf("list other: %v", err)
	}
	if len(list) != 1 || list[0].IsRead {
		t.Fatalf("expected other user to see broadcast as unread, got %+v", list)
	}

	fresh, err := repo.GetByID(ctx, n.ID)
	if err != nil {
		t.Fatalf("get broadcast: %v", err)
	}
	if fresh.IsRead {
		t.Fatal("expected broadcast notification row to remain unread")
	}

	// Marking the same broadcast again is a no-op, not an error.
	ok, err = repo.MarkRead(ctx, n.ID, u.ID)
	if err != nil {
		t.Fatalf("mark read broadcast again: %v", err)
	}
	if !ok {
		t.Fatal("expected repeat mark read to succeed")
	}
}

func TestEntNotificationRepository_ListByUserIDPagination(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	repo := NewEntNotificationRepository(client)

	for i := 0; i < 3; i++ {
		if _, err := repo.Create(ctx, CreateNotificationInput{
			UserID:  &u.ID,
			Type:    string(notification.TypeSystem),
			Title:   "个人通知",
			Content: "仅该用户可见",
		}); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	// Page 2 must not fail on the count query.
	list, total, err := repo.ListByUserID(ctx, u.ID, 2, 2)
	if err != nil {
		t.Fatalf("list page 2: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected total 3, got %d", total)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 item on page 2, got %d", len(list))
	}
}

func TestEntNotificationRepository_ListAdminPagination(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	repo := NewEntNotificationRepository(client)
	for i := 0; i < 3; i++ {
		if _, err := repo.Create(ctx, CreateNotificationInput{
			Type:    string(notification.TypeAnnouncement),
			Title:   "公告",
			Content: "内容",
		}); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	list, total, err := repo.List(ctx, ListNotificationFilter{Offset: 2, Limit: 2})
	if err != nil {
		t.Fatalf("list page 2: %v", err)
	}
	if total != 3 || len(list) != 1 {
		t.Fatalf("expected total=3 len=1, got total=%d len=%d", total, len(list))
	}
}

func TestEntNotificationRepository_GetByIDNotFound(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	repo := NewEntNotificationRepository(client)
	_, err := repo.GetByID(ctx, uuid.New())
	if err == nil {
		t.Fatal("expected error for missing notification")
	}
}

func TestEntNotificationRepository_UnreadCountAndReadAll(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	other, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	repo := NewEntNotificationRepository(client)

	// One broadcast + one personal (unread), for two different users.
	broadcast, err := repo.Create(ctx, CreateNotificationInput{
		Type:    string(notification.TypeAnnouncement),
		Title:   "公告A",
		Content: "内容",
	})
	if err != nil {
		t.Fatalf("create broadcast: %v", err)
	}
	if _, err := repo.Create(ctx, CreateNotificationInput{
		UserID:  &u.ID,
		Type:    string(notification.TypeSystem),
		Title:   "个人A",
		Content: "内容",
	}); err != nil {
		t.Fatalf("create personal: %v", err)
	}

	count, err := repo.UnreadCount(ctx, u.ID)
	if err != nil {
		t.Fatalf("unread count: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected unread 2 (1 personal + 1 broadcast), got %d", count)
	}

	// Other user sees only the broadcast.
	count, err = repo.UnreadCount(ctx, other.ID)
	if err != nil {
		t.Fatalf("unread count other: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected unread 1 for other user, got %d", count)
	}

	// u reads only the broadcast; personal must stay unread.
	if ok, err := repo.MarkRead(ctx, broadcast.ID, u.ID); err != nil || !ok {
		t.Fatalf("mark broadcast read: ok=%v err=%v", ok, err)
	}
	count, err = repo.UnreadCount(ctx, u.ID)
	if err != nil {
		t.Fatalf("unread count after broadcast read: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected unread 1 after broadcast read, got %d", count)
	}

	// Read-all clears everything and is idempotent.
	if err := repo.ReadAll(ctx, u.ID); err != nil {
		t.Fatalf("read all: %v", err)
	}
	if err := repo.ReadAll(ctx, u.ID); err != nil {
		t.Fatalf("read all second time: %v", err)
	}
	count, err = repo.UnreadCount(ctx, u.ID)
	if err != nil {
		t.Fatalf("unread count after read all: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected unread 0 after read-all, got %d", count)
	}

	// Other user's broadcast state is untouched.
	count, err = repo.UnreadCount(ctx, other.ID)
	if err != nil {
		t.Fatalf("unread count other final: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected other user unread 1, got %d", count)
	}
}
