package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Security Tests for Path Traversal Prevention

func TestSecurity_PathTraversal(t *testing.T) {
	config := &Config{}
	validator := NewFileValidator(config)

	// Comprehensive list of malicious path inputs
	maliciousPaths := []struct {
		name string
		path string
		desc string
	}{
		{
			name: "unix_path_traversal_basic",
			path: "../../../etc/passwd",
			desc: "Basic Unix path traversal to /etc/passwd",
		},
		{
			name: "unix_path_traversal_shadow",
			path: "../../../etc/shadow",
			desc: "Unix path traversal to /etc/shadow",
		},
		{
			name: "windows_path_traversal_basic",
			path: "..\\..\\..\\windows\\system32\\config\\sam",
			desc: "Windows path traversal to SAM file",
		},
		{
			name: "windows_path_traversal_hosts",
			path: "..\\..\\..\\windows\\system32\\drivers\\etc\\hosts",
			desc: "Windows path traversal to hosts file",
		},
		{
			name: "mixed_separators",
			path: "../..\\../etc/passwd",
			desc: "Mixed path separators for traversal",
		},
		{
			name: "home_directory_traversal",
			path: "~/../../etc/passwd",
			desc: "Home directory based traversal",
		},
		{
			name: "url_encoded_traversal",
			path: "..%2F..%2F..%2Fetc%2Fpasswd",
			desc: "URL encoded path traversal",
		},
		{
			name: "double_encoded_traversal",
			path: "..%252F..%252F..%252Fetc%252Fpasswd",
			desc: "Double URL encoded path traversal",
		},
		{
			name: "unicode_traversal",
			path: "..\u002F..\u002F..\u002Fetc\u002Fpasswd",
			desc: "Unicode encoded path traversal",
		},
		{
			name: "null_byte_injection",
			path: "../../../etc/passwd\x00.txt",
			desc: "Null byte injection attempt",
		},
		{
			name: "deep_traversal",
			path: "../../../../../../../../../../../../../etc/passwd",
			desc: "Deep path traversal with many levels",
		},
		{
			name: "traversal_with_valid_suffix",
			path: "../../../etc/passwd.json",
			desc: "Path traversal disguised with valid file extension",
		},
		{
			name: "traversal_in_middle",
			path: "reports/../../../etc/passwd",
			desc: "Path traversal in middle of seemingly valid path",
		},
		{
			name: "dot_dot_slash_variations",
			path: "....//....//....//etc/passwd",
			desc: "Variations of dot-dot-slash pattern",
		},
		{
			name: "backslash_traversal",
			path: "..\\..\\..\\etc\\passwd",
			desc: "Backslash-based path traversal",
		},
		{
			name: "current_dir_traversal",
			path: "./../../etc/passwd",
			desc: "Current directory based traversal",
		},
		{
			name: "symlink_traversal",
			path: "/tmp/../etc/passwd",
			desc: "Symlink-style traversal",
		},
	}

	for _, tt := range maliciousPaths {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validatePathSafety(tt.path)
			if err == nil {
				t.Errorf("Expected path traversal to be blocked for %s: %s", tt.desc, tt.path)
				return
			}

			// Verify the error message indicates path traversal prevention
			errorMsg := err.Error()
			if !strings.Contains(strings.ToLower(errorMsg), "path traversal") &&
				!strings.Contains(strings.ToLower(errorMsg), "not allowed") {
				t.Errorf("Error message should indicate path traversal prevention. Got: %s", errorMsg)
			}
		})
	}
}

