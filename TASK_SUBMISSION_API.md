# Task Submission API Documentation

## Overview
The Task Submission API allows pentesters to submit completed tasks with notes and file attachments, and enables admins/stakeholders to review submissions with status tracking.

---

## 📋 API Endpoints

### 1. Submit Task (Pentester)
**Endpoint**: `POST /api/v1/task-submission/:task_id`

**Authentication**: Required (Bearer Token - Pentester)

**Description**: Pentester submits a completed task with notes and optional file attachments.

#### Request Headers
```http
Authorization: Bearer {pentester_token}
Content-Type: multipart/form-data
```

#### Request Body (Form Data)
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `notes` | Text | Yes | Task completion notes/report |
| `attachments` | File[] | No | Multiple file uploads (max 50MB per file) |
| `attachment_notes` | Text[] | No | Optional notes for each attachment |

#### Example Request
```bash
curl -X POST http://localhost:8080/api/v1/task-submission/task-uuid-123 \
  -H "Authorization: Bearer eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9..." \
  -F "notes=Completed penetration test. Found 3 critical vulnerabilities in authentication module." \
  -F "attachments=@pentest_report.pdf" \
  -F "attachments=@vulnerability_details.xlsx" \
  -F "attachment_notes=Full detailed pentest report" \
  -F "attachment_notes=Vulnerability list with exploitation steps"
```

#### Response (201 Created)
```json
{
  "success": true,
  "message": "Task submitted successfully",
  "data": {
    "submission": {
      "id": "submission-uuid-1",
      "task_id": "task-uuid-123",
      "pentester_id": "pentester-uuid",
      "notes": "Completed penetration test. Found 3 critical vulnerabilities in authentication module.",
      "submitted_at": "2026-04-14T10:30:00Z"
    },
    "attachments": [
      {
        "id": "attachment-uuid-1",
        "submission_id": "submission-uuid-1",
        "file_name": "pentest_report.pdf",
        "file_path": "storage/task-submissions/submission_a1b2c3d4_1713097800.pdf",
        "file_size": 2048576,
        "notes": "Full detailed pentest report",
        "uploaded_at": "2026-04-14T10:30:00Z"
      },
      {
        "id": "attachment-uuid-2",
        "submission_id": "submission-uuid-1",
        "file_name": "vulnerability_details.xlsx",
        "file_path": "storage/task-submissions/submission_e5f6g7h8_1713097801.xlsx",
        "file_size": 1024576,
        "notes": "Vulnerability list with exploitation steps",
        "uploaded_at": "2026-04-14T10:30:01Z"
      }
    ],
    "reviews": []
  }
}
```

#### Error Responses

**401 Unauthorized** - Missing or invalid token
```json
{
  "success": false,
  "message": "Unauthorized",
  "error": "missing token"
}
```

**403 Forbidden** - Only pentesters can submit
```json
{
  "success": false,
  "message": "Forbidden",
  "error": "only pentesters can submit tasks"
}
```

**400 Bad Request** - File too large
```json
{
  "success": false,
  "message": "File too large",
  "error": "attachment 1 must be less than 50MB"
}
```

---

### 2. Get Task Submissions (All Roles)
**Endpoint**: `GET /api/v1/task-submission/:task_id`

**Authentication**: Required (Bearer Token - Any role)

**Description**: Retrieve all submissions for a specific task, including attachments and all reviews by different reviewers.

#### Request Headers
```http
Authorization: Bearer {token}
```

