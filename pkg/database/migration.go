package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// Migrator handles database migrations
type Migrator struct {
	db *sql.DB
}

// TableSchema defines a table's structure
type TableSchema struct {
	Name       string   // Table name
	CreateStmt string   // CREATE TABLE statement
	Indexes    []string // Index statements
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

// RunMigrations executes schema verification and creates missing tables on every restart
func (m *Migrator) RunMigrations() error {
	log.Println("[INFO] Starting database migration and schema verification...")

	// Enable UUID extension
	if err := m.enablePgCrypto(); err != nil {
		return err
	}

	// Verify and create all tables
	if err := m.verifyAndCreateTables(); err != nil {
		return err
	}

	log.Println("✓ Database migrations completed successfully")
	return nil
}

// enablePgCrypto enables the pgcrypto extension for UUID support
func (m *Migrator) enablePgCrypto() error {
	query := `CREATE EXTENSION IF NOT EXISTS "pgcrypto"`
	if _, err := m.db.Exec(query); err != nil {
		return fmt.Errorf("failed to enable pgcrypto extension: %w", err)
	}
	log.Println("[INFO] ✓ pgcrypto extension enabled")
	return nil
}

// createMigrationTable creates the migrations tracking table if it doesn't exist
func (m *Migrator) createMigrationTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

// isMigrationApplied checks if a migration has already been applied
func (m *Migrator) isMigrationApplied(name string) (bool, error) {
	query := "SELECT COUNT(*) FROM migrations WHERE name = $1"

	var count int
	err := m.db.QueryRow(query, name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check migration status: %w", err)
	}

	return count > 0, nil
}

// verifyAndCreateTables checks each table on every server restart and creates missing tables
func (m *Migrator) verifyAndCreateTables() error {
	log.Println("[INFO] Verifying database schema on startup...")

	// Get all table schemas
	schemas := m.getTableSchemas()

	// Check each table and create if missing
	for _, schema := range schemas {
		exists, err := m.tableExists(schema.Name)
		if err != nil {
			return err
		}

		if exists {
			log.Printf("[INFO] ✓ Table '%s' exists", schema.Name)
		} else {
			log.Printf("[WARN] ✗ Table '%s' NOT FOUND - creating...", schema.Name)
			if err := m.createTable(schema); err != nil {
				return err
			}
		}
	}

	log.Println("[INFO] ✓ Schema verification completed - all tables present")
	return nil
}

// verifySchema verifies that all required tables exist in the database
func (m *Migrator) verifySchema() error {
	log.Println("[INFO] Verifying database schema...")

	requiredTables := []string{
		"admin",
		"users",
		"pentesters",
		"stakeholders",
		"forgot_password_requests",
		"token_blacklist",
		"chat_messages",
		"projects",
		"project_members",
		"tasks",
		"task_submissions",
		"task_submission_attachments",
		"task_submission_reviews",
		"vulnerabilities",
		"messages",
		"message_attachments",
		"notifications",
		"notification_delivery_logs",
		"activities",
	}

	var missingTables []string

	for _, table := range requiredTables {
		exists, err := m.tableExists(table)
		if err != nil {
			return err
		}

		if exists {
			log.Printf("[INFO] ✓ Table '%s' exists", table)
		} else {
			log.Printf("[WARN] ✗ Table '%s' NOT FOUND", table)
			missingTables = append(missingTables, table)
		}
	}

	if len(missingTables) > 0 {
		log.Printf("[ERROR] Missing tables: %v", missingTables)
		log.Println("[INFO] Attempting to recreate missing tables...")
		if err := m.applyInitialSchema(); err != nil {
			return err
		}
	}

	log.Println("[INFO] Schema verification completed")
	return nil
}

// tableExists checks if a table exists in the database
func (m *Migrator) tableExists(tableName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)
	`

	var exists bool
	err := m.db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence for '%s': %w", tableName, err)
	}

	return exists, nil
}

// getTableSchemas returns all table schemas
func (m *Migrator) getTableSchemas() []TableSchema {
	return []TableSchema{
		// Admin table
		{
			Name: "admin",
			CreateStmt: `CREATE TABLE admin (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				email VARCHAR(255) UNIQUE NOT NULL,
				password_hash VARCHAR(255) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_admin_email ON admin(email)`,
			},
		},
		// Users table
		{
			Name: "users",
			CreateStmt: `CREATE TABLE users (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				email VARCHAR(255) UNIQUE NOT NULL,
				password_hash VARCHAR(255) NOT NULL,
				role VARCHAR(15) NOT NULL CHECK (role IN ('pentester', 'stakeholder')),
				status VARCHAR(15) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'freeze', 'suspended')),
				last_login TIMESTAMP,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_users_email ON users(email)`,
				`CREATE INDEX idx_users_status ON users(status)`,
				`CREATE INDEX idx_users_role ON users(role)`,
			},
		},
		// Pentesters table
		{
			Name: "pentesters",
			CreateStmt: `CREATE TABLE pentesters (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
				first_name VARCHAR(255) NOT NULL,
				last_name VARCHAR(255) NOT NULL,
				email VARCHAR(255) UNIQUE NOT NULL,
				specialization VARCHAR(255),
				experience_years INT,
				status VARCHAR(15) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'freeze', 'suspended')),
				skills TEXT,
				certifications TEXT,
				resume_file_path VARCHAR(500),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_pentesters_user_id ON pentesters(user_id)`,
				`CREATE INDEX idx_pentesters_first_name ON pentesters(first_name)`,
				`CREATE INDEX idx_pentesters_last_name ON pentesters(last_name)`,
				`CREATE INDEX idx_pentesters_email ON pentesters(email)`,
			},
		},
		// Stakeholders table
		{
			Name: "stakeholders",
			CreateStmt: `CREATE TABLE stakeholders (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
				first_name VARCHAR(255) NOT NULL,
				last_name VARCHAR(255) NOT NULL,
				email VARCHAR(255) UNIQUE NOT NULL,
				company VARCHAR(255),
				address TEXT,
				about TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_stakeholders_user_id ON stakeholders(user_id)`,
				`CREATE INDEX idx_stakeholders_email ON stakeholders(email)`,
			},
		},
		// Forgot Password Requests table
		{
			Name: "forgot_password_requests",
			CreateStmt: `CREATE TABLE forgot_password_requests (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID REFERENCES users(id) ON DELETE CASCADE,
				user_email VARCHAR(255) NOT NULL,
				status VARCHAR(15) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
				temporary_password VARCHAR(255),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_forgot_password_user_id ON forgot_password_requests(user_id)`,
				`CREATE INDEX idx_forgot_password_email ON forgot_password_requests(user_email)`,
				`CREATE INDEX idx_forgot_password_status ON forgot_password_requests(status)`,
			},
		},
		// Token Blacklist table
		{
			Name: "token_blacklist",
			CreateStmt: `CREATE TABLE token_blacklist (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				token_hash VARCHAR(255) UNIQUE NOT NULL,
				user_id UUID NOT NULL,
				user_type VARCHAR(15) NOT NULL CHECK (user_type IN ('admin', 'pentester', 'stakeholder')),
				blacklisted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				expires_at TIMESTAMP NOT NULL
			)`,
			Indexes: []string{
				`CREATE INDEX idx_token_blacklist_user ON token_blacklist(user_id)`,
				`CREATE INDEX idx_token_blacklist_expires ON token_blacklist(expires_at)`,
			},
		},
		// Chat Messages table
		{
			Name: "chat_messages",
			CreateStmt: `CREATE TABLE chat_messages (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID NOT NULL,
				user_type VARCHAR(15) NOT NULL CHECK (user_type IN ('admin', 'pentester', 'stakeholder')),
				role VARCHAR(15) NOT NULL CHECK (role IN ('user', 'assistant')),
				content TEXT NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_chat_user_id ON chat_messages(user_id)`,
				`CREATE INDEX idx_chat_user_type ON chat_messages(user_type)`,
				`CREATE INDEX idx_chat_created_at ON chat_messages(created_at)`,
				`CREATE INDEX idx_chat_user_created ON chat_messages(user_id, created_at DESC)`,
			},
		},
		// Projects table
		{
			Name: "projects",
			CreateStmt: `CREATE TABLE projects (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				name VARCHAR(255) NOT NULL,
				description TEXT,
				status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active','paused','completed','archived')),
				start_date DATE,
				end_date DATE,
				created_by UUID REFERENCES admin(id) ON DELETE SET NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_projects_created_by ON projects(created_by)`,
				`CREATE INDEX idx_projects_status ON projects(status)`,
			},
		},
		// Project Members table
		{
			Name: "project_members",
			CreateStmt: `CREATE TABLE project_members (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
				user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				role VARCHAR(15) NOT NULL CHECK (role IN ('pentester','stakeholder')),
				added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(project_id, user_id)
			)`,
			Indexes: []string{
				`CREATE INDEX idx_pm_project ON project_members(project_id)`,
				`CREATE INDEX idx_pm_user ON project_members(user_id)`,
			},
		},
		// Tasks table
		{
			Name: "tasks",
			CreateStmt: `CREATE TABLE tasks (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
				assigned_to UUID REFERENCES pentesters(id) ON DELETE SET NULL,
				title VARCHAR(255) NOT NULL,
				description TEXT,
				priority VARCHAR(10) CHECK (priority IN ('low','medium','high','critical')),
				status VARCHAR(20) NOT NULL DEFAULT 'assigned' CHECK (status IN ('assigned','in_progress','submitted','review','closed')),
				deadline TIMESTAMP,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_tasks_project ON tasks(project_id)`,
				`CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to)`,
				`CREATE INDEX idx_tasks_status ON tasks(status)`,
			},
		},
		// Task Submissions table
		{
			Name: "task_submissions",
			CreateStmt: `CREATE TABLE task_submissions (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
				pentester_id UUID REFERENCES pentesters(id),
				notes TEXT,
				submission_file TEXT,
				submission_file_path VARCHAR(500),
				status VARCHAR(20) CHECK (status IN ('submitted','changes_requested','approved')),
				reviewed_by UUID ,
				review_notes TEXT,
				reviewed_at TIMESTAMP,
				submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_task_submissions_task ON task_submissions(task_id)`,
				`CREATE INDEX idx_task_submissions_pentester ON task_submissions(pentester_id)`,
				`CREATE INDEX idx_task_submissions_reviewed_by ON task_submissions(reviewed_by)`,
				`CREATE INDEX idx_task_submissions_status ON task_submissions(status)`,
			},
		},
		// Vulnerabilities table
		{
			Name: "vulnerabilities",
			CreateStmt: `CREATE TABLE vulnerabilities (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
				task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
				reported_by UUID ,
				title VARCHAR(255) NOT NULL,
				description TEXT,
				severity VARCHAR(10) CHECK (severity IN ('low','medium','high','critical')),
				cvss_score NUMERIC(3,1),
				cwe_id VARCHAR(20),
				owasp_category VARCHAR(100),
				affected_asset TEXT,
				status VARCHAR(25) NOT NULL DEFAULT 'new' CHECK (status IN ('new','triaged','assigned','in_progress','retesting','verified','resolved','closed')),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_vuln_project ON vulnerabilities(project_id)`,
				`CREATE INDEX idx_vuln_task ON vulnerabilities(task_id)`,
				`CREATE INDEX idx_vuln_reported_by ON vulnerabilities(reported_by)`,
				`CREATE INDEX idx_vuln_status ON vulnerabilities(status)`,
				`CREATE INDEX idx_vuln_severity ON vulnerabilities(severity)`,
			},
		},
		// Messages table (1-to-1 chat)
		{
			Name: "messages",
			CreateStmt: `CREATE TABLE messages (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				sender_id UUID NOT NULL,
				receiver_id UUID NOT NULL,
				content TEXT NOT NULL,
				status VARCHAR(15) NOT NULL DEFAULT 'sent' CHECK (status IN ('sent', 'delivered', 'read')),
				sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				delivered_at TIMESTAMP,
				read_at TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_messages_sender ON messages(sender_id)`,
				`CREATE INDEX idx_messages_receiver ON messages(receiver_id)`,
				`CREATE INDEX idx_messages_status ON messages(status)`,
				`CREATE INDEX idx_messages_sent_at ON messages(sent_at DESC)`,
				`CREATE INDEX idx_messages_conversation ON messages(sender_id, receiver_id, sent_at DESC)`,
			},
		},
		// Message Attachments table
		{
			Name: "message_attachments",
			CreateStmt: `CREATE TABLE message_attachments (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
				file_name VARCHAR(255) NOT NULL,
				file_path VARCHAR(500) NOT NULL,
				file_size INT NOT NULL,
				mime_type VARCHAR(100),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_msg_attachments_message ON message_attachments(message_id)`,
			},
		},
		// Task submission attachments table
		{
			Name: "task_submission_attachments",
			CreateStmt: `CREATE TABLE task_submission_attachments (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				submission_id UUID NOT NULL REFERENCES task_submissions(id) ON DELETE CASCADE,
				file_name VARCHAR(255) NOT NULL,
				file_path VARCHAR(500) NOT NULL,
				file_size INT NOT NULL,
				notes TEXT,
				uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_task_attachments_submission ON task_submission_attachments(submission_id)`,
			},
		},
		// Task submission reviews table
		{
			Name: "task_submission_reviews",
			CreateStmt: `CREATE TABLE task_submission_reviews (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				submission_id UUID NOT NULL REFERENCES task_submissions(id) ON DELETE CASCADE,
				reviewer_id UUID NOT NULL,
				reviewer_role VARCHAR(15) NOT NULL CHECK (reviewer_role IN ('admin', 'stakeholder')),
				status VARCHAR(20) NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'completed')),
				review_notes TEXT,
				reviewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_task_submission_reviews_submission ON task_submission_reviews(submission_id)`,
				`CREATE INDEX idx_task_submission_reviews_reviewer ON task_submission_reviews(reviewer_id)`,
			},
		},
		// Notifications table
		{
			Name: "notifications",
			CreateStmt: `CREATE TABLE notifications (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				sender_id UUID NOT NULL,
				receiver_type VARCHAR(15) NOT NULL CHECK (receiver_type IN ('single', 'multiple', 'broadcast')),
				receiver_ids JSON,
				subject VARCHAR(255) NOT NULL,
				message TEXT NOT NULL,
				priority VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'critical')),
				in_app BOOLEAN DEFAULT false,
				email BOOLEAN DEFAULT false,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				delivered_at TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_notifications_sender ON notifications(sender_id)`,
				`CREATE INDEX idx_notifications_priority ON notifications(priority)`,
				`CREATE INDEX idx_notifications_created ON notifications(created_at DESC)`,
			},
		},
		// Notification Delivery Logs table
		{
			Name: "notification_delivery_logs",
			CreateStmt: `CREATE TABLE notification_delivery_logs (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
				receiver_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				channel VARCHAR(20) NOT NULL CHECK (channel IN ('in_app', 'email')),
				status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'delivered', 'failed', 'bounced')),
				attempted_at TIMESTAMP,
				delivered_at TIMESTAMP,
				error_message TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_notif_logs_notification ON notification_delivery_logs(notification_id)`,
				`CREATE INDEX idx_notif_logs_receiver ON notification_delivery_logs(receiver_id)`,
				`CREATE INDEX idx_notif_logs_channel ON notification_delivery_logs(channel)`,
				`CREATE INDEX idx_notif_logs_status ON notification_delivery_logs(status)`,
			},
		}, // Activities/Audit Log table
		{
			Name: "activities",
			CreateStmt: `CREATE TABLE activities (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id UUID NOT NULL,
				email VARCHAR(255),
				activity_type VARCHAR(50) NOT NULL,
				description TEXT,
				metadata JSONB,
				ip_address VARCHAR(45),
				user_agent TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			Indexes: []string{
				`CREATE INDEX idx_activities_user ON activities(user_id)`,
				`CREATE INDEX idx_activities_type ON activities(activity_type)`,
				`CREATE INDEX idx_activities_created ON activities(created_at DESC)`,
				`CREATE INDEX idx_activities_user_created ON activities(user_id, created_at DESC)`,
			},
		}}
}

// createTable creates a single table with its indexes
func (m *Migrator) createTable(schema TableSchema) error {
	// Create table
	_, err := m.db.Exec(schema.CreateStmt)
	if err != nil {
		return fmt.Errorf("failed to create table '%s': %w", schema.Name, err)
	}
	log.Printf("[INFO] ✓ Table '%s' created successfully", schema.Name)

	// Create indexes for the table
	for _, indexStmt := range schema.Indexes {
		if _, err := m.db.Exec(indexStmt); err != nil {
			// Log error but continue - index might already exist from a previous run
			log.Printf("[WARN] Index creation: %v", err)
		}
	}

	return nil
}

// applyInitialSchema applies the initial database schema (fallback function)
func (m *Migrator) applyInitialSchema() error {
	log.Println("[INFO] Applying initial schema...")

	// Get all table schemas
	schemas := m.getTableSchemas()

	// Create each table
	for _, schema := range schemas {
		if err := m.createTable(schema); err != nil {
			return err
		}
	}

	log.Println("[INFO] ✓ Initial schema applied successfully")
	return nil
}

// extractTableName extracts table name from CREATE TABLE statement (legacy - not used in new logic)
func extractTableName(stmt string) string {
	stmt = strings.ToUpper(stmt)
	idx := strings.Index(stmt, "CREATE TABLE IF NOT EXISTS")
	if idx >= 0 {
		start := idx + len("CREATE TABLE IF NOT EXISTS")
		rest := stmt[start:]
		rest = strings.TrimSpace(rest)
		end := strings.IndexAny(rest, " \t\n(")
		if end > 0 {
			return strings.TrimSpace(rest[:end])
		}
		return strings.TrimSpace(rest)
	}

	idx = strings.Index(stmt, "CREATE TABLE")
	if idx >= 0 {
		start := idx + len("CREATE TABLE")
		rest := stmt[start:]
		rest = strings.TrimSpace(rest)
		end := strings.IndexAny(rest, " \t\n(")
		if end > 0 {
			return strings.TrimSpace(rest[:end])
		}
		return strings.TrimSpace(rest)
	}

	return "unknown"
}

// RollbackLastMigration rolls back the most recent migration (for development only)
func (m *Migrator) RollbackLastMigration() error {
	query := "DELETE FROM migrations ORDER BY id DESC LIMIT 1 RETURNING name"

	var name string
	err := m.db.QueryRow(query).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no migrations to rollback")
		}
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	log.Printf("✓ Rolled back migration: %s\n", name)
	return nil
}

// GetMigrationStatus returns the status of all applied migrations
func (m *Migrator) GetMigrationStatus() ([]string, error) {
	query := "SELECT name FROM migrations ORDER BY id DESC"

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration status: %w", err)
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan migration: %w", err)
		}
		migrations = append(migrations, name)
	}

	return migrations, nil
}

// DropAllTables drops all application tables (for development/testing only)
func (m *Migrator) DropAllTables() error {
	// Drop in reverse order of creation to respect foreign key constraints
	dropStatements := []string{
		"DROP TABLE IF EXISTS pentesters CASCADE",
		"DROP TABLE IF EXISTS stakeholders CASCADE",
		"DROP TABLE IF EXISTS users CASCADE",
		"DROP TABLE IF EXISTS admin CASCADE",
		"DELETE FROM migrations",
	}

	for _, stmt := range dropStatements {
		if _, err := m.db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to drop tables: %w\n%s", err, stmt)
		}
	}

	log.Println("✓ All tables dropped successfully")
	return nil
}
