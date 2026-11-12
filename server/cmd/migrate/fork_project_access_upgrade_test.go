package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// Execute the same data bridge and rollback fixture used for deployment
// rehearsals. Its temporary tables shadow public tables on this connection.
func TestForkProjectAccessBridgeAndRollback(t *testing.T) {
	t.Parallel()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("integration test requires Postgres at DATABASE_URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect to fixture database: %v", err)
	}
	defer conn.Close(context.Background())
	fixturePath := "../../internal/migrations/testdata/fork_project_access_to_official.sql"
	fixture, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	var sql strings.Builder
	for _, line := range strings.Split(string(fixture), "\n") {
		switch {
		case line == `\set ON_ERROR_STOP on`:
			continue // pgx Exec returns the first SQL error.
		case strings.HasPrefix(line, `\ir `):
			name := strings.TrimPrefix(line, `\ir `)
			if name != "../../../migrations/642_migrate_fork_project_access.up.sql" && name != "../../../migrations/642_migrate_fork_project_access.down.sql" {
				t.Fatalf("unexpected fixture include: %s", name)
			}
			migration, readErr := os.ReadFile(filepath.Join(filepath.Dir(fixturePath), name))
			if readErr != nil {
				t.Fatal(readErr)
			}
			sql.Write(migration)
		case strings.HasPrefix(line, `\`):
			t.Fatalf("unsupported fixture directive: %s", line)
		default:
			sql.WriteString(line)
		}
		sql.WriteByte('\n')
	}
	if _, err := conn.Exec(ctx, sql.String(), pgx.QueryExecModeSimpleProtocol); err != nil {
		t.Fatalf("project access migration/rollback fixture: %v", err)
	}
}
