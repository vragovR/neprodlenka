package store

import (
	"context"
	"fmt"
	"time"
)

type DayClub struct {
	ChildID   int64
	LastName  string
	FirstName string
	Start     string
	Name      string
}

func (s *Store) DayClubs(ctx context.Context, date time.Time) ([]DayClub, error) {
	return s.dayClubs(ctx, `
		SELECT dc.child_id, c.last_name, c.first_name, dc.start, dc.name
		FROM day_clubs dc JOIN children c ON c.id = dc.child_id
		WHERE dc.date = ?
		ORDER BY dc.start, dc.name`,
		date.Format(time.DateOnly),
	)
}

func (s *Store) ScheduleClubs(ctx context.Context, date time.Time) ([]DayClub, error) {
	return s.dayClubs(ctx, `
		SELECT cc.child_id, c.last_name, c.first_name, cc.start, cc.name
		FROM child_clubs cc JOIN children c ON c.id = cc.child_id
		WHERE c.active = 1 AND cc.weekday = ?
		ORDER BY cc.start, cc.name`,
		Weekday(date),
	)
}

func (s *Store) dayClubs(ctx context.Context, query string, arg any) ([]DayClub, error) {
	rows, err := s.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("list day clubs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	clubs := []DayClub{}
	for rows.Next() {
		var c DayClub
		if err := rows.Scan(&c.ChildID, &c.LastName, &c.FirstName, &c.Start, &c.Name); err != nil {
			return nil, fmt.Errorf("scan day club: %w", err)
		}
		clubs = append(clubs, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list day clubs: %w", err)
	}
	return clubs, nil
}

func (s *Store) SetDayClubs(ctx context.Context, date time.Time, childID int64, clubs []DayClub) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("set day clubs: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM children WHERE id = ?)`, childID).Scan(&exists); err != nil {
		return fmt.Errorf("set day clubs: %w", err)
	}
	if !exists {
		return fmt.Errorf("%w: child %d", ErrNotFound, childID)
	}
	d := date.Format(time.DateOnly)
	if _, err := tx.ExecContext(ctx, `DELETE FROM day_clubs WHERE date = ? AND child_id = ?`, d, childID); err != nil {
		return fmt.Errorf("set day clubs: %w", err)
	}
	for _, c := range clubs {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO day_clubs (date, child_id, start, name) VALUES (?, ?, ?, ?)`,
			d, childID, c.Start, c.Name)
		if err != nil {
			return fmt.Errorf("set day clubs: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("set day clubs: %w", err)
	}
	return nil
}
