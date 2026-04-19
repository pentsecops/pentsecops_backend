package domain

// UserRepository defines the contract for user data operations
type UserRepository interface {
	CreateUser(user *User) error
	GetUserByID(id string) (*User, error)
	GetAnyUserByID(id string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateUser(user *User) error
	UpdateUserLastLogin(userID string) error
	UpdateUserStatus(userID string, status string) error
	DeleteUser(id string) error
	GetAllUsers() ([]*User, error)
	GetAllUsersPaginated(page, limit int) ([]*User, int64, error)
	CountAllUsers() (int64, error)
	GetUserRole(userID string) (string, error)
}

// AdminRepository defines the contract for admin data operations
type AdminRepository interface {
	CreateAdmin(admin *Admin) error
	GetAdminByID(id string) (*Admin, error)
	GetAdminByEmail(email string) (*Admin, error)
	UpdateAdmin(admin *Admin) error
}

// PentesterRepository defines the contract for pentester profile operations
type PentesterRepository interface {
	CreatePentester(pentester *Pentester) error
	GetPentesterByID(id string) (*Pentester, error)
	GetPentesterByUserID(userID string) (*Pentester, error)
	GetPentesterByEmail(email string) (*Pentester, error)
	UpdatePentester(pentester *Pentester) error
}

// StakeholderRepository defines the contract for stakeholder profile operations
type StakeholderRepository interface {
	CreateStakeholder(stakeholder *Stakeholder) error
	GetStakeholderByID(id string) (*Stakeholder, error)
	GetStakeholderByUserID(userID string) (*Stakeholder, error)
	GetStakeholderByEmail(email string) (*Stakeholder, error)
	UpdateStakeholder(stakeholder *Stakeholder) error
}

// ForgotPasswordRequestRepository defines the contract for forgot password request operations
type ForgotPasswordRequestRepository interface {
	CreateRequest(userID, email string) (*ForgotPasswordRequest, error)
	GetPendingRequests() ([]ForgotPasswordRequest, error)
	GetRequestByID(requestID string) (*ForgotPasswordRequest, error)
	UpdateRequestStatus(requestID, status, tempPassword string) error
	RejectRequest(requestID string) error
	GetRequestByEmail(email string) (*ForgotPasswordRequest, error)
}

// ProjectRepository defines the contract for project data operations
type ProjectRepository interface {
	CreateProject(project *Project) error
	GetProjectByID(id string) (*Project, error)
	GetAllProjects() ([]*Project, error)
	GetProjectsByAdmin(adminID string) ([]*Project, error)
	UpdateProject(project *Project) error
	DeleteProject(id string) error
}

// ProjectMemberRepository defines the contract for project member operations
type ProjectMemberRepository interface {
	AddMember(member *ProjectMember) error
	GetProjectMembers(projectID string) ([]*ProjectMember, error)
	GetUserProjects(userID string) ([]*ProjectMember, error)
	GetMemberRole(projectID, userID string) (string, error)
	RemoveMember(projectID, userID string) error
	IsMemberOfProject(projectID, userID string) (bool, error)
}

// TaskRepository defines the contract for task data operations
type TaskRepository interface {
	CreateTask(task *Task) error
	GetTaskByID(id string) (*Task, error)
	GetTasksByProject(projectID string) ([]*Task, error)
	GetTasksByPentester(pentesterID string) ([]*Task, error)
	GetTasksByStatus(status string) ([]*Task, error)
	UpdateTask(task *Task) error
	DeleteTask(id string) error
	UpdateTaskStatus(taskID, status string) error
}

// TaskSubmissionRepository defines the contract for task submission data operations
type TaskSubmissionRepository interface {
	CreateSubmission(submission *TaskSubmission) error
	GetSubmissionByID(id string) (*TaskSubmission, error)
	GetSubmissionsByTask(taskID string) ([]*TaskSubmission, error)
	DeleteSubmission(id string) error
}

// TaskSubmissionAttachmentRepository defines the contract for task submission attachment operations
type TaskSubmissionAttachmentRepository interface {
	CreateAttachment(attachment *TaskSubmissionAttachment) error
	GetAttachmentByID(id string) (*TaskSubmissionAttachment, error)
	GetAttachmentsBySubmission(submissionID string) ([]*TaskSubmissionAttachment, error)
	DeleteAttachment(id string) error
}

// TaskSubmissionReviewRepository defines the contract for task submission review operations
type TaskSubmissionReviewRepository interface {
	CreateReview(review *TaskSubmissionReview) error
	GetReviewsBySubmission(submissionID string) ([]*TaskSubmissionReview, error)
	GetReviewByID(id string) (*TaskSubmissionReview, error)
	UpdateReview(review *TaskSubmissionReview) error
	DeleteReview(id string) error
}

// VulnerabilityRepository defines the contract for vulnerability data operations
type VulnerabilityRepository interface {
	CreateVulnerability(vulnerability *Vulnerability) error
	GetVulnerabilityByID(id string) (*Vulnerability, error)
	GetVulnerabilitiesByProject(projectID string) ([]*Vulnerability, error)
	GetVulnerabilitiesByTask(taskID string) ([]*Vulnerability, error)
	GetVulnerabilitiesByPentester(pentesterID string) ([]*Vulnerability, error)
	GetVulnerabilitiesBySeverity(severity string) ([]*Vulnerability, error)
	GetVulnerabilitiesByStatus(status string) ([]*Vulnerability, error)
	UpdateVulnerability(vulnerability *Vulnerability) error
	DeleteVulnerability(id string) error
	UpdateVulnerabilityStatus(id, status string) error
}