#### Response (200 OK)
```json
{
  "success": true,
  "message": "Submissions retrieved successfully",
  "data": [
    {
      "submission": {
        "id": "submission-uuid-1",
        "task_id": "task-uuid-123",
        "pentester_id": "pentester-uuid-1",
        "notes": "First submission attempt - initial findings",
        "submitted_at": "2026-04-14T09:00:00Z"
      },
      "attachments": [
        {
          "id": "attachment-uuid-1",
          "submission_id": "submission-uuid-1",
          "file_name": "initial_report.pdf",
          "file_path": "storage/task-submissions/submission_a1b2c3d4_1713093600.pdf",
          "file_size": 1048576,
          "notes": "Initial findings",
          "uploaded_at": "2026-04-14T09:00:00Z"
        }
      ],
      "reviews": [
        {
          "id": "review-uuid-1",
          "submission_id": "submission-uuid-1",
          "reviewer_id": "admin-uuid-1",
          "reviewer_role": "admin",
          "status": "in_progress",
          "review_notes": "More details needed on the vulnerability exploitation. Please provide proof-of-concept.",
          "reviewed_at": "2026-04-14T10:15:00Z"
        },
        {
          "id": "review-uuid-2",
          "submission_id": "submission-uuid-1",
          "reviewer_id": "stakeholder-uuid-1",
          "reviewer_role": "stakeholder",
          "status": "in_progress",
          "review_notes": "Good initial work. Expected completion by end of week.",
          "reviewed_at": "2026-04-14T10:20:00Z"
        }
      ]
    },
    {
      "submission": {
        "id": "submission-uuid-2",
        "task_id": "task-uuid-123",
        "pentester_id": "pentester-uuid-1",
        "notes": "Final submission with all findings and remediations",
        "submitted_at": "2026-04-14T11:00:00Z"
      },
      "attachments": [
        {
          "id": "attachment-uuid-3",
          "submission_id": "submission-uuid-2",
          "file_name": "final_report.pdf",
          "file_path": "storage/task-submissions/submission_i9j0k1l2_1713097200.pdf",
          "file_size": 2560000,
          "notes": "Comprehensive final report",
          "uploaded_at": "2026-04-14T11:00:00Z"
        }
      ],
      "reviews": [
        {
          "id": "review-uuid-3",
          "submission_id": "submission-uuid-2",
          "reviewer_id": "admin-uuid-1",
          "reviewer_role": "admin",
          "status": "completed",
          "review_notes": "Excellent work. All findings are documented properly. Approved for client presentation.",
          "reviewed_at": "2026-04-14T11:30:00Z"
        }
      ]
    }
  ]
}
```

#### Error Responses

**401 Unauthorized** - Missing or invalid token
```json
{
  "success": false,
  "message": "Unauthorized",
  "error": "missing token"
}
```

**400 Bad Request** - Task not found
```json
{
  "success": false,
  "message": "Failed to retrieve submissions",
  "error": "task not found"
}
```

---

### 3. Review Task Submission (Admin/Stakeholder)
**Endpoint**: `POST /api/v1/task-submission/:submission_id/review`

**Authentication**: Required (Bearer Token - Admin or Stakeholder only)

**Description**: Create a review record for a task submission. Each reviewer (admin/stakeholder) can create separate review entries with their own status and notes.

#### Request Headers
```http
Authorization: Bearer {admin_or_stakeholder_token}
Content-Type: application/json
```

#### Request Body (JSON)
```json
{
  "status": "in_progress|completed",
  "review_notes": "Your review comments here"
}
```

| Field | Type | Required | Values | Description |
|-------|------|----------|--------|-------------|
| `status` | String | Yes | `in_progress`, `completed` | Review status |
| `review_notes` | String | Yes | Any text | Comments/feedback on submission |

#### Example Request
```bash
curl -X POST http://localhost:8080/api/v1/task-submission/submission-uuid-1/review \
  -H "Authorization: Bearer eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "status": "completed",
    "review_notes": "Excellent work! All findings documented properly. Approved for client delivery."
  }'
```

#### Response (201 Created)
```json
{
  "success": true,
  "message": "Review created successfully",
  "data": {
    "id": "review-uuid-123",
    "submission_id": "submission-uuid-1",
    "reviewer_id": "admin-uuid-1",
    "reviewer_role": "admin",
    "status": "completed",
    "review_notes": "Excellent work! All findings documented properly. Approved for client delivery.",
    "reviewed_at": "2026-04-14T12:00:00Z"
  }
}
```

#### Error Responses

**401 Unauthorized** - Missing or invalid token
```json
{
  "success": false,
  "message": "Unauthorized",
  "error": "missing token"
}
```

**403 Forbidden** - Only admins and stakeholders can review
```json
{
  "success": false,
  "message": "Forbidden",
  "error": "only admins and stakeholders can review submissions"
}
```

**400 Bad Request** - Invalid request body
```json
{
  "success": false,
  "message": "Invalid request",
  "error": "status must be 'in_progress' or 'completed'"
}
```

**400 Bad Request** - Submission not found
```json
{
  "success": false,
  "message": "Failed to review submission",
  "error": "submission not found"
}
```

---

### 4. Download Task Submission Attachment (All Roles)
**Endpoint**: `GET /api/v1/task-submission/attachments/:attachment_id`

**Authentication**: Required (Bearer Token - Any role)

**Description**: Download a file attachment from a task submission as a binary blob.

#### Request Headers
```http
Authorization: Bearer {token}
```

#### URL Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `attachment_id` | String (UUID) | Yes | The unique ID of the attachment to download |

