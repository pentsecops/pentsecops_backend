package services

import (
	"fmt"
	"log"
	"strings"
)

// KnowledgeBase stores and retrieves Pentsecops platform information
type KnowledgeBase struct {
	content string
	topics  map[string]string // keyword -> answer mapping
}

// NewKnowledgeBase creates and initializes the knowledge base
func NewKnowledgeBase() *KnowledgeBase {
	kb := &KnowledgeBase{
		topics: make(map[string]string),
	}

	// Initialize with Pentsecops information
	kb.initializePentsecopsKnowledge()
	return kb
}

// initializePentsecopsKnowledge sets up the knowledge base with platform info
func (kb *KnowledgeBase) initializePentsecopsKnowledge() {
	kb.content = `
PENTSECOPS PLATFORM OVERVIEW:

Pentsecops is a unified, comprehensive platform designed for seamless execution of security operations 
within an organization. It acts as a centralized hub where security professionals can efficiently manage, 
track, and collaborate on various security activities.

KEY ROLES AND RESPONSIBILITIES:

1. ADMIN ROLE:
The Admin is the cornerstone of Pentsecops, possessing full control over all aspects of the system.
- User Management: Create, modify, and deactivate user accounts across all roles
- Project Management: Create, assign, and oversee security project lifecycle
- Task Management: Create, assign, and track security tasks effectively
- Vulnerability Management: Identify, prioritize, and manage security vulnerabilities
- Overall System Management: Maintain platform security and smooth operation

2. PENTESTER ROLE:
The Pentester plays a vital role in carrying out hands-on security operations.
- Task Execution: Execute assigned security tasks (pen testing, vulnerability assessments, audits)
- Vulnerability Discovery & Exploitation: Identify and exploit security weaknesses
- Documentation: Document findings, methodologies, and remediation steps
- Collaboration: Work with Admins and Stakeholders to communicate results

3. STAKEHOLDER ROLE:
The Stakeholder role provides visibility into ongoing security operations without direct participation.
- Project Monitoring: View progress of ongoing security projects
- Task Updates: Stay updated with assigned tasks and completion statuses
- Vulnerability Awareness: Informed about identified vulnerabilities and remediation efforts
- Reporting: Receive regular reports and updates on security operations

PLATFORM WORKFLOW:

Admin Responsibilities:
- Creates and manages projects, tasks, and users
- Assigns specific security tasks to Pentesters based on skill sets
- Reviews documentation submitted by Pentesters
- Monitors and tracks progress, making adjustments

Pentester Responsibilities:
- Executes security tests and identifies vulnerabilities
- Communicates findings, risks, and remediation suggestions
- Submits well-documented reports with discovered vulnerabilities
- Works within defined timeframes to meet project deadlines

Stakeholder Responsibilities:
- Reviews high-level project overviews and task progress
- Monitors organization's security posture
- Ensures security efforts align with business goals

CORE BENEFITS:
- Centralized Management: All security operations in single platform
- Role-Based Access: Each role has appropriate access and tools
- Clear Communication: Real-time updates and streamlined documentation
- Comprehensive Security Oversight: Organized, efficient security operations

KEY FEATURES:
- Real-time project and task tracking
- Vulnerability identification and management
- Role-based access control (RBAC)
- Document submission and review workflows
- Stakeholder reporting and visibility
- Team collaboration tools
- Security audit trails
`

	// Initialize topic keywords for quick matching
	kb.topics = map[string]string{
		"admin":           "The Admin role has full control over Pentsecops platform management including user management, project/task creation, vulnerability management, and overall system oversight.",
		"pentester":       "The Pentester role executes security tasks, identifies vulnerabilities, documents findings, and collaborates with Admins and Stakeholders to communicate security results.",
		"stakeholder":     "The Stakeholder role provides visibility into security operations, monitors projects, stays updated on tasks and vulnerabilities, and ensures alignment with business goals.",
		"roles":           "Pentsecops has three roles: Admin (full control), Pentester (hands-on operations), and Stakeholder (visibility/oversight).",
		"user management": "Admins manage users by creating, modifying, and deactivating accounts across all three roles (Admin, Pentester, Stakeholder).",
		"project":         "Projects are created by Admins to organize security operations. They can be assigned to Pentesters and monitored by Stakeholders.",
		"task":            "Tasks are specific security assignments created by Admins and assigned to Pentesters for execution. Stakeholders can monitor task progress.",
		"vulnerability":   "Vulnerabilities are identified by Pentesters during testing. Admins manage and prioritize them, while Stakeholders are informed about remediation efforts.",
		"workflow":        "Admins create and assign tasks → Pentesters execute and document → Stakeholders monitor and stay informed.",
		"pentsecops":      "Pentsecops is a unified platform for security operations providing centralized management, role-based access, and clear communication.",
		"features":        "Key features include real-time project tracking, vulnerability management, RBAC, document workflows, stakeholder reporting, collaboration tools, and audit trails.",
	}

	log.Printf("[INFO] KnowledgeBase: Initialized with Pentsecops platform information")
}

// SearchKnowledgeBase searches for relevant information based on a query
func (kb *KnowledgeBase) SearchKnowledgeBase(query string) (string, bool) {
	queryLower := strings.ToLower(query)

	// Check for exact keyword matches first (more specific answers)
	for keyword, answer := range kb.topics {
		if strings.Contains(queryLower, keyword) {
			log.Printf("[INFO] KnowledgeBase: Found topic match for keyword '%s'", keyword)
			return answer, true
		}
	}

	// Check if query is about Pentsecops at all
	pentsecopsKeywords := []string{"pentsecops", "platform", "role", "admin", "pentester", "stakeholder",
		"project", "task", "vulnerability", "user", "workflow", "feature", "management"}

	for _, keyword := range pentsecopsKeywords {
		if strings.Contains(queryLower, keyword) {
			log.Printf("[INFO] KnowledgeBase: Detected Pentsecops-related query")
			return kb.getGeneralResponse(query), true
		}
	}

	return "", false
}

// getGeneralResponse returns a general response about Pentsecops
func (kb *KnowledgeBase) getGeneralResponse(query string) string {
	response := fmt.Sprintf(`Based on Pentsecops platform knowledge:

%s

For more specific information about Pentsecops operations, roles, and features, please refer to the platform documentation.`, kb.content)
	return response
}

// IsPentsecopsQuestion determines if the question is about Pentsecops
func (kb *KnowledgeBase) IsPentsecopsQuestion(query string) bool {
	queryLower := strings.ToLower(query)

	pentsecopsKeywords := []string{
		"pentsecops", "platform", "admin role", "pentester role", "stakeholder role",
		"project management", "task assignment", "vulnerability", "security operations",
		"role-based", "rbac", "user management", "workflow", "collaboration",
	}

	for _, keyword := range pentsecopsKeywords {
		if strings.Contains(queryLower, keyword) {
			return true
		}
	}

	return false
}

// GetDetailedInformation returns detailed information about a specific topic
func (kb *KnowledgeBase) GetDetailedInformation(topic string) (string, error) {
	topicLower := strings.ToLower(topic)

	if answer, exists := kb.topics[topicLower]; exists {
		return answer, nil
	}

	return "", fmt.Errorf("topic '%s' not found in knowledge base", topic)
}
