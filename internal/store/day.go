package store

import (
	"context"
	"fmt"
	"time"
)

type DayClub struct {
	ChildID int64
	Start   string
	Name    string
}

func (s *Store) DayClubs(ctx context.Context, date time.Time) ([]DayClub, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT child_id, start, name FROM day_clubs WHERE date = ? ORDER BY start, name`,
		date.Format(time.DateOnly),
	)
	if err != nil {
		return nil, fmt.Errorf("list day clubs: %w", err)
	}
	defer rows.Close()

	clubs := []DayClub{}
	for rows.Next() {
		var c DayClub
		if err := rows.Scan(&c.ChildID, &c.Start, &c.Name); err != nil {
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
	defer tx.Rollback()

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
