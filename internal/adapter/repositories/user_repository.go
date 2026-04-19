package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
)

// UserRepositoryImpl implements the UserRepository interface
type UserRepositoryImpl struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &UserRepositoryImpl{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepositoryImpl) CreateUser(user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.Status,
		time.Now().UTC(),
		time.Now().UTC(),
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepositoryImpl) GetUserByID(id string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, role, status, last_login, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &domain.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepositoryImpl) GetUserByEmail(email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, role, status, last_login, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &domain.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetAnyUserByID retrieves any user (from users, admin, pentester, or stakeholder table)
// This is useful for messaging/notifications where users can be of any type
func (r *UserRepositoryImpl) GetAnyUserByID(id string) (*domain.User, error) {
	// First try to get from users table
	query := `
		SELECT id, email, password_hash, role, status, last_login, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &domain.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == nil {
		return user, nil
	}

	// If not found in users table, check admin table
	adminQuery := `
		SELECT id, email, password_hash, 'admin' as role, 'active' as status, NULL as last_login, created_at, updated_at
		FROM admin
		WHERE id = $1
	`

	err = r.db.QueryRow(adminQuery, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == nil {
		return user, nil
	}

	// If not found in either table
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}

	return nil, fmt.Errorf("failed to get user: %w", err)
}

// UpdateUser updates an existing user
func (r *UserRepositoryImpl) UpdateUser(user *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, role = $3, status = $4, updated_at = $5
		WHERE id = $6
	`

	_, err := r.db.Exec(query,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.Status,
		time.Now().UTC(),
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdateUserLastLogin updates the last login timestamp for a user
func (r *UserRepositoryImpl) UpdateUserLastLogin(userID string) error {
	query := `
		UPDATE users
		SET last_login = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(query, time.Now().UTC(), time.Now().UTC(), userID)

	if err != nil {
		return fmt.Errorf("failed to update user last login: %w", err)
	}

	return nil
}

// GetUserRole retrieves the role of a user
func (r *UserRepositoryImpl) GetUserRole(userID string) (string, error) {
	query := `
		SELECT role FROM users WHERE id = $1
	`

	var role string
	err := r.db.QueryRow(query, userID).Scan(&role)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user not found")
		}
		return "", fmt.Errorf("failed to get user role: %w", err)
	}

	return role, nil
}

// UpdateUserStatus updates the status of a user (active, freeze, suspended)
func (r *UserRepositoryImpl) UpdateUserStatus(userID string, status string) error {
	query := `
		UPDATE users
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(query, status, time.Now().UTC(), userID)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// DeleteUser deletes a user by ID
func (r *UserRepositoryImpl) DeleteUser(id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// GetAllUsers retrieves all users from the database
func (r *UserRepositoryImpl) GetAllUsers() ([]*domain.User, error) {
	query := `
		SELECT id, email, password_hash, role, status, last_login, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.Status,
			&user.LastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return users, nil
}

// GetAllUsersPaginated retrieves paginated users from the database
func (r *UserRepositoryImpl) GetAllUsersPaginated(page, limit int) ([]*domain.User, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Max 100 per page
	}

	offset := (page - 1) * limit

	// Get total count
	countQuery := `SELECT COUNT(*) FROM users`
	var total int64
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, email, password_hash, role, status, last_login, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.Status,
			&user.LastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return users, total, nil
}

// CountAllUsers retrieves the total count of all users
func (r *UserRepositoryImpl) CountAllUsers() (int64, error) {
	query := `SELECT COUNT(*) FROM users`
	var count int64
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// UserWithName represents a user with their name for messaging
type UserWithName struct {
	ID        string
	Email     string
	Role      string
	FirstName string
	LastName  string
}

// GetAllUsersWithNames retrieves all users (admin + pentester + stakeholder) with their names
func (r *UserRepositoryImpl) GetAllUsersWithNames() ([]*UserWithName, error) {
	var users []*UserWithName

	// Get all regular users (pentester/stakeholder) with their names
	userQuery := `
		SELECT u.id, u.email, u.role, COALESCE(p.first_name, s.first_name, ''), COALESCE(p.last_name, s.last_name, '')
		FROM users u
		LEFT JOIN pentesters p ON u.id = p.user_id
		LEFT JOIN stakeholders s ON u.id = s.user_id
		ORDER BY u.created_at DESC
	`

	rows, err := r.db.Query(userQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user UserWithName
		if err := rows.Scan(&user.ID, &user.Email, &user.Role, &user.FirstName, &user.LastName); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	// Get all admins
	adminQuery := `SELECT id, email FROM admin ORDER BY created_at DESC`

	adminRows, err := r.db.Query(adminQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get admins: %w", err)
	}
	defer adminRows.Close()

	for adminRows.Next() {
		var admin UserWithName
		admin.Role = "admin"
		if err := adminRows.Scan(&admin.ID, &admin.Email); err != nil {
			return nil, fmt.Errorf("failed to scan admin: %w", err)
		}
		users = append(users, &admin)
	}

	if err := adminRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating admins: %w", err)
	}

	return users, nil
}
