package store

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"neprodlenka/internal/names"
)

var (
	ErrPostExists    = errors.New("post already exists")
	ErrChildInactive = errors.New("child is inactive")
)

type DayClub struct {
	ChildID   int64
	LastName  string
	FirstName string
	Start     string
	Name      string
}

func (s *Store) DayClubs(ctx context.Context, date time.Time) ([]DayClub, error) {
	return queryDayClubs(ctx, s.db, `
		SELECT dc.child_id, c.last_name, c.first_name, dc.start, dc.name
		FROM day_clubs dc JOIN children c ON c.id = dc.child_id
		WHERE dc.date = ?
		ORDER BY dc.start, dc.name`,
		date.Format(time.DateOnly),
	)
}

func queryDayClubs(ctx context.Context, q queryer, query string, args ...any) ([]DayClub, error) {
	rows, err := q.QueryContext(ctx, query, args...)
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

type ClubChange DayClub

func (c ClubChange) Canceled() bool {
	return c.Start == ""
}

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func (s *Store) PlannedClubs(ctx context.Context, date time.Time) ([]DayClub, error) {
	return plannedClubs(ctx, s.db, date)
}

func plannedClubs(ctx context.Context, q queryer, date time.Time) ([]DayClub, error) {
	schedule, err := queryDayClubs(ctx, q, `
		SELECT cc.child_id, c.last_name, c.first_name, cc.start, cc.name
		FROM child_clubs cc JOIN children c ON c.id = cc.child_id
		WHERE c.active = 1 AND c.weekdays & ? <> 0 AND cc.weekday = ?`,
		weekdayBit(date), Weekday(date),
	)
	if err != nil {
		return nil, err
	}
	changes, err := clubChanges(ctx, q, date)
	if err != nil {
		return nil, err
	}
	return applyClubChanges(schedule, changes), nil
}

func applyClubChanges(schedule []DayClub, changes []ClubChange) []DayClub {
	clubs := slices.DeleteFunc(slices.Clone(schedule), func(club DayClub) bool {
		return slices.ContainsFunc(changes, func(ch ClubChange) bool { return ch.ChildID == club.ChildID && ch.Name == club.Name })
	})
	for _, ch := range changes {
		if !ch.Canceled() {
			clubs = append(clubs, DayClub(ch))
		}
	}
	slices.SortStableFunc(clubs, func(a, b DayClub) int {
		return cmp.Or(strings.Compare(a.Start, b.Start), strings.Compare(a.Name, b.Name))
	})
	return clubs
}

func (s *Store) ClubChanges(ctx context.Context, date time.Time) ([]ClubChange, error) {
	changes, err := clubChanges(ctx, s.db, date)
	if err != nil {
		return nil, err
	}
	names.Sort(changes, func(c ClubChange) (string, string) { return c.LastName, c.FirstName })
	return changes, nil
}

func clubChanges(ctx context.Context, q queryer, date time.Time) ([]ClubChange, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT ch.child_id, c.last_name, c.first_name, COALESCE(ch.start, ''), ch.name
		FROM club_changes ch JOIN children c ON c.id = ch.child_id
		WHERE ch.date = ? AND c.active = 1
		ORDER BY ch.name`,
		date.Format(time.DateOnly),
	)
	if err != nil {
		return nil, fmt.Errorf("list club changes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	changes := []ClubChange{}
	for rows.Next() {
		var c ClubChange
		if err := rows.Scan(&c.ChildID, &c.LastName, &c.FirstName, &c.Start, &c.Name); err != nil {
			return nil, fmt.Errorf("scan club change: %w", err)
		}
		changes = append(changes, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list club changes: %w", err)
	}
	return changes, nil
}

type ChildClubChange struct {
	Date  string
	Name  string
	Start string
}

func (c ChildClubChange) Canceled() bool {
	return c.Start == ""
}

func (s *Store) ChildClubChanges(ctx context.Context, childID int64, from time.Time) ([]ChildClubChange, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT date, name, COALESCE(start, '')
		FROM club_changes
		WHERE child_id = ? AND date >= ?
		ORDER BY date, name`,
		childID, from.Format(time.DateOnly),
	)
	if err != nil {
		return nil, fmt.Errorf("list child club changes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	changes := []ChildClubChange{}
	for rows.Next() {
		var c ChildClubChange
		if err := rows.Scan(&c.Date, &c.Name, &c.Start); err != nil {
			return nil, fmt.Errorf("scan child club change: %w", err)
		}
		changes = append(changes, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list child club changes: %w", err)
	}
	return changes, nil
}

func (s *Store) SetClubChanges(ctx context.Context, date time.Time, childID int64, changes []ClubChange) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("set club changes: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	d := date.Format(time.DateOnly)
	var posted bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM daily_posts WHERE date = ?)`, d).Scan(&posted); err != nil {
		return fmt.Errorf("set club changes: %w", err)
	}
	if posted {
		return ErrPostExists
	}
	var active bool
	err = tx.QueryRowContext(ctx, `SELECT active FROM children WHERE id = ?`, childID).Scan(&active)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: child %d", ErrNotFound, childID)
	}
	if err != nil {
		return fmt.Errorf("set club changes: %w", err)
	}
	if !active {
		return fmt.Errorf("%w: child %d", ErrChildInactive, childID)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM club_changes WHERE date = ? AND child_id = ?`, d, childID); err != nil {
		return fmt.Errorf("set club changes: %w", err)
	}
	for _, c := range changes {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO club_changes (date, child_id, name, start) VALUES (?, ?, ?, NULLIF(?, ''))`,
			d, childID, c.Name, c.Start)
		if err != nil {
			return fmt.Errorf("set club changes: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("set club changes: %w", err)
	}
	return nil
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
