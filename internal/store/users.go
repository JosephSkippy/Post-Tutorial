package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Password    password `json:"-"`
	CreatedAt   string   `json:"created_at"`
	IsActivated bool     `json:"is_activated"`
}

type password struct {
	text *string
	hash []byte
}

var (
	ErrDuplicateEmail    = errors.New("a user with that email already exists")
	ErrDuplicateUsername = errors.New("a user with that username already exists")
	ErrInvalidToken      = errors.New("nnvalid token or expired")
)

func (p *password) Set(text string) error {
	password, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	p.hash = password
	p.text = &text
	return nil
}

type UsersStore struct {
	db *sql.DB
}

func (s *UsersStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {

	query := `
			INSERT INTO users (username, email, password)
			VALUES ($1, $2, $3) RETURNING id, created_at
	`
	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password.hash).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (s *UsersStore) GetUserbyID(ctx context.Context, id int64) (*User, error) {

	query := `
			SELECT
				id,
				username,
				email,
				created_at
			
			FROM users WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	var user User
	err := s.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}

	}

	return &user, nil
}

func (s *UsersStore) CreateandInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {

	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}

		if err := s.createUserInvitation(ctx, tx, token, invitationExp, user.ID); err != nil {
			return err
		}

		return nil

	})

}

func (s *UsersStore) createUserInvitation(ctx context.Context, tx *sql.Tx, hashtoken string, invitationExp time.Duration, userID int64) error {

	query := `
			INSERT INTO user_invitation (token, user_id, expiry)
			VALUES ($1, $2, $3)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, hashtoken, userID, time.Now().Add(invitationExp))

	if err != nil {
		return err
	}

	return nil
}

func (s *UsersStore) Activate(ctx context.Context, hashtoken string) error {

	return withTx(s.db, ctx, func(tx *sql.Tx) error {

		//  get user from token
		user, err := s.getUserFromToken(ctx, tx, hashtoken)
		if err != nil {
			return err
		}

		// update user status
		if err := s.updateStatus(ctx, tx, user.ID); err != nil {
			return err
		}

		// delete invitation
		if err := s.deleteInvitation(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil

	})

}

func (s *UsersStore) getUserFromToken(ctx context.Context, tx *sql.Tx, hashtoken string) (*User, error) {

	query := `
			SELECT user_id FROM  users u
			JOIN user_invitation ui ON
			u.id = ui.user_id
			WHERE
			ui.token = $1 AND ui.expiry > $2
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	var user User
	current_time := time.Now()
	if err := tx.QueryRowContext(ctx, query, hashtoken, current_time).Scan(&user.ID); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrInvalidToken
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (s *UsersStore) updateStatus(ctx context.Context, tx *sql.Tx, userID int64) error {

	query := `
	UPDATE users
		SET is_active = true
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	if _, err := tx.ExecContext(ctx, query, userID); err != nil {
		return err
	}
	return nil
}

func (s *UsersStore) deleteInvitation(ctx context.Context, tx *sql.Tx, userID int64) error {

	query := `
	DELETE FROM user_invitation
	WHERE user_id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	if _, err := tx.ExecContext(ctx, query, userID); err != nil {
		return err
	}
	return nil
}

func (s *UsersStore) Delete(ctx context.Context, id int64) error {

	return withTx(s.db, ctx, func(tx *sql.Tx) error {

		// delete user
		if err := s.delete(ctx, tx, id); err != nil {
			return err
		}

		// delete invitation
		if err := s.deleteInvitation(ctx, tx, id); err != nil {
			return err
		}

		return nil

	})
}

func (s *UsersStore) delete(ctx context.Context, tx *sql.Tx, id int64) error {

	query := `
		DELETE FROM users
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	if _, err := tx.ExecContext(ctx, query, id); err != nil {
		return err
	}
	return nil
}
