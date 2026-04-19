package services

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strings"

	"github.com/ledongthuc/pdf"
)

// DocumentParser handles document parsing for various file types
type DocumentParser struct{}

// NewDocumentParser creates a new document parser
func NewDocumentParser() *DocumentParser {
	return &DocumentParser{}
}

// ParseDocument extracts text from uploaded files (PDF, DOCX, TXT)
func (dp *DocumentParser) ParseDocument(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Determine file type by extension
	fileName := fileHeader.Filename
	extension := fileName[len(fileName)-4:]
	if len(fileName) > 5 {
		extension = fileName[len(fileName)-5:]
	}

	log.Printf("[INFO] DocumentParser: Parsing file %s (extension: %s)", fileName, extension)

	switch {
	case strings.HasSuffix(strings.ToLower(fileName), ".pdf"):
		return dp.parsePDF(file)
	case strings.HasSuffix(strings.ToLower(fileName), ".txt"):
		return dp.parseTXT(file)
	case strings.HasSuffix(strings.ToLower(fileName), ".docx"):
		return dp.parseDOCX(file)
	default:
		return "", fmt.Errorf("unsupported file type: %s", fileName)
	}
}

// parsePDF extracts text from PDF files
func (dp *DocumentParser) parsePDF(file io.Reader) (string, error) {
	log.Printf("[INFO] DocumentParser: Extracting text from PDF")

	// Read file into memory
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read PDF file: %w", err)
	}

	// Parse PDF
	pdfReader, err := pdf.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to parse PDF: %w", err)
	}

	// Extract text from all pages
	var extractedText strings.Builder
	pageCount := pdfReader.NumPage()

	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		page := pdfReader.Page(pageNum)

		text, err := page.GetPlainText(nil)
		if err != nil {
			log.Printf("[WARN] DocumentParser: Failed to extract text from page %d: %v", pageNum, err)
			continue
		}

		extractedText.WriteString(text)
		extractedText.WriteString("\n")
	}

	result := strings.TrimSpace(extractedText.String())
	if result == "" {
		return "", fmt.Errorf("no text could be extracted from PDF")
	}

	log.Printf("[SUCCESS] DocumentParser: Extracted %d characters from PDF", len(result))
	return result, nil
}

// parseTXT extracts text from plain text files
func (dp *DocumentParser) parseTXT(file io.Reader) (string, error) {
	log.Printf("[INFO] DocumentParser: Extracting text from TXT")

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read TXT file: %w", err)
	}

	text := string(fileBytes)
	if text == "" {
		return "", fmt.Errorf("TXT file is empty")
	}

	log.Printf("[SUCCESS] DocumentParser: Extracted %d characters from TXT", len(text))
	return text, nil
}

// parseDOCX extracts text from DOCX files (Office Open XML)
func (dp *DocumentParser) parseDOCX(file io.Reader) (string, error) {
	log.Printf("[INFO] DocumentParser: Extracting text from DOCX")

	// For DOCX, we need to extract from the document.xml within the zip archive
	// This is a simplified implementation
	// For production, consider using github.com/unidoc/unidoc or similar

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read DOCX file: %w", err)
	}

	// Extract text by looking for text tags in the XML
	text := extractDOCXText(fileBytes)
	if text == "" {
		return "", fmt.Errorf("no text could be extracted from DOCX")
	}

	log.Printf("[SUCCESS] DocumentParser: Extracted %d characters from DOCX", len(text))
	return text, nil
}

// extractDOCXText performs basic DOCX text extraction
func extractDOCXText(data []byte) string {
	// DOCX is a ZIP file containing XML files
	// This is a basic implementation that searches for text patterns

	content := string(data)

	// Find and extract text from <w:t> tags (Word text elements)
	var extractedText strings.Builder
	start := 0

	for {
		// Find opening tag
		idx := strings.Index(content[start:], "<w:t")
		if idx == -1 {
			break
		}

		// Find closing bracket of tag
		tagEnd := strings.Index(content[start+idx:], ">")
		if tagEnd == -1 {
			break
		}

		// Find text content
		textStart := start + idx + tagEnd + 1
		textEnd := strings.Index(content[textStart:], "</w:t>")
		if textEnd == -1 {
			break
		}

		text := content[textStart : textStart+textEnd]
		extractedText.WriteString(text)
		extractedText.WriteString(" ")

		start = textStart + textEnd + 6
	}

	return strings.TrimSpace(extractedText.String())
}