#### Example Request
```bash
curl -X GET http://localhost:8080/api/v1/task-submission/attachments/attachment-uuid-123 \
  -H "Authorization: Bearer eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9..." \
  -o pentest_report.pdf
```

#### Response (200 OK)
- **Content-Type**: `application/octet-stream`
- **Body**: Binary file content (blob)
- **Headers**: 
  - `Content-Disposition: attachment; filename=<original_filename>`
  - `Content-Length: <file_size_bytes>`

#### Error Responses

**401 Unauthorized** - Missing or invalid token
```json
{
  "error": "unauthorized"
}
```

**400 Bad Request** - Missing attachment ID
```json
{
  "error": "attachment_id is required"
}
```

**404 Not Found** - Attachment does not exist
```json
{
  "error": "attachment not found"
}
```

**500 Internal Server Error** - File read failure
```json
{
  "error": "failed to read file"
}
```

---

## 🔐 Role-Based Access Control

| Endpoint | Pentester | Admin | Stakeholder | User |
|----------|-----------|-------|-------------|------|
| POST `/task-submission/:task_id` | ✅ Submit | ❌ | ❌ | ❌ |
| GET `/task-submission/:task_id` | ✅ View | ✅ View | ✅ View | ✅ View |
| POST `/task-submission/:submission_id/review` | ❌ | ✅ Review | ✅ Review | ❌ |
| GET `/task-submission/attachments/:attachment_id` | ✅ Download | ✅ Download | ✅ Download | ✅ Download |

---

## 📁 File Storage

### Storage Location
- **Base Directory**: `storage/task-submissions/`
- **File Naming Pattern**: `submission_{uuid}_{timestamp}{extension}`
- **Example**: `submission_a1b2c3d4_1713097800.pdf`

### File Constraints
- **Maximum File Size**: 50 MB per file
- **Maximum Attachments**: Unlimited (within system resources)
- **Supported Formats**: All formats (validation done client-side)

---

## 🔄 Workflow Example

### Step 1: Submit Task (Pentester)
```bash
# Pentester submits completed task with files
POST /api/v1/task-submission/task-123
- Notes: "Completed penetration test"
- Files: [report.pdf, vulnerabilities.xlsx]
- Attachment Notes: ["Main report", "Vulnerability details"]
```

### Step 2: Retrieve Submissions (Admin/Stakeholder)
```bash
# Admin views all submissions for the task
GET /api/v1/task-submission/task-123
- Returns: All submissions, attachments, and existing reviews
```

### Step 3: Admin Review
```bash
# Admin creates review record
POST /api/v1/task-submission/submission-123/review
- Status: "in_progress"
- Notes: "Needs more details on x,y,z"
```

### Step 4: Stakeholder Review
```bash
# Stakeholder creates separate review record
POST /api/v1/task-submission/submission-123/review
- Status: "completed"
- Notes: "Satisfied with findings. Ready for client."
```

### Step 5: Get Final Status
```bash
# Anyone can see all reviews
GET /api/v1/task-submission/task-123
- Returns: Submission with BOTH admin and stakeholder reviews
```

---

## 📊 Database Schema

### task_submissions
```sql
- id (UUID PRIMARY KEY)
- task_id (UUID, FK → tasks)
- pentester_id (UUID, FK → pentesters)
- notes (TEXT)
- submitted_at (TIMESTAMP)
```

### task_submission_attachments
```sql
- id (UUID PRIMARY KEY)
- submission_id (UUID, FK → task_submissions)
- file_name (VARCHAR)
- file_path (VARCHAR)
- file_size (BIGINT)
- notes (TEXT, optional)
- uploaded_at (TIMESTAMP)
```

### task_submission_reviews
```sql
- id (UUID PRIMARY KEY)
- submission_id (UUID, FK → task_submissions)
- reviewer_id (UUID)
- reviewer_role (VARCHAR: 'admin' or 'stakeholder')
- status (VARCHAR: 'in_progress' or 'completed')
- review_notes (TEXT, optional)
- reviewed_at (TIMESTAMP)
```

---

## 🛠️ Implementation Details

### Technology Stack
- **Framework**: GoFiber v2
- **Database**: PostgreSQL with UUID extensions
- **Authentication**: PASETO tokens
- **File Service**: Native Go file operations

### Key Features
✅ Multi-file upload support
✅ Role-based access control
✅ Separate review tracking per reviewer
✅ Automatic file cleanup on errors
✅ File size validation
✅ Comprehensive error handling
✅ Activity logging
