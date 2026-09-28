package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"neprodlenka/internal/names"
)

var ErrNotFound = errors.New("not found")

type Child struct {
	ID         int64
	LastName   string
	FirstName  string
	MiddleName string
	Price      int64
	Active     bool
	ParentIDs  []int64
	Clubs      []Club
}

type Club struct {
	Weekday int
	Start   string
	Name    string
}

func (s *Store) Children(ctx context.Context) ([]Child, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, last_name, first_name, middle_name, price, active FROM children`)
	if err != nil {
		return nil, fmt.Errorf("list children: %w", err)
	}
	defer rows.Close()

	children := []Child{}
	index := map[int64]int{}
	for rows.Next() {
		c := Child{ParentIDs: []int64{}, Clubs: []Club{}}
		if err := rows.Scan(&c.ID, &c.LastName, &c.FirstName, &c.MiddleName, &c.Price, &c.Active); err != nil {
			return nil, fmt.Errorf("scan child: %w", err)
		}
		index[c.ID] = len(children)
		children = append(children, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list children: %w", err)
	}

	parents, err := s.db.QueryContext(ctx, `SELECT user_id, child_id FROM child_parents ORDER BY user_id`)
	if err != nil {
		return nil, fmt.Errorf("list parents: %w", err)
	}
	defer parents.Close()
	for parents.Next() {
		var userID, childID int64
		if err := parents.Scan(&userID, &childID); err != nil {
			return nil, fmt.Errorf("scan parent: %w", err)
		}
		c := &children[index[childID]]
		c.ParentIDs = append(c.ParentIDs, userID)
	}
	if err := parents.Err(); err != nil {
		return nil, fmt.Errorf("list parents: %w", err)
	}

	clubs, err := s.db.QueryContext(ctx, `SELECT child_id, weekday, start, name FROM child_clubs ORDER BY weekday, start, name`)
	if err != nil {
		return nil, fmt.Errorf("list clubs: %w", err)
	}
	defer clubs.Close()
	for clubs.Next() {
		var childID int64
		var club Club
		if err := clubs.Scan(&childID, &club.Weekday, &club.Start, &club.Name); err != nil {
			return nil, fmt.Errorf("scan club: %w", err)
		}
		c := &children[index[childID]]
		c.Clubs = append(c.Clubs, club)
	}
	if err := clubs.Err(); err != nil {
		return nil, fmt.Errorf("list clubs: %w", err)
	}

	names.Sort(children, func(c Child) (string, string) { return c.LastName, c.FirstName })
	return children, nil
}

func (s *Store) ChildByID(ctx context.Context, id int64) (Child, bool, error) {
	c := Child{ID: id, ParentIDs: []int64{}, Clubs: []Club{}}
	err := s.db.QueryRowContext(ctx,
		`SELECT last_name, first_name, middle_name, price, active FROM children WHERE id = ?`, id,
	).Scan(&c.LastName, &c.FirstName, &c.MiddleName, &c.Price, &c.Active)
	if errors.Is(err, sql.ErrNoRows) {
		return Child{}, false, nil
	}
	if err != nil {
		return Child{}, false, fmt.Errorf("get child: %w", err)
	}
	return c, true, nil
}

func (s *Store) AddChild(ctx context.Context, c Child) (Child, error) {
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO children (last_name, first_name, middle_name, price) VALUES (?, ?, ?, ?)
		RETURNING id, active`,
		c.LastName, c.FirstName, c.MiddleName, c.Price,
	).Scan(&c.ID, &c.Active)
	if err != nil {
		return Child{}, fmt.Errorf("add child: %w", err)
	}
	c.ParentIDs = []int64{}
	c.Clubs = []Club{}
	return c, nil
}

func (s *Store) UpdateChild(ctx context.Context, id int64, update func(*Child) error) (Child, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Child{}, fmt.Errorf("update child: %w", err)
	}
	defer tx.Rollback()

	c := Child{ID: id, ParentIDs: []int64{}, Clubs: []Club{}}
	err = tx.QueryRowContext(ctx,
		`SELECT last_name, first_name, middle_name, price, active FROM children WHERE id = ?`, id,
	).Scan(&c.LastName, &c.FirstName, &c.MiddleName, &c.Price, &c.Active)
	if errors.Is(err, sql.ErrNoRows) {
		return Child{}, ErrNotFound
	}
	if err != nil {
		return Child{}, fmt.Errorf("get child: %w", err)
	}

	rows, err := tx.QueryContext(ctx, `SELECT user_id FROM child_parents WHERE child_id = ? ORDER BY user_id`, id)
	if err != nil {
		return Child{}, fmt.Errorf("get parents: %w", err)
	}
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			rows.Close()
			return Child{}, fmt.Errorf("scan parent: %w", err)
		}
		c.ParentIDs = append(c.ParentIDs, userID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return Child{}, fmt.Errorf("get parents: %w", err)
	}

	clubs, err := tx.QueryContext(ctx,
		`SELECT weekday, start, name FROM child_clubs WHERE child_id = ? ORDER BY weekday, start, name`, id)
	if err != nil {
		return Child{}, fmt.Errorf("get clubs: %w", err)
	}
	for clubs.Next() {
		var club Club
		if err := clubs.Scan(&club.Weekday, &club.Start, &club.Name); err != nil {
			clubs.Close()
			return Child{}, fmt.Errorf("scan club: %w", err)
		}
		c.Clubs = append(c.Clubs, club)
	}
	clubs.Close()
	if err := clubs.Err(); err != nil {
		return Child{}, fmt.Errorf("get clubs: %w", err)
	}

	if err := update(&c); err != nil {
		return Child{}, err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE children SET last_name = ?, first_name = ?, middle_name = ?, price = ?, active = ?
		WHERE id = ?`,
		c.LastName, c.FirstName, c.MiddleName, c.Price, c.Active, id)
	if err != nil {
		return Child{}, fmt.Errorf("update child: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM child_parents WHERE child_id = ?`, id); err != nil {
		return Child{}, fmt.Errorf("unbind parents: %w", err)
	}
	for _, userID := range c.ParentIDs {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO child_parents (user_id, child_id) VALUES (?, ?)`,
			userID, id)
		if err != nil {
			return Child{}, fmt.Errorf("bind parent: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM child_clubs WHERE child_id = ?`, id); err != nil {
		return Child{}, fmt.Errorf("delete clubs: %w", err)
	}
	for _, club := range c.Clubs {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO child_clubs (child_id, weekday, start, name) VALUES (?, ?, ?, ?)`,
			id, club.Weekday, club.Start, club.Name)
		if err != nil {
			return Child{}, fmt.Errorf("add club: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return Child{}, fmt.Errorf("update child: %w", err)
	}
	return c, nil
}
