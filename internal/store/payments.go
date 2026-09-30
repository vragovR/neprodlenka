package store

import (
	"context"
	"fmt"
	"time"
)

type PaymentChange struct {
	ChildID int64
	Paid    bool
}

func (s *Store) MonthPaid(ctx context.Context, month time.Time) (map[int64]bool, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT child_id FROM payments WHERE month = ?`, month.Format(monthLayout))
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()

	paid := map[int64]bool{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}
		paid[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	return paid, nil
}
