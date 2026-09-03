package postgres

import (
	"context"
	"fmt"

	"github.com/clubmax/clubmax/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type migration struct {
	version int
	name    string
	sql     string
}

var migrations = []migration{
	{version: 1, name: "initial_schema", sql: domain.Schema},
	{version: 2, name: "support_tickets", sql: `
CREATE TABLE IF NOT EXISTS support_tickets (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), number bigint GENERATED ALWAYS AS IDENTITY UNIQUE,
 owner_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 category varchar(20) NOT NULL CHECK(category IN ('payment','account','gym','technical','other')),
 subject varchar(200) NOT NULL, status varchar(20) NOT NULL CHECK(status IN ('waiting_admin','waiting_member','closed')),
 assigned_admin_id uuid REFERENCES users(id) ON DELETE SET NULL,
 last_message_at timestamptz NOT NULL DEFAULT now(), member_last_read_at timestamptz NOT NULL DEFAULT now(),
 admin_last_read_at timestamptz, closed_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS support_tickets_owner_updated ON support_tickets(owner_id,last_message_at DESC);
CREATE INDEX IF NOT EXISTS support_tickets_admin_queue ON support_tickets(status,last_message_at DESC);
CREATE TABLE IF NOT EXISTS support_ticket_messages (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), ticket_id uuid NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
 sender_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE, body text NOT NULL CHECK(length(body) BETWEEN 1 AND 5000),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS support_ticket_messages_thread ON support_ticket_messages(ticket_id,created_at,id);`},
	// Some pre-migration installations already have the ticket tables with this
	// obsolete column. IF EXISTS keeps this migration safe for fresh databases.
	{version: 3, name: "remove_ticket_priority", sql: `ALTER TABLE support_tickets DROP COLUMN IF EXISTS priority;`},
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version integer PRIMARY KEY,
		name text NOT NULL,
		applied_at timestamptz NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("create migration history: %w", err)
	}
	for _, item := range migrations {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", item.version, err)
		}
		if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(434652829)`); err == nil {
			var applied bool
			err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, item.version).Scan(&applied)
			if err == nil && !applied {
				_, err = tx.Exec(ctx, item.sql)
			}
			if err == nil && !applied {
				_, err = tx.Exec(ctx, `INSERT INTO schema_migrations(version,name) VALUES($1,$2)`, item.version, item.name)
			}
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %d (%s): %w", item.version, item.name, err)
		}
		if err = tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %d (%s): %w", item.version, item.name, err)
		}
	}
	return nil
}
