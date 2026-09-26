package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Mark struct {
	ChildID int64
	Date    string
	Status  Status
	Price   int64
}

func (s *Store) MonthByReportUID(ctx context.Context, uid string) (time.Time, bool, error) {
	var month string
	err := s.db.QueryRowContext(ctx, `SELECT month FROM report_months WHERE uid = ?`, uid).Scan(&month)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, fmt.Errorf("get report month: %w", err)
	}
	m, err := time.ParseInLocation(monthLayout, month, time.Local)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("parse report month %q: %w", month, err)
	}
	return m, true, nil
}

func (s *Store) MonthMarks(ctx context.Context, month time.Time) ([]Mark, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT child_id, date, status, price FROM attendance WHERE date LIKE ? ORDER BY date`,
		month.Format(monthLayout)+"-%",
	)
	if err != nil {
		return nil, fmt.Errorf("list month marks: %w", err)
	}
	defer rows.Close()

	var marks []Mark
	for rows.Next() {
		var m Mark
		if err := rows.Scan(&m.ChildID, &m.Date, &m.Status, &m.Price); err != nil {
			return nil, fmt.Errorf("scan mark: %w", err)
		}
		marks = append(marks, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list month marks: %w", err)
	}
	return marks, nil
}