func TestSecurity_FileOverwriteScenarios(t *testing.T) {
	config := &Config{}
	validator := NewFileValidator(config)

	// Create temporary directory and files for testing
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "existing.json")

	// Create an existing file
	file, err := os.Create(existingFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if _, err := file.WriteString(`{"test": "data"}`); err != nil {
		t.Fatalf("Failed to write to test file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Failed to close test file: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		expectWarn  bool
		description string
	}{
		{
			name:        "overwrite_existing_file",
			filePath:    existingFile,
			expectWarn:  true,
			description: "Should warn when overwriting existing file",
		},
		{
			name:        "new_file_no_warning",
			filePath:    filepath.Join(tempDir, "new_file.json"),
			expectWarn:  false,
			description: "Should not warn for new file",
		},
		{
			name:        "overwrite_system_file_simulation",
			filePath:    "/etc/hosts", // This will fail validation before overwrite check
			expectWarn:  true,         // File exists so warning is expected
			description: "System files generate overwrite warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{Valid: true}
			validator.checkFileOverwrite(tt.filePath, result)

			hasWarning := len(result.Warnings) > 0
			if hasWarning != tt.expectWarn {
				t.Errorf("%s: expected warning=%v, got warning=%v", tt.description, tt.expectWarn, hasWarning)
			}

			if tt.expectWarn && hasWarning {
				// Verify warning message mentions overwrite
				warningMsg := result.Warnings[0]
				if !strings.Contains(strings.ToLower(warningMsg), "overwritten") {
					t.Errorf("Warning message should mention overwrite. Got: %s", warningMsg)
				}
			}
		})
	}
}

func TestSecurity_SensitivePathBlocking(t *testing.T) {
	config := &Config{}
	validator := NewFileValidator(config)

	// List of sensitive system paths that should be blocked
	sensitivePaths := []struct {
		name string
		path string
		desc string
	}{
		{
			name: "etc_passwd",
			path: "/etc/passwd",
			desc: "Unix password file",
		},
		{
			name: "etc_shadow",
			path: "/etc/shadow",
			desc: "Unix shadow password file",
		},
		{
			name: "etc_hosts",
			path: "/etc/hosts",
			desc: "System hosts file",
		},
		{
			name: "proc_version",
			path: "/proc/version",
			desc: "Kernel version information",
		},
		{
			name: "proc_meminfo",
			path: "/proc/meminfo",
			desc: "Memory information",
		},
		{
			name: "windows_sam",
			path: "/nonexistent/Windows/System32/config/SAM",
			desc: "Windows Security Account Manager (Unix-style path)",
		},
		{
			name: "windows_system",
			path: "/nonexistent/Windows/System32/config/SYSTEM",
			desc: "Windows system registry (Unix-style path)",
		},
		{
			name: "boot_ini",
			path: "/nonexistent/boot.ini",
			desc: "Windows boot configuration (Unix-style path)",
		},
		{
			name: "ssh_private_key",
			path: "/home/user/.ssh/id_rsa",
			desc: "SSH private key",
		},
		{
			name: "aws_credentials",
			path: "/home/user/.aws/credentials",
			desc: "AWS credentials file",
		},
	}

	for _, tt := range sensitivePaths {
		t.Run(tt.name, func(t *testing.T) {
			// Test direct path access
			settings := &OutputConfiguration{
				OutputFile:       tt.path,
				OutputFileFormat: "json",
			}

			err := validator.ValidateFileOutput(settings)
			if err == nil {
				t.Errorf("Expected validation to fail for sensitive path %s (%s)", tt.path, tt.desc)
				return
			}

			// The error should be related to directory permissions or path validation
			errorMsg := strings.ToLower(err.Error())
			if !strings.Contains(errorMsg, "permission") &&
				!strings.Contains(errorMsg, "not exist") &&
				!strings.Contains(errorMsg, "validation") {
				t.Errorf("Error should indicate permission or validation issue. Got: %s", err.Error())
			}
		})
	}
}

func TestSecurity_PathNormalization(t *testing.T) {
	config := &Config{}
	validator := NewFileValidator(config)

	tests := []struct {
		name        string
		input       string
		expectError bool
		description string
	}{
		{
			name:        "double_slashes",
			input:       "reports//output.json",
			expectError: false,
			description: "Double slashes should be normalized",
		},
		{
			name:        "trailing_slash",
			input:       "reports/",
			expectError: false,
			description: "Trailing slash should be handled",
		},
		{
			name:        "current_directory_reference",
			input:       "./reports/output.json",
			expectError: false,
			description: "Current directory reference should be normalized",
		},
		{
			name:        "parent_directory_in_safe_context",
			input:       "reports/../reports/output.json",
			expectError: true, // This contains .. so should be blocked
			description: "Parent directory reference should be blocked even in safe context",
		},
		{
			name:        "complex_path_with_dots",
			input:       "reports/./subdir/../output.json",
			expectError: true, // Contains .. so should be blocked
			description: "Complex path with parent reference should be blocked",
		},
		{
			name:        "hidden_file",
			input:       ".hidden_output.json",
			expectError: false,
			description: "Hidden files should be allowed",
		},
		{
			name:        "hidden_directory",
			input:       ".config/output.json",
			expectError: false,
			description: "Hidden directories should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.sanitizeFilePath(tt.input)
			hasError := err != nil

			if hasError != tt.expectError {
				t.Errorf("%s: expected error=%v, got error=%v (error: %v)", tt.description, tt.expectError, hasError, err)
			}
		})
	}
}

