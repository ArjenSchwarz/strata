package config

import (
	"fmt"
	"os"
	"path/filepath"
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

// sanitizeFilePath cleans and validates a file path for security
func (fv *FileValidator) sanitizeFilePath(path string) (string, error) {
	// Clean path and resolve any relative components
	clean := filepath.Clean(path)

	// Check for path traversal attempts
	if strings.Contains(clean, "..") {
		return "", fmt.Errorf("path traversal not allowed: %s", path)
	}

	// Convert to absolute path for consistency
	abs, err := filepath.Abs(clean)
	if err != nil {
		return "", fmt.Errorf("invalid file path: %s", path)
	}

	return abs, nil
}

// validatePathSafety ensures the file path is safe and doesn't contain traversal attempts
func (fv *FileValidator) validatePathSafety(filePath string) error {
	_, err := fv.sanitizeFilePath(filePath)
	return err
}

// validateDirectoryPermissions checks if the directory exists and is writable
func (fv *FileValidator) validateDirectoryPermissions(filePath string) error {
	dir := filepath.Dir(filePath)

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	// Test write permissions by creating a temporary file
	tempFile := filepath.Join(dir, ".strata_write_test")
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("no write permission for directory: %s", dir)
	}
	file.Close()
	os.Remove(tempFile)

	return nil
}

// validateFormatSupport checks if the specified output format is supported
func (fv *FileValidator) validateFormatSupport(format string) error {
	supportedFormats := []string{
		"table",
		"json",
		"csv",
		"markdown",
		"html",
		"dot",
	}

	formatLower := strings.ToLower(format)
	for _, supported := range supportedFormats {
		if formatLower == supported {
			return nil
		}
	}

	return fmt.Errorf("unsupported output format: %s, supported formats: %v", format, supportedFormats)
}

// FileOutputError represents errors that occur during file output operations
type FileOutputError struct {
	Type    string // "validation", "permission", "format", "write"
	Path    string
	Format  string
	Message string
	Cause   error
}

func (e *FileOutputError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s error for %s: %s (caused by: %v)",
			e.Type, e.Path, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s error for %s: %s", e.Type, e.Path, e.Message)
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

	return result
}

// checkFileOverwrite checks if a file will be overwritten and adds appropriate warnings
func (fv *FileValidator) checkFileOverwrite(filePath string, result *ValidationResult) {
	if _, err := os.Stat(filePath); err == nil {
		result.AddWarning(fmt.Sprintf("File %s will be overwritten", filePath))
	}
}
