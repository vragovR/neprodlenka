package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const monthLayout = "2006-01"

func (s *Store) NoSchoolDays(ctx context.Context, month time.Time) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT date FROM no_school_days WHERE date LIKE ? ORDER BY date`,
		month.Format(monthLayout)+"-%",
	)
	if err != nil {
		return nil, fmt.Errorf("list no school days: %w", err)
	}
	defer rows.Close()

	days := []string{}
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return nil, fmt.Errorf("scan no school day: %w", err)
		}
		days = append(days, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list no school days: %w", err)
	}
	return days, nil
}

func (s *Store) SetNoSchoolDays(ctx context.Context, month time.Time, days []time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("set no school days: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM no_school_days WHERE date LIKE ?`, month.Format(monthLayout)+"-%"); err != nil {
		return fmt.Errorf("set no school days: %w", err)
	}
	for _, d := range days {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO no_school_days (date) VALUES (?) ON CONFLICT DO NOTHING`, d.Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("set no school days: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("set no school days: %w", err)
	}
	return nil
}

func (s *Store) ReportUID(ctx context.Context, month time.Time) (string, error) {
	m := month.Format(monthLayout)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO report_months (month, uid) VALUES (?, ?) ON CONFLICT (month) DO NOTHING`,
		m, uuid.NewString())
	if err != nil {
		return "", fmt.Errorf("create report month: %w", err)
	}
	var uid string
	if err := s.db.QueryRowContext(ctx, `SELECT uid FROM report_months WHERE month = ?`, m).Scan(&uid); err != nil {
		return "", fmt.Errorf("get report uid: %w", err)
	}
	return uid, nil
}

type ReportMonth struct {
	UID       string
	MessageID string
}

func (s *Store) ReportMonth(ctx context.Context, month time.Time) (ReportMonth, error) {
	uid, err := s.ReportUID(ctx, month)
	if err != nil {
		return ReportMonth{}, err
	}
	var messageID sql.NullString
	err = s.db.QueryRowContext(ctx, `SELECT message_id FROM report_months WHERE month = ?`, month.Format(monthLayout)).Scan(&messageID)
	if err != nil {
		return ReportMonth{}, fmt.Errorf("get report month: %w", err)
	}
	return ReportMonth{UID: uid, MessageID: messageID.String}, nil
}

func (s *Store) SetReportMessageID(ctx context.Context, month time.Time, messageID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE report_months SET message_id = NULLIF(?, '') WHERE month = ?`,
		messageID, month.Format(monthLayout),
	)
	if err != nil {
		return fmt.Errorf("set report message id: %w", err)
	}
	return nil
}

func (s *Store) SummaryMonths(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT month FROM report_months WHERE message_id IS NOT NULL ORDER BY month`)
	if err != nil {
		return nil, fmt.Errorf("list summary months: %w", err)
	}
	defer rows.Close()
	months := []string{}
	for rows.Next() {
		var m string
		if err := rows.Scan(&m); err != nil {
			return nil, fmt.Errorf("list summary months: %w", err)
		}
		months = append(months, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list summary months: %w", err)
	}
	return months, nil
}
