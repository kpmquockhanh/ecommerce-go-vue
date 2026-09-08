package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"ecommerce-api-go/internal/models"
)

type UserRepository interface {
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Create(ctx context.Context, email, passwordHash, firstName, lastName string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, string, error)
	FindByID(ctx context.Context, id int) (*models.User, error)
	UpdateProfile(ctx context.Context, id int, firstName, lastName string) error
	UpdateRole(ctx context.Context, id int, role string) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, limit, offset int) ([]models.User, error)
	Count(ctx context.Context) (int, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	return exists, err
}

func (r *userRepository) Create(ctx context.Context, email, passwordHash, firstName, lastName string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, first_name, last_name) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id, email, first_name, last_name, role, created_at`,
		email, passwordHash, firstName, lastName,
	).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, string, error) {
	var user models.User
	var passwordHash string
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, first_name, last_name, role, created_at 
		 FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Email, &passwordHash, &user.FirstName, &user.LastName, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, "", err
	}
	return &user, passwordHash, nil
}

func (r *userRepository) FindByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx,
		`SELECT id, email, first_name, last_name, role, created_at 
		 FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateProfile(ctx context.Context, id int, firstName, lastName string) error {
	result, err := r.db.Exec(ctx,
		`UPDATE users SET first_name = $1, last_name = $2, updated_at = NOW() WHERE id = $3`,
		firstName, lastName, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *userRepository) UpdateRole(ctx context.Context, id int, role string) error {
	result, err := r.db.Exec(ctx,
		`UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`, role, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, email, first_name, last_name, role, created_at 
		 FROM users ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *userRepository) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	return total, err
}
