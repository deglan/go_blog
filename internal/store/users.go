package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrorDuplicatedEmail    = errors.New("duplicated email")
	ErrorDuplicatedUsername = errors.New("duplicated username")
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  Password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
	RoleId    int64    `json:"role_id"`
	Role      Role     `json:"role"`
}

type Password struct {
	text *string
	hash []byte
}

func (p *Password) Set(text string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.text = &text
	p.hash = bytes
	return nil
}

func (p *Password) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(password))
	return err == nil
}

func (p *Password) Hash() []byte {
	return p.hash
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
	INSERT INTO users (username, email, password, role_id) 
	VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = $4)) RETURNING id, created_at
	`

	role := user.Role
	if role.Name == "" {
		role.Name = "user"
	}

	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password.hash,
		role.Name,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrorDuplicatedEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrorDuplicatedUsername
		default:
			return err
		}
	}
	return nil
}

func (s *UserStore) GetById(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT users.id, users.username, users.email, users.created_at, users.is_active, roles.*
		FROM users
		JOIN roles ON users.role_id = roles.id
		WHERE users.id = $1
	`

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
		&user.IsActive,
		&user.RoleId,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &User{}, ErrorNotFound
		}
		return &User{}, err
	}

	return &user, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, username, email, password, created_at
		FROM users
		WHERE email = $1 AND is_active = true
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return &User{}, ErrorNotFound
		default:
			return &User{}, err
		}
	}
	return user, nil
}

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}

		err := s.createUserInvitation(ctx, tx, token, invitationExp, user.ID)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) Activate(ctx context.Context, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		user.IsActive = true
		if err := s.update(ctx, tx, user); err != nil {
			return err
		}

		err = s.deleteUserInvitation(ctx, tx, user.ID)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) Delete(ctx context.Context, userId int64) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.delete(ctx, tx, userId); err != nil {
			return err
		}
		if err := s.deleteUserInvitation(ctx, tx, userId); err != nil {
			return err
		}
		return nil
	})
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
		SELECT users.id, users.username, users.email, users.created_at
		FROM users
		JOIN users_invitations ON users.id = users_invitations.user_id
		WHERE users_invitations.token = $1 AND users_invitations.expiry > $2
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])
	user := &User{}
	err := tx.QueryRowContext(
		ctx,
		query,
		hashToken,
		time.Now(),
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrorNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, invitationExp time.Duration, userID int64) error {
	query := `
		INSERT INTO users_invitations (token, user_id, expiry)
		VALUES ($1, $2, $3)
		`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		token,
		userID,
		time.Now().Add(invitationExp),
	)
	return err
}

func (s *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, is_active = $3
		WHERE id = $4
		`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.IsActive,
		user.ID,
	)

	if err != nil {
		return err
	}
	return nil
}

func (s *UserStore) deleteUserInvitation(ctx context.Context, tx *sql.Tx, userId int64) error {
	query := `
		DELETE FROM users_invitations
		WHERE user_id = $1
		`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		userId,
	)
	return err
}

func (s *UserStore) delete(ctx context.Context, tx *sql.Tx, userId int64) error {
	query := `
		DELETE FROM users
		WHERE id = $1
		`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(
		ctx,
		query,
		userId,
	)
	return err
}
