package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FileService handles file upload and retrieval operations
type FileService struct {
	resumeDir         string
	attachmentDir     string
	taskSubmissionDir string
}

// NewFileService creates a new file service with specified directories
func NewFileService(resumeDir, attachmentDir string) *FileService {
	// Ensure directories exist
	os.MkdirAll(resumeDir, 0755)
	os.MkdirAll(attachmentDir, 0755)

	// Add task submission directory
	taskSubmissionDir := filepath.Join("storage", "task-submissions")
	os.MkdirAll(taskSubmissionDir, 0755)

	return &FileService{
		resumeDir:         resumeDir,
		attachmentDir:     attachmentDir,
		taskSubmissionDir: taskSubmissionDir,
	}
}

// UploadResume saves a resume file and returns the file path
func (fs *FileService) UploadResume(fileHeader *os.File, fileName string) (string, error) {
	return fs.uploadFile(fileHeader, fileName, fs.resumeDir, "resume")
}

// UploadAttachment saves an attachment file and returns the file path
func (fs *FileService) UploadAttachment(fileHeader *os.File, fileName string) (string, error) {
	return fs.uploadFile(fileHeader, fileName, fs.attachmentDir, "attachment")
}

// UploadTaskSubmission saves a task submission file and returns the file path
func (fs *FileService) UploadTaskSubmission(fileHeader *os.File, fileName string) (string, error) {
	return fs.uploadFile(fileHeader, fileName, fs.taskSubmissionDir, "submission")
}

// uploadFile is a helper method that saves a file to the specified directory
func (fs *FileService) uploadFile(fileHeader *os.File, originalFileName string, targetDir string, fileType string) (string, error) {
	// Generate unique filename to avoid conflicts
	ext := filepath.Ext(originalFileName)
	uniqueName := fmt.Sprintf("%s_%s_%d%s", fileType, uuid.New().String()[:8], time.Now().Unix(), ext)
	filePath := filepath.Join(targetDir, uniqueName)

	fmt.Printf("[DEBUG] uploadFile - Original: %s, Target: %s, Type: %s\n", originalFileName, filePath, fileType)

	// Create destination file
	destFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer destFile.Close()

	// Copy content from source to destination
	bytesWritten, err := io.Copy(destFile, fileHeader)
	if err != nil {
		os.Remove(filePath) // Clean up on error
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Printf("[DEBUG] uploadFile - Wrote %d bytes to %s\n", bytesWritten, filePath)

	// Verify file was written
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to verify file: %w", err)
	}
	fmt.Printf("[DEBUG] uploadFile - File verified: %d bytes on disk\n", fileInfo.Size())

	return filePath, nil
}

// DeleteResume removes a resume file from storage
func (fs *FileService) DeleteResume(filePath string) error {
	return fs.deleteFile(filePath, fs.resumeDir)
}

// DeleteAttachment removes an attachment file from storage
func (fs *FileService) DeleteAttachment(filePath string) error {
	return fs.deleteFile(filePath, fs.attachmentDir)
}

// DeleteTaskSubmission removes a task submission file from storage
func (fs *FileService) DeleteTaskSubmission(filePath string) error {
	return fs.deleteFile(filePath, fs.taskSubmissionDir)
}

// deleteFile is a helper method that removes a file from the specified directory
func (fs *FileService) deleteFile(filePath string, expectedDir string) error {
	// Validate that file is within the expected directory (security check)
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	absExpectedDir, err := filepath.Abs(expectedDir)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(absPath, absExpectedDir) {
		return fmt.Errorf("file path is outside allowed directory")
	}

	if err := os.Remove(absPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// ReadFile reads and returns the content of a file
func (fs *FileService) ReadFile(filePath string) ([]byte, error) {
	// Validate that file is within one of the allowed directories
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	absResumeDir, err := filepath.Abs(fs.resumeDir)
	if err != nil {
		return nil, err
	}
	absAttachmentDir, err := filepath.Abs(fs.attachmentDir)
	if err != nil {
		return nil, err
	}
	absTaskSubmissionDir, err := filepath.Abs(fs.taskSubmissionDir)
	if err != nil {
		return nil, err
	}

	isInResumeDir := strings.HasPrefix(absPath, absResumeDir)
	isInAttachmentDir := strings.HasPrefix(absPath, absAttachmentDir)
	isInTaskSubmissionDir := strings.HasPrefix(absPath, absTaskSubmissionDir)

	if !isInResumeDir && !isInAttachmentDir && !isInTaskSubmissionDir {
		return nil, fmt.Errorf("file path is outside allowed directories")
	}

	return os.ReadFile(absPath)
}

// FileExists checks if a file exists
func (fs *FileService) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
