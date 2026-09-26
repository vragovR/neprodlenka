package store

import (
	"context"
	"fmt"
	"time"
)

type MarkChange struct {
	Date    time.Time
	ChildID int64
	Status  Status
}

func (s *Store) SaveMarks(ctx context.Context, changes []MarkChange) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("save marks: %w", err)
	}
	defer tx.Rollback()

	for _, c := range changes {
		d := c.Date.Format(time.DateOnly)
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM children WHERE id = ?)`, c.ChildID).Scan(&exists); err != nil {
			return fmt.Errorf("save marks: %w", err)
		}
		if !exists {
			return fmt.Errorf("%w: child %d", ErrNotFound, c.ChildID)
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("save marks: %w", err)
	}
	return nil
}
