-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 1. ADMIN
CREATE TABLE IF NOT EXISTS admin(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_admin_email ON admin(email);

-- 2. USERS 
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(15) NOT NULL CHECK (role IN ('pentester', 'stakeholder')),
    status VARCHAR(15) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'freeze', 'suspended')),
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

-- 3. Pentesters
CREATE TABLE IF NOT EXISTS pentesters (
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
);

CREATE INDEX idx_pentesters_user_id ON pentesters(user_id);
CREATE INDEX idx_pentesters_first_name ON pentesters(first_name);
CREATE INDEX idx_pentesters_last_name ON pentesters(last_name);
CREATE INDEX idx_pentesters_email ON pentesters(email);

-- 3.5 User Status and Role Indexes
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_role ON users(role);

-- 4. Stakeholders
CREATE TABLE IF NOT EXISTS stakeholders (
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
);

CREATE INDEX idx_stakeholders_user_id ON stakeholders(user_id);
CREATE INDEX idx_stakeholders_email ON stakeholders(email);

-- 5. Forgot Password Requests
CREATE TABLE IF NOT EXISTS forgot_password_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    user_email VARCHAR(255) NOT NULL,
    status VARCHAR(15) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    temporary_password VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_forgot_password_user_id ON forgot_password_requests(user_id);
CREATE INDEX idx_forgot_password_email ON forgot_password_requests(user_email);
CREATE INDEX idx_forgot_password_status ON forgot_password_requests(status);

-- 6. Chat Messages
CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    user_type VARCHAR(15) NOT NULL CHECK (user_type IN ('admin', 'pentester', 'stakeholder')),
    role VARCHAR(15) NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_chat_user_id ON chat_messages(user_id);
CREATE INDEX idx_chat_user_type ON chat_messages(user_type);
CREATE INDEX idx_chat_created_at ON chat_messages(created_at);
CREATE INDEX idx_chat_user_created ON chat_messages(user_id, created_at DESC);

-- 6.5 Token Blacklist (for logout/session invalidation)
CREATE TABLE IF NOT EXISTS token_blacklist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    user_type VARCHAR(15) NOT NULL CHECK (user_type IN ('admin', 'pentester', 'stakeholder')),
    blacklisted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_token_blacklist_user ON token_blacklist(user_id);
CREATE INDEX idx_token_blacklist_expires ON token_blacklist(expires_at);

-- 7. Project Management
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active','paused','completed','archived')),
    start_date DATE,
    end_date DATE,
    created_by UUID REFERENCES admin(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_projects_created_by ON projects(created_by);
CREATE INDEX idx_projects_status ON projects(status);

CREATE TABLE IF NOT EXISTS project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(15) NOT NULL CHECK (role IN ('pentester','stakeholder')),
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, user_id)
);

CREATE INDEX idx_pm_project ON project_members(project_id);
CREATE INDEX idx_pm_user ON project_members(user_id);

-- 8. Task Management
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    assigned_to UUID REFERENCES pentesters(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    priority VARCHAR(10) CHECK (priority IN ('low','medium','high','critical')),
    status VARCHAR(20) NOT NULL DEFAULT 'assigned'
        CHECK (status IN (
            'assigned',
            'in_progress',
            'submitted',
            'review',
            'closed'
        )),
    deadline TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tasks_project ON tasks(project_id);
CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX idx_tasks_status ON tasks(status);

CREATE TABLE IF NOT EXISTS task_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    pentester_id UUID REFERENCES pentesters(id),
    notes TEXT,
    submission_file TEXT,
    submission_file_path VARCHAR(500),
    status VARCHAR(20)
        CHECK (status IN ('submitted','changes_requested','approved')),
    reviewed_by UUID ,
    review_notes TEXT,
    reviewed_at TIMESTAMP,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_task_submissions_task ON task_submissions(task_id);
CREATE INDEX idx_task_submissions_pentester ON task_submissions(pentester_id);

-- 8.1 Task Submission Attachments
CREATE TABLE IF NOT EXISTS task_submission_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id UUID NOT NULL REFERENCES task_submissions(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size INT NOT NULL,
    notes TEXT,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_task_attachments_submission ON task_submission_attachments(submission_id);

-- 8.2 Task Submission Reviews (for admin and stakeholder feedback)
CREATE TABLE IF NOT EXISTS task_submission_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id UUID NOT NULL REFERENCES task_submissions(id) ON DELETE CASCADE,
    reviewer_id UUID NOT NULL,
    reviewer_role VARCHAR(15) NOT NULL CHECK (reviewer_role IN ('admin', 'stakeholder')),
    status VARCHAR(20) NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'completed')),
    review_notes TEXT,
    reviewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_task_submission_reviews_submission ON task_submission_reviews(submission_id);
