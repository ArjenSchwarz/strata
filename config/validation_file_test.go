package config

import (
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