func TestSecurity_ComprehensiveValidation(t *testing.T) {
	config := &Config{}
	validator := NewFileValidator(config)

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		settings    *OutputConfiguration
		expectError bool
		errorType   string
		description string
	}{
		{
			name: "valid_configuration",
			settings: &OutputConfiguration{
				OutputFile:       filepath.Join(tempDir, "valid_output.json"),
				OutputFileFormat: "json",
			},
			expectError: false,
			description: "Valid configuration should pass all validations",
		},
		{
			name: "path_traversal_attack",
			settings: &OutputConfiguration{
				OutputFile:       "../../../etc/passwd",
				OutputFileFormat: "json",
			},
			expectError: true,
			errorType:   "path",
			description: "Path traversal should be blocked",
		},
		{
			name: "unsupported_format",
			settings: &OutputConfiguration{
				OutputFile:       filepath.Join(tempDir, "output.xml"),
				OutputFileFormat: "xml",
			},
			expectError: true,
			errorType:   "format",
			description: "Unsupported format should be rejected",
		},
		{
			name: "nonexistent_directory",
			settings: &OutputConfiguration{
				OutputFile:       "/nonexistent/directory/output.json",
				OutputFileFormat: "json",
			},
			expectError: true,
			errorType:   "permission",
			description: "Nonexistent directory should be rejected",
		},
		{
			name: "empty_file_path",
			settings: &OutputConfiguration{
				OutputFile:       "",
				OutputFileFormat: "json",
			},
			expectError: false,
			description: "Empty file path should be allowed (no file output)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateFileOutput(tt.settings)
			hasError := err != nil

			if hasError != tt.expectError {
				t.Errorf("%s: expected error=%v, got error=%v (error: %v)", tt.description, tt.expectError, hasError, err)
				return
			}

			if tt.expectError && tt.errorType != "" {
				errorMsg := strings.ToLower(err.Error())
				if !strings.Contains(errorMsg, tt.errorType) {
					t.Errorf("%s: expected error type '%s' in error message. Got: %s", tt.description, tt.errorType, err.Error())
				}
			}
		})
	}
}

func TestStructuredErrorCodes(t *testing.T) {
	config := &Config{}
	validator := NewFileValidator(config)

	tests := []struct {
		name         string
		operation    func() error
		expectedCode string
		expectedType string
	}{
		{
			name: "path traversal error",
			operation: func() error {
				return validator.validatePathSafety("../../../etc/passwd")
			},
			expectedCode: "PATH_TRAVERSAL",
			expectedType: "validation",
		},
		{
			name: "unsupported format error",
			operation: func() error {
				return validator.validateFormatSupport("xml")
			},
			expectedCode: "UNSUPPORTED_FORMAT",
			expectedType: "format",
		},
		{
			name: "directory not found error",
			operation: func() error {
				return validator.validateDirectoryPermissions("/nonexistent/directory/file.json")
			},
			expectedCode: "DIRECTORY_NOT_FOUND",
			expectedType: "permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operation()
			if err == nil {
				t.Errorf("Expected error but got nil")
				return
			}

			fileErr, ok := err.(*FileOutputError)
			if !ok {
				t.Errorf("Expected FileOutputError but got %T", err)
				return
			}

			if fileErr.Code != tt.expectedCode {
				t.Errorf("Expected error code %s but got %s", tt.expectedCode, fileErr.Code)
			}

			if fileErr.Type != tt.expectedType {
				t.Errorf("Expected error type %s but got %s", tt.expectedType, fileErr.Type)
			}
		})
	}
}
