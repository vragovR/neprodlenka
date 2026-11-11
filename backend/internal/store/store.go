package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrNotAttending = errors.New("child does not attend this day")

type Status string

const (
	Pending Status = "pending"
	Present Status = "present"
	Absent  Status = "absent"
)

type PostStatus string

const (
	PostOpen   PostStatus = "open"
	PostClosed PostStatus = "closed"
)

type Post struct {
	Date      time.Time
	Status    PostStatus
	MessageID string
}

type DayChild struct {
	ID         int64
	LastName   string
	FirstName  string
	MiddleName string
	Active     bool
	Status     Status
	Price      int64
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Post(ctx context.Context, date time.Time) (Post, bool, error) {
	var (
		p         Post
		messageID sql.NullString
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT status, message_id FROM daily_posts WHERE date = ?`,
		date.Format(time.DateOnly),
	).Scan(&p.Status, &messageID)
	if errors.Is(err, sql.ErrNoRows) {
		return Post{}, false, nil
	}
	if err != nil {
		return Post{}, false, fmt.Errorf("get post: %w", err)
	}
	p.Date = date
	p.MessageID = messageID.String
	return p, true, nil
}

func (s *Store) MonthPosts(ctx context.Context, month time.Time) (map[string]Post, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT date, status, message_id FROM daily_posts WHERE date LIKE ?`,
		month.Format(monthLayout)+"-%",
	)
	if err != nil {
		return nil, fmt.Errorf("get month posts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	posts := map[string]Post{}
	for rows.Next() {
		var (
			date      string
			p         Post
			messageID sql.NullString
		)
		if err := rows.Scan(&date, &p.Status, &messageID); err != nil {
			return nil, fmt.Errorf("scan month post: %w", err)
		}
		if p.Date, err = parseDate(date); err != nil {
			return nil, err
		}
		p.MessageID = messageID.String
		posts[date] = p
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get month posts: %w", err)
	}
	return posts, nil
}

func (s *Store) DeletePost(ctx context.Context, date time.Time) error {
	return s.forgetPost(ctx, date, `DELETE FROM daily_posts WHERE date = ? AND message_id IS NULL`)
}

func (s *Store) RemovePost(ctx context.Context, date time.Time) error {
	return s.forgetPost(ctx, date, `DELETE FROM daily_posts WHERE date = ?`)
}

func (s *Store) forgetPost(ctx context.Context, date time.Time, query string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	d := date.Format(time.DateOnly)
	res, err := tx.ExecContext(ctx, query, d)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	if n == 0 {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM attendance WHERE date = ? AND status = 'pending'`, d); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM day_clubs WHERE date = ?`, d); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	return nil
}

func (s *Store) ClosePost(ctx context.Context, date time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("close post: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	d := date.Format(time.DateOnly)
	_, err = tx.ExecContext(ctx, `
		DELETE FROM attendance WHERE date = ? AND status = 'pending'
		AND child_id IN (SELECT id FROM children WHERE active = 0 OR weekdays & ? = 0)`, d, weekdayBit(date))
	if err != nil {
		return fmt.Errorf("close post: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO attendance (date, child_id, status, price)
		SELECT ?, id, 'present', price FROM children WHERE active = 1 AND weekdays & ? <> 0
		ON CONFLICT (date, child_id) DO UPDATE SET status = excluded.status, price = excluded.price
		WHERE attendance.status = 'pending'`, d, weekdayBit(date))
	if err != nil {
		return fmt.Errorf("close post: %w", err)
	}
	_, err = tx.ExecContext(ctx, `UPDATE daily_posts SET status = 'closed' WHERE date = ?`, d)
	if err != nil {
		return fmt.Errorf("close post: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("close post: %w", err)
	}
	return nil
}

func (s *Store) IsSchoolDay(ctx context.Context, date time.Time) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM no_school_days WHERE date = ?`,
		date.Format(time.DateOnly),
	).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("check school day: %w", err)
	}
	return n == 0, nil
}

func (s *Store) IsLastSchoolDay(ctx context.Context, date time.Time) (bool, error) {
	end := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, date.Location())
	var today, noSchool int
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(date = ?1), 0), COUNT(*) FROM no_school_days WHERE date BETWEEN ?1 AND ?2`,
		date.Format(time.DateOnly), end.Format(time.DateOnly),
	).Scan(&today, &noSchool)
	if err != nil {
		return false, fmt.Errorf("check last school day: %w", err)
	}
	return today == 0 && noSchool == end.Day()-date.Day(), nil
}

func (s *Store) CreatePost(ctx context.Context, date time.Time) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("create post: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	d := date.Format(time.DateOnly)
	res, err := tx.ExecContext(ctx, `INSERT INTO daily_posts (date) VALUES (?) ON CONFLICT DO NOTHING`, d)
	if err != nil {
		return false, fmt.Errorf("create post: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("create post: %w", err)
	}
	if n == 0 {
		return false, nil
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO attendance (date, child_id, status, price)
		SELECT ?, id, 'pending', 0 FROM children WHERE active = 1 AND weekdays & ? <> 0
		ON CONFLICT (date, child_id) DO NOTHING`, d, weekdayBit(date))
	if err != nil {
		return false, fmt.Errorf("create post: %w", err)
	}
	clubs, err := plannedClubs(ctx, tx, date)
	if err != nil {
		return false, fmt.Errorf("create post: %w", err)
	}
	for _, c := range clubs {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO day_clubs (date, child_id, start, name) VALUES (?, ?, ?, ?)`,
			d, c.ChildID, c.Start, c.Name)
		if err != nil {
			return false, fmt.Errorf("create post: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("create post: %w", err)
	}
	return true, nil
}

func Weekday(date time.Time) int {
	return (int(date.Weekday())+6)%7 + 1
}

func weekdayBit(date time.Time) int {
	return 1 << (Weekday(date) - 1)
}

func (s *Store) SetPostMessageID(ctx context.Context, date time.Time, messageID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("set post message id: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	d := date.Format(time.DateOnly)
	if _, err := tx.ExecContext(ctx, `UPDATE daily_posts SET message_id = ? WHERE date = ?`, messageID, d); err != nil {
		return fmt.Errorf("set post message id: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM club_changes WHERE date <= ?`, d); err != nil {
		return fmt.Errorf("set post message id: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("set post message id: %w", err)
	}
	return nil
}

func (s *Store) ToggleAbsent(ctx context.Context, messageID string, userID int64) (time.Time, bool, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return time.Time{}, false, false, fmt.Errorf("toggle absent: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var date string
	err = tx.QueryRowContext(ctx,
		`SELECT date FROM daily_posts WHERE message_id = ? AND status = 'open'`, messageID,
	).Scan(&date)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, false, false, nil
	}
	if err != nil {
		return time.Time{}, false, false, fmt.Errorf("toggle absent: %w", err)
	}

	var linked, total, absent int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*), COUNT(a.status), COALESCE(SUM(a.status = 'absent'), 0)
		FROM child_parents cp
		JOIN children c ON c.id = cp.child_id
		LEFT JOIN attendance a ON a.child_id = c.id AND a.date = ?1
		WHERE cp.user_id = ?2 AND c.active = 1`,
		date, userID,
	).Scan(&linked, &total, &absent)
	if err != nil {
		return time.Time{}, false, false, fmt.Errorf("toggle absent: %w", err)
	}
	if linked == 0 {
		return time.Time{}, false, false, nil
	}
	if total == 0 {
		return time.Time{}, false, false, fmt.Errorf("toggle absent: %w", ErrNotAttending)
	}

	status := Absent
	if absent == total {
		status = Pending
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE attendance SET status = ?3, price = 0
		WHERE date = ?1 AND child_id IN (
			SELECT c.id FROM child_parents cp JOIN children c ON c.id = cp.child_id
			WHERE cp.user_id = ?2 AND c.active = 1)`,
		date, userID, status)
	if err != nil {
		return time.Time{}, false, false, fmt.Errorf("toggle absent: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return time.Time{}, false, false, fmt.Errorf("toggle absent: %w", err)
	}

	d, err := parseDate(date)
	if err != nil {
		return time.Time{}, false, false, fmt.Errorf("toggle absent: %w", err)
	}
	return d, absent != total, true, nil
}

func (s *Store) DayChildren(ctx context.Context, date time.Time) ([]DayChild, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, c.last_name, c.first_name, c.middle_name, c.active, a.status, a.price
		FROM children c
		LEFT JOIN attendance a ON a.child_id = c.id AND a.date = ?
		WHERE (c.active = 1 AND c.weekdays & ? <> 0) OR a.status IS NOT NULL`,
		date.Format(time.DateOnly), weekdayBit(date),
	)
	if err != nil {
		return nil, fmt.Errorf("get day children: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var children []DayChild
	for rows.Next() {
		var (
			c      DayChild
			status sql.NullString
			price  sql.NullInt64
		)
		if err := rows.Scan(&c.ID, &c.LastName, &c.FirstName, &c.MiddleName, &c.Active, &status, &price); err != nil {
			return nil, fmt.Errorf("scan day child: %w", err)
		}
		c.Status = Status(status.String)
		c.Price = price.Int64
		children = append(children, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get day children: %w", err)
	}
	return children, nil
}

func parseDate(s string) (time.Time, error) {
	d, err := time.ParseInLocation(time.DateOnly, s, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date %q: %w", s, err)
	}
	return d, nil
}
