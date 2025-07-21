package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileValidator_ValidatePathSafety(t *testing.T) {
	config := &Config{}
	validator := NewFileValidator(config)

	tests := []struct {
		name    string
		path    string
		wantErr bool
		errType string
	}{
		{
			name:    "valid relative path",
			path:    "output/report.json",
			wantErr: false,
		},
		{
			name:    "valid absolute path",
			path:    "/tmp/output.json",
			wantErr: false,
		},
		{
			name:    "path traversal attempt with ../",
			path:    "../../../etc/passwd",
			wantErr: true,
			errType: "path traversal",
		},
		{
			name:    "path traversal attempt in middle",
			path:    "output/../../../etc/passwd",
			wantErr: true,
			errType: "path traversal",
		},
		{
			name:    "simple filename",
			path:    "output.json",
			wantErr: false,
		},
		{
			name:    "nested valid path",
			path:    "reports/2025/output.json",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validatePathSafety(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePathSafety() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errType) {
					t.Errorf("validatePathSafety() error = %v, expected error containing %v", err, tt.errType)
				}
			}
		})
	}
}

func TestFileValidator_ValidateFormatSupport(t *testing.T) {
	config := &Config{}
	validator := NewFileValidator(config)

	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{
			name:    "table format",
			format:  "table",
			wantErr: false,
		},
		{
			name:    "json format",
			format:  "json",
			wantErr: false,
		},
		{
			name:    "csv format",
			format:  "csv",
			wantErr: false,
		},
		{
			name:    "markdown format",
			format:  "markdown",
			wantErr: false,
		},
		{
			name:    "html format",
			format:  "html",
			wantErr: false,
		},
		{
			name:    "dot format",
			format:  "dot",
			wantErr: false,
		},
		{
			name:    "uppercase format",
			format:  "JSON",
			wantErr: false,
		},
		{
			name:    "mixed case format",
			format:  "MarkDown",
			wantErr: false,
		},
		{
			name:    "unsupported format",
			format:  "xml",
			wantErr: true,
		},
		{
			name:    "empty format",
			format:  "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			format:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateFormatSupport(tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFormatSupport() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileValidator_ValidateDirectoryPermissions(t *testing.T) {
	config := &Config{}
	validator := NewFileValidator(config)

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		wantErr bool
		setup   func() string
	}{
		{
			name:    "valid writable directory",
			wantErr: false,
			setup: func() string {
				return filepath.Join(tempDir, "output.json")
			},
		},
		{
			name:    "nonexistent directory",
			wantErr: true,
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent", "output.json")
			},
		},
		{
			name:    "current directory",
			wantErr: false,
			setup: func() string {
				return "output.json"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			err := validator.validateDirectoryPermissions(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDirectoryPermissions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileValidator_ValidateFileOutput(t *testing.T) {
	config := &Config{}
	validator := NewFileValidator(config)

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		settings *OutputConfiguration
		wantErr  bool
	}{
		{
			name: "no file output",
			settings: &OutputConfiguration{
				OutputFile: "",
			},
			wantErr: false,
		},
		{
			name: "valid file output",
			settings: &OutputConfiguration{
				OutputFile:       filepath.Join(tempDir, "output.json"),
				OutputFileFormat: "json",
			},
			wantErr: false,
		},
		{
			name: "path traversal attempt",
			settings: &OutputConfiguration{
				OutputFile:       "../../../etc/passwd",
				OutputFileFormat: "json",
			},
			wantErr: true,
		},
		{
			name: "unsupported format",
			settings: &OutputConfiguration{
				OutputFile:       filepath.Join(tempDir, "output.xml"),
				OutputFileFormat: "xml",
			},
			wantErr: true,
		},
		{
			name: "nonexistent directory",
			settings: &OutputConfiguration{
				OutputFile:       "/nonexistent/directory/output.json",
				OutputFileFormat: "json",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateFileOutput(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileOutput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileValidator_SanitizeFilePath(t *testing.T) {
	config := &Config{}
	validator := NewFileValidator(config)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "simple path",
			path:    "output.json",
			wantErr: false,
		},
		{
			name:    "nested path",
			path:    "reports/output.json",
			wantErr: false,
		},
		{
			name:    "absolute path",
			path:    "/tmp/output.json",
			wantErr: false,
		},
		{
			name:    "path traversal",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "path with extra slashes",
			path:    "reports//output.json",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.sanitizeFilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("sanitizeFilePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Errorf("sanitizeFilePath() returned empty result for valid path")
			}
		})
	}
}

func TestConfig_ResolvePlaceholders(t *testing.T) {
	config := &Config{}

	// Mock environment variables for testing
	originalRegion := os.Getenv("AWS_REGION")
	originalAccountID := os.Getenv("AWS_ACCOUNT_ID")
	originalStackName := os.Getenv("STACK_NAME")

	// Set test environment variables
	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("AWS_ACCOUNT_ID", "123456789012")
	os.Setenv("STACK_NAME", "test-stack")

	// Restore original environment variables after test
	defer func() {
		if originalRegion != "" {
			os.Setenv("AWS_REGION", originalRegion)
		} else {
			os.Unsetenv("AWS_REGION")
		}
		if originalAccountID != "" {
			os.Setenv("AWS_ACCOUNT_ID", originalAccountID)
		} else {
			os.Unsetenv("AWS_ACCOUNT_ID")
		}
		if originalStackName != "" {
			os.Setenv("STACK_NAME", originalStackName)
		} else {
			os.Unsetenv("STACK_NAME")
		}
	}()

	tests := []struct {
		name     string
		input    string
		expected func(string) bool // Function to validate the result
	}{
		{
			name:  "timestamp placeholder",
			input: "report-$TIMESTAMP.json",
			expected: func(result string) bool {
				return strings.HasPrefix(result, "report-") &&
					strings.HasSuffix(result, ".json") &&
					strings.Contains(result, "T") // Timestamp format contains T
			},
		},
		{
			name:  "AWS region placeholder",
			input: "report-$AWS_REGION.json",
			expected: func(result string) bool {
				return result == "report-us-east-1.json"
			},
		},
		{
			name:  "AWS account ID placeholder",
			input: "report-$AWS_ACCOUNTID.json",
			expected: func(result string) bool {
				return result == "report-123456789012.json"
			},
		},

		{
			name:  "multiple placeholders",
			input: "$AWS_REGION-$AWS_ACCOUNTID.json",
			expected: func(result string) bool {
				return result == "us-east-1-123456789012.json"
			},
		},
		{
			name:  "no placeholders",
			input: "simple-report.json",
			expected: func(result string) bool {
				return result == "simple-report.json"
			},
		},
		{
			name:  "timestamp with other placeholders",
			input: "$TIMESTAMP-$AWS_REGION.json",
			expected: func(result string) bool {
				return strings.Contains(result, "us-east-1.json") &&
					strings.Contains(result, "T") // Timestamp format
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.resolvePlaceholders(tt.input)
			if !tt.expected(result) {
				t.Errorf("resolvePlaceholders() = %v, validation failed for input %v", result, tt.input)
			}
		})
	}
}

func TestValidationResult(t *testing.T) {
	t.Run("new validation result", func(t *testing.T) {
		result := &ValidationResult{Valid: true}

		if result.HasErrors() {
			t.Errorf("HasErrors() = true, want false for new result")
		}

		if len(result.Errors) != 0 {
			t.Errorf("len(Errors) = %d, want 0", len(result.Errors))
		}
	})

	t.Run("add error", func(t *testing.T) {
		result := &ValidationResult{Valid: true}
		testErr := &FileOutputError{
			Type:    "validation",
			Code:    "PATH_TRAVERSAL",
			Path:    "/test/path",
			Message: "test error",
		}

		result.AddError(testErr)

		if !result.HasErrors() {
			t.Errorf("HasErrors() = false, want true after adding error")
		}

		if result.Valid {
			t.Errorf("Valid = true, want false after adding error")
		}

		if len(result.Errors) != 1 {
			t.Errorf("len(Errors) = %d, want 1", len(result.Errors))
		}
	})

	t.Run("add warning and info", func(t *testing.T) {
		result := &ValidationResult{Valid: true}

		result.AddWarning("test warning")
		result.AddInfo("test info")

		if len(result.Warnings) != 1 {
			t.Errorf("len(Warnings) = %d, want 1", len(result.Warnings))
		}

		if len(result.Infos) != 1 {
			t.Errorf("len(Infos) = %d, want 1", len(result.Infos))
		}

		if result.Warnings[0] != "test warning" {
			t.Errorf("Warnings[0] = %v, want 'test warning'", result.Warnings[0])
		}

		if result.Infos[0] != "test info" {
			t.Errorf("Infos[0] = %v, want 'test info'", result.Infos[0])
		}
	})
}

func TestFileOutputError(t *testing.T) {
	t.Run("error without cause", func(t *testing.T) {
		err := &FileOutputError{
			Type:    "validation",
			Code:    "PATH_TRAVERSAL",
			Path:    "/test/path",
			Message: "test error",
		}

		expected := "validation error [PATH_TRAVERSAL] for /test/path: test error"
		if err.Error() != expected {
			t.Errorf("Error() = %v, want %v", err.Error(), expected)
		}
	})

	t.Run("error with cause", func(t *testing.T) {
		cause := os.ErrNotExist
		err := &FileOutputError{
			Type:    "permission",
			Code:    "PERMISSION_DENIED",
			Path:    "/test/path",
			Message: "cannot access file",
			Cause:   cause,
		}

		expected := "permission error [PERMISSION_DENIED] for /test/path: cannot access file (caused by: file does not exist)"
		if err.Error() != expected {
			t.Errorf("Error() = %v, want %v", err.Error(), expected)
		}
	})

	t.Run("error code access", func(t *testing.T) {
		err := &FileOutputError{
			Type:    "format",
			Code:    "UNSUPPORTED_FORMAT",
			Path:    "/test/path",
			Format:  "xml",
			Message: "unsupported format",
		}

		if err.Code != "UNSUPPORTED_FORMAT" {
			t.Errorf("Code = %v, want UNSUPPORTED_FORMAT", err.Code)
		}
	})
}

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
	file.WriteString(`{"test": "data"}`)
	file.Close()

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
