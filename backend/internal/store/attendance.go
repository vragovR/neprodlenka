package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type MarkChange struct {
	Date    time.Time
	ChildID int64
	Status  Status
}

func (s *Store) SaveMarks(ctx context.Context, month time.Time, changes []MarkChange, payments []PaymentChange) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("save marks: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, c := range changes {
		d := c.Date.Format(time.DateOnly)
		if err := childExists(ctx, tx, c.ChildID); err != nil {
			return fmt.Errorf("save marks: %w", err)
		}
		if c.Status == "" {
			if _, err := tx.ExecContext(ctx, `DELETE FROM attendance WHERE date = ? AND child_id = ?`, d, c.ChildID); err != nil {
				return fmt.Errorf("save marks: %w", err)
			}
			continue
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO attendance (date, child_id, status, price)
			SELECT ?1, id, ?2, CASE WHEN ?2 = 'present' THEN price ELSE 0 END FROM children WHERE id = ?3
			ON CONFLICT (date, child_id) DO UPDATE SET status = excluded.status, price = excluded.price
			WHERE attendance.status <> excluded.status`,
			d, c.Status, c.ChildID)
		if err != nil {
			return fmt.Errorf("save marks: %w", err)
		}
	}

	m := month.Format(monthLayout)
	for _, p := range payments {
		if err := childExists(ctx, tx, p.ChildID); err != nil {
			return fmt.Errorf("save marks: %w", err)
		}
		query := `DELETE FROM payments WHERE month = ? AND child_id = ?`
		if p.Paid {
			query = `INSERT INTO payments (month, child_id) VALUES (?, ?) ON CONFLICT DO NOTHING`
		}
		if _, err := tx.ExecContext(ctx, query, m, p.ChildID); err != nil {
			return fmt.Errorf("save marks: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("save marks: %w", err)
	}
	return nil
}

func childExists(ctx context.Context, tx *sql.Tx, id int64) error {
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM children WHERE id = ?)`, id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%w: child %d", ErrNotFound, id)
	}
	return nil
}
