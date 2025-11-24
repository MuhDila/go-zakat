package postgres

import (
	"context"
	"errors"
	"time"

	"go-zakat/internal/domain/entity"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// UserRepository mengimplementasikan interface UserRepository
type UserRepository struct {
	db  *pgxpool.Pool
	log *logrus.Logger
}

// NewUserRepository membuat instance baru userRepository
func NewUserRepository(db *pgxpool.Pool, log *logrus.Logger) *UserRepository {
	return &UserRepository{db: db, log: log}
}

// timeout default untuk operasi DB, supaya ga nunggu selamanya kalau DB bermasalah
const dbTimeout = 5 * time.Second

func (r *UserRepository) Create(user *entity.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		INSERT INTO users (id, email, password, google_id, name, role, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at;
	`

	var googleID interface{}
	if user.GoogleID != nil {
		googleID = *user.GoogleID
	} else {
		googleID = nil
	}

	// Default role if empty
	if user.Role == "" {
		user.Role = entity.RoleViewer
	}

	err := r.db.QueryRow(ctx, query, user.Email, user.Password, googleID, user.Name, user.Role).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"email":    user.Email,
			"googleID": googleID,
		}).Error("gagal insert user ke database: ", err)

		return err
	}

	// contoh logging sukses
	r.log.WithFields(logrus.Fields{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	}).Info("berhasil membuat user baru")

	return nil
}

func (r *UserRepository) FindByEmail(email string) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, email, password, google_id, name, role, created_at, updated_at
		FROM users
		WHERE email = $1
		LIMIT 1;
	`

	row := r.db.QueryRow(ctx, query, email)

	user := &entity.User{}
	var googleID *string
	err := row.Scan(&user.ID, &user.Email, &user.Password, &googleID, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		// kalau no rows, sebaiknya kembalikan error khusus "not found"
		return nil, err
	}
	user.GoogleID = googleID
	return user, nil
}

func (r *UserRepository) FindByID(id string) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, email, password, google_id, name, role, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1;
	`

	row := r.db.QueryRow(ctx, query, id)

	user := &entity.User{}
	var googleID *string
	err := row.Scan(&user.ID, &user.Email, &user.Password, &googleID, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	user.GoogleID = googleID
	return user, nil
}

func (r *UserRepository) FindByGoogleID(googleID string) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, email, password, google_id, name, role, created_at, updated_at
		FROM users
		WHERE google_id = $1
		LIMIT 1;
	`

	row := r.db.QueryRow(ctx, query, googleID)

	user := &entity.User{}
	var googleIDPtr *string
	err := row.Scan(&user.ID, &user.Email, &user.Password, &googleIDPtr, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	user.GoogleID = googleIDPtr
	return user, nil
}

func (r *UserRepository) Update(user *entity.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		UPDATE users
		SET email = $1,
			password = $2,
			google_id = $3,
			name = $4,
			role = $5,
			updated_at = NOW()
		WHERE id = $6;
	`

	var googleID interface{}
	if user.GoogleID != nil {
		googleID = *user.GoogleID
	} else {
		googleID = nil
	}

	ct, err := r.db.Exec(ctx, query, user.Email, user.Password, googleID, user.Name, user.Role, user.ID)
	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}