CREATE INDEX idx_task_submission_reviews_reviewer ON task_submission_reviews(reviewer_id);

-- 9. Vulnerability Management
CREATE TABLE IF NOT EXISTS vulnerabilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    reported_by UUID ,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(10)
        CHECK (severity IN ('low','medium','high','critical')),
    cvss_score NUMERIC(3,1),
    cwe_id VARCHAR(20),
    owasp_category VARCHAR(100),
    affected_asset TEXT,
    status VARCHAR(25) NOT NULL DEFAULT 'new'
        CHECK (status IN (
            'new',
            'triaged',
            'assigned',
            'in_progress',
            'retesting',
            'verified',
            'resolved',
            'closed'
        )),
    remediation TEXT,
    evidence TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_vuln_project ON vulnerabilities(project_id);
CREATE INDEX idx_vuln_task ON vulnerabilities(task_id);
CREATE INDEX idx_vuln_reported_by ON vulnerabilities(reported_by);
CREATE INDEX idx_vuln_status ON vulnerabilities(status);
CREATE INDEX idx_vuln_severity ON vulnerabilities(severity);

-- 10. Real-Time Messaging System
-- Messages (1-to-1 Chat)
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id UUID NOT NULL ,
    receiver_id UUID NOT NULL ,
    content TEXT NOT NULL,
    status VARCHAR(15) NOT NULL DEFAULT 'sent' CHECK (status IN ('sent', 'delivered', 'read')),
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    delivered_at TIMESTAMP,
    read_at TIMESTAMP
);

CREATE INDEX idx_messages_sender ON messages(sender_id);
CREATE INDEX idx_messages_receiver ON messages(receiver_id);
CREATE INDEX idx_messages_status ON messages(status);
CREATE INDEX idx_messages_sent_at ON messages(sent_at DESC);
CREATE INDEX idx_messages_conversation ON messages(sender_id, receiver_id, sent_at DESC);

-- Message Attachments
CREATE TABLE IF NOT EXISTS message_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size INT NOT NULL,
    mime_type VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_msg_attachments_message ON message_attachments(message_id);

-- 11. Notification System
CREATE TABLE IF NOT EXISTS notifications (
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
);

CREATE INDEX idx_notifications_sender ON notifications(sender_id);
CREATE INDEX idx_notifications_priority ON notifications(priority);
CREATE INDEX idx_notifications_created ON notifications(created_at DESC);

-- Notification Delivery Logs (tracks email, in-app delivery per user/notification)
CREATE TABLE IF NOT EXISTS notification_delivery_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    receiver_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel VARCHAR(20) NOT NULL CHECK (channel IN ('in_app', 'email')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'delivered', 'failed', 'bounced')),
    attempted_at TIMESTAMP,
    delivered_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notif_logs_notification ON notification_delivery_logs(notification_id);
CREATE INDEX idx_notif_logs_receiver ON notification_delivery_logs(receiver_id);
CREATE INDEX idx_notif_logs_channel ON notification_delivery_logs(channel);
CREATE INDEX idx_notif_logs_status ON notification_delivery_logs(status);

-- 12. Activity/Audit Log System
CREATE TABLE IF NOT EXISTS activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255),
    activity_type VARCHAR(50) NOT NULL,
    description TEXT,
    metadata JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_activities_user_id ON activities(user_id);
CREATE INDEX idx_activities_activity_type ON activities(activity_type);
CREATE INDEX idx_activities_created_at ON activities(created_at DESC);
CREATE INDEX idx_activities_user_created ON activities(user_id, created_at DESC);