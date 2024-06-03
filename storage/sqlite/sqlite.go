package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/digkill/sso-auth-go/internal/storage"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

// Construct Storage
func New(storagePath string) (*Storage, error) {
	const method = "storage.sqlite.New"

	// Path db file
	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", method, err)
	}

	return &Storage{db: db}, nil
}

// SaveUser saves user to db.
func (s *Storage) SaveUser(ctx context.Context,
	email string,
	passHash []byte,
) (int64, error) {
	const method = "storage.sqlite.SaveUser"

	// Query added user
	stmt, err := s.db.Prepare("INSERT INTO users (email, pass_hash) VALUES (?,?)")
	defer stmt.Close()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", method, err)
	}

	// Run query, give params
	res, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", method, storage.ErrUserExists)
		}

		return 0, fmt.Errorf("%s: %w", method, err)
	}

	// Get ID created record
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", method, err)
	}

	return id, nil
}

//
