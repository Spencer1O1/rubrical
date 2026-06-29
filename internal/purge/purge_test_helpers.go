package purge

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"rubrical/internal/auth"
	"rubrical/internal/db"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	database, err := db.Connect(context.Background(), "postgres://rubrical:rubrical@localhost:5432/rubrical?sslmode=disable")
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	t.Cleanup(database.Close)
	return database.Pool
}

func testUserID(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()
	svc := auth.NewService(pool, auth.DefaultSessionTTL)
	email := fmt.Sprintf("purge-%d@rubrical.dev", time.Now().UnixNano())
	user, err := svc.CreateUserWithPassword(context.Background(), email, "password123", "Purge Test")
	if err != nil {
		t.Fatal(err)
	}
	return user.ID
}
