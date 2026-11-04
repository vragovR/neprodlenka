package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrAdminExists = errors.New("admin already exists")

type Admin struct {
	ID           int64
	Login        string
	PasswordHash string
}

func (s *Store) AddAdmin(ctx context.Context, login, passwordHash string) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO admins (login, password_hash) VALUES (?, ?) ON CONFLICT (login) DO NOTHING`,
		login, passwordHash,
	)
	if err != nil {
		return fmt.Errorf("add admin: %w", err)
	}
	if n, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("add admin: %w", err)
	} else if n == 0 {
		return ErrAdminExists
	}
	return nil
}

func (s *Store) SetAdminPassword(ctx context.Context, login, passwordHash string) (bool, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE admins SET password_hash = ? WHERE login = ?`, passwordHash, login)
	if err != nil {
		return false, fmt.Errorf("set admin password: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("set admin password: %w", err)
	}
	return n == 1, nil
}

func (s *Store) RemoveAdmin(ctx context.Context, login string) (bool, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM admins WHERE login = ?`, login)
	if err != nil {
		return false, fmt.Errorf("remove admin: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("remove admin: %w", err)
	}
	return n == 1, nil
}

func (s *Store) AdminLogins(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT login FROM admins ORDER BY login`)
	if err != nil {
		return nil, fmt.Errorf("list admins: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var logins []string
	for rows.Next() {
		var login string
		if err := rows.Scan(&login); err != nil {
			return nil, fmt.Errorf("scan admin: %w", err)
		}
		logins = append(logins, login)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list admins: %w", err)
	}
	return logins, nil
}

func (s *Store) AdminByLogin(ctx context.Context, login string) (Admin, bool, error) {
	return s.admin(ctx, `SELECT id, login, password_hash FROM admins WHERE login = ?`, login)
}

func (s *Store) AdminByID(ctx context.Context, id int64) (Admin, bool, error) {
	return s.admin(ctx, `SELECT id, login, password_hash FROM admins WHERE id = ?`, id)
}

func (s *Store) admin(ctx context.Context, query string, arg any) (Admin, bool, error) {
	var a Admin
	err := s.db.QueryRowContext(ctx, query, arg).Scan(&a.ID, &a.Login, &a.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return Admin{}, false, nil
	}
	if err != nil {
		return Admin{}, false, fmt.Errorf("get admin: %w", err)
	}
	return a, true, nil
}
