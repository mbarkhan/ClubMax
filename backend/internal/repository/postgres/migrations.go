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
	{version: 4, name: "payment_orders", sql: `
CREATE TABLE IF NOT EXISTS payment_orders (
 id uuid PRIMARY KEY, athlete_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 gym_id uuid NOT NULL REFERENCES gyms(id) ON DELETE CASCADE,
 kind varchar(20) NOT NULL CHECK(kind IN ('registration','monthly')),
 tuition_amount bigint NOT NULL CHECK(tuition_amount>=0), platform_fee bigint NOT NULL CHECK(platform_fee>=0),
 total_amount bigint NOT NULL CHECK(total_amount=tuition_amount+platform_fee),
 billing_year smallint NOT NULL CHECK(billing_year BETWEEN 1300 AND 1600),
 billing_month smallint NOT NULL CHECK(billing_month BETWEEN 1 AND 12),
 status varchar(20) NOT NULL CHECK(status IN ('created','pending','succeeded','failed')),
 zibal_track_id bigint UNIQUE, zibal_ref_number varchar(100),
 failure_reason text, expires_at timestamptz NOT NULL, completed_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
 CONSTRAINT payment_orders_gateway_status_check CHECK(status NOT IN ('pending','succeeded') OR zibal_track_id IS NOT NULL),
 CHECK((status='succeeded') = (completed_at IS NOT NULL))
);
CREATE INDEX IF NOT EXISTS payment_orders_athlete_gym_active ON payment_orders(athlete_id,gym_id,created_at DESC) WHERE status IN ('created','pending');
CREATE UNIQUE INDEX IF NOT EXISTS payment_orders_zibal_ref_unique ON payment_orders(zibal_ref_number) WHERE zibal_ref_number IS NOT NULL;
CREATE TABLE IF NOT EXISTS payment_order_months (
 order_id uuid NOT NULL REFERENCES payment_orders(id) ON DELETE CASCADE,
 billing_year smallint NOT NULL CHECK(billing_year BETWEEN 1300 AND 1600),
 billing_month smallint NOT NULL CHECK(billing_month BETWEEN 1 AND 12),
 amount bigint NOT NULL CHECK(amount>=0), PRIMARY KEY(order_id,billing_year,billing_month)
);
ALTER TABLE tuition_payments ADD COLUMN IF NOT EXISTS payment_order_id uuid REFERENCES payment_orders(id) ON DELETE RESTRICT;
CREATE INDEX IF NOT EXISTS tuition_payments_order_id ON tuition_payments(payment_order_id);`},
	// This is intentionally separate so installations that briefly ran migration
	// 4 before the security review also stop retaining card data.
	{version: 5, name: "harden_payment_orders", sql: `
ALTER TABLE payment_orders DROP COLUMN IF EXISTS zibal_card_number;
CREATE UNIQUE INDEX IF NOT EXISTS payment_orders_zibal_ref_unique ON payment_orders(zibal_ref_number) WHERE zibal_ref_number IS NOT NULL;
ALTER TABLE tuition_payments ADD COLUMN IF NOT EXISTS payment_order_id uuid REFERENCES payment_orders(id) ON DELETE RESTRICT;
CREATE INDEX IF NOT EXISTS tuition_payments_order_id ON tuition_payments(payment_order_id);`},
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
