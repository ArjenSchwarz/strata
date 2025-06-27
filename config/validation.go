package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	format "github.com/ArjenSchwarz/go-output"
)

// FileValidator provides validation functionality for file output settings
type FileValidator struct {
	config *Config
}

// NewFileValidator creates a new FileValidator instance
func NewFileValidator(config *Config) *FileValidator {
	return &FileValidator{config: config}
}

// ValidateFileOutput performs comprehensive validation of file output settings
func (fv *FileValidator) ValidateFileOutput(settings *format.OutputSettings) error {
	if settings.OutputFile == "" {
		return nil // No file output, nothing to validate
	}

	// Validate file path safety
	if err := fv.validatePathSafety(settings.OutputFile); err != nil {
		return fmt.Errorf("file path validation failed: %w", err)
	}

	// Validate directory permissions
	if err := fv.validateDirectoryPermissions(settings.OutputFile); err != nil {
		return fmt.Errorf("directory permission validation failed: %w", err)
	}

	// Validate format support
	if err := fv.validateFormatSupport(settings.OutputFileFormat); err != nil {
		return fmt.Errorf("format validation failed: %w", err)
	}

	return nil
}

// sanitizeFilePath cleans and validates a file path for security.
// Returns the cleaned absolute path or a structured error for security violations.
func (fv *FileValidator) sanitizeFilePath(path string) (string, error) {
	// Check for path traversal attempts before cleaning
	if strings.Contains(path, "..") {
		return "", &FileOutputError{
			Type:    "validation",
			Code:    "PATH_TRAVERSAL",
			Path:    path,
			Message: "path traversal not allowed",
		}
	}

	// Clean path and resolve any relative components
	clean := filepath.Clean(path)

	// Convert to absolute path for consistency
	abs, err := filepath.Abs(clean)
	if err != nil {
		return "", &FileOutputError{
			Type:    "validation",
			Code:    "INVALID_PATH",
			Path:    path,
			Message: "invalid file path",
			Cause:   err,
		}
	}

	return abs, nil
}

// validatePathSafety ensures the file path is safe and doesn't contain traversal attempts.
// Examples of blocked paths: "../../../etc/passwd", "reports/../../../sensitive"
// Examples of allowed paths: "output.json", "reports/2025/summary.txt"
func (fv *FileValidator) validatePathSafety(filePath string) error {
	_, err := fv.sanitizeFilePath(filePath)
	return err
}

// validateDirectoryPermissions checks if the directory exists and is writable.
// Uses efficient permission checking without creating temporary files when possible.
func (fv *FileValidator) validateDirectoryPermissions(filePath string) error {
	dir := filepath.Dir(filePath)

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return &FileOutputError{
			Type:    "permission",
			Code:    "DIRECTORY_NOT_FOUND",
			Path:    dir,
			Message: "directory does not exist",
			Cause:   err,
		}
	}

	// Test write permissions using os.OpenFile with O_WRONLY for better cross-platform compatibility
	testFile := filepath.Join(dir, ".strata_write_test")
	file, err := os.OpenFile(testFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return &FileOutputError{
			Type:    "permission",
			Code:    "PERMISSION_DENIED",
			Path:    dir,
			Message: "no write permission for directory",
			Cause:   err,
		}
	}
	file.Close()
	os.Remove(testFile)

	return nil
}

// validateFormatSupport checks if the specified output format is supported.
// Supported formats include: table, json, csv, markdown, html, dot
func (fv *FileValidator) validateFormatSupport(formatName string) error {
	supportedFormats := []string{
		"table",
		"json",
		"csv",
		"markdown",
		"html",
		"dot",
	}

	formatLower := strings.ToLower(formatName)
	if !slices.Contains(supportedFormats, formatLower) {
		return &FileOutputError{
			Type:    "format",
			Code:    "UNSUPPORTED_FORMAT",
			Path:    "",
			Format:  formatName,
			Message: fmt.Sprintf("unsupported output format: %s, supported formats: %v", formatName, supportedFormats),
		}
	}

	return nil
}

// FileOutputError represents errors that occur during file output operations
type FileOutputError struct {
	Type    string // "validation", "permission", "format", "write"
	Code    string // e.g., "PATH_TRAVERSAL", "PERMISSION_DENIED", "UNSUPPORTED_FORMAT"
	Path    string
	Format  string
	Message string
	Cause   error
}

func (e *FileOutputError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s error [%s] for %s: %s (caused by: %v)",
			e.Type, e.Code, e.Path, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s error [%s] for %s: %s", e.Type, e.Code, e.Path, e.Message)
}

// ValidationResult holds the result of validation operations
type ValidationResult struct {
	Valid    bool
	Errors   []error
	Warnings []string
	Infos    []string
}

func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

func (vr *ValidationResult) AddError(err error) {
	vr.Errors = append(vr.Errors, err)
	vr.Valid = false
}

func (vr *ValidationResult) AddWarning(warning string) {
	vr.Warnings = append(vr.Warnings, warning)
}

func (vr *ValidationResult) AddInfo(info string) {
	vr.Infos = append(vr.Infos, info)
}

// checkFileOverwrite checks if a file already exists and adds a warning if it does
func (fv *FileValidator) checkFileOverwrite(filePath string, result *ValidationResult) {
	if _, err := os.Stat(filePath); err == nil {
		result.AddWarning(fmt.Sprintf("file %s already exists and will be overwritten", filePath))
	}
}

// ValidateAll performs comprehensive validation and returns detailed results
func (fv *FileValidator) ValidateAll(settings *format.OutputSettings) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if settings.OutputFile == "" {
		return result // No file output, nothing to validate
	}

	// Validate path safety
	if err := fv.validatePathSafety(settings.OutputFile); err != nil {
		result.AddError(err)
	}

	// Validate directory permissions
	if err := fv.validateDirectoryPermissions(settings.OutputFile); err != nil {
		result.AddError(err)
	}

	// Validate format support
	if err := fv.validateFormatSupport(settings.OutputFileFormat); err != nil {
		result.AddError(err)
	}

	// Check for file overwrite scenarios
	fv.checkFileOverwrite(settings.OutputFile, result)

	return result
}
