package terraform

import (
	"regexp"
	"strconv"
	"strings"
)

// DefaultOutputParser is the default implementation of TerraformOutputParser
type DefaultOutputParser struct{}

// NewOutputParser creates a new Terraform output parser
func NewOutputParser() TerraformOutputParser {
	return &DefaultOutputParser{}
}

// ParsePlanOutput parses the output of terraform plan
func (p *DefaultOutputParser) ParsePlanOutput(output string) (*PlanOutput, error) {
	result := &PlanOutput{
		RawOutput: output,
	}

	// Check if there are any changes
	if strings.Contains(output, "No changes. Your infrastructure matches the configuration.") ||
		strings.Contains(output, "No changes. Infrastructure is up-to-date.") {
		result.HasChanges = false
		return result, nil
	}

	// Look for the plan summary line (e.g., "Plan: 2 to add, 1 to change, 0 to destroy.")
	planRegex := regexp.MustCompile(`Plan: (\d+) to add, (\d+) to change, (\d+) to destroy\.`)
	matches := planRegex.FindStringSubmatch(output)

	if len(matches) == 4 {
		result.HasChanges = true

		// Parse the numbers
		if add, err := strconv.Atoi(matches[1]); err == nil {
			result.ResourceChanges.Add = add
		}
		if change, err := strconv.Atoi(matches[2]); err == nil {
			result.ResourceChanges.Change = change
		}
		if destroy, err := strconv.Atoi(matches[3]); err == nil {
			result.ResourceChanges.Destroy = destroy
		}
	} else {
		// Try alternative formats for plan summary
		// Look for individual resource change indicators
		result.HasChanges = p.detectChangesFromResourceLines(output, result)
	}

	// Extract additional information
	result.ExitCode = p.extractExitCode(output)

	return result, nil
}

// detectChangesFromResourceLines detects changes by analyzing individual resource lines
func (p *DefaultOutputParser) detectChangesFromResourceLines(output string, result *PlanOutput) bool {
	lines := strings.Split(output, "\n")
	hasChanges := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for resource change indicators
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "++") {
			result.ResourceChanges.Add++
			hasChanges = true
		} else if strings.HasPrefix(line, "~") {
			result.ResourceChanges.Change++
			hasChanges = true
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "--") {
			result.ResourceChanges.Destroy++
			hasChanges = true
		} else if strings.Contains(line, "will be created") {
			result.ResourceChanges.Add++
			hasChanges = true
		} else if strings.Contains(line, "will be updated") || strings.Contains(line, "will be modified") {
			result.ResourceChanges.Change++
			hasChanges = true
		} else if strings.Contains(line, "will be destroyed") {
			result.ResourceChanges.Destroy++
			hasChanges = true
		}
	}

	return hasChanges
}

// extractExitCode attempts to extract exit code information from output
func (p *DefaultOutputParser) extractExitCode(output string) int {
	// This is a placeholder - in practice, exit codes come from the command execution
	// We can enhance this to detect error patterns that indicate specific exit codes
	if strings.Contains(output, "Error:") {
		return 1
	}
	return 0
}

// ParseApplyOutput parses the output of terraform apply
func (p *DefaultOutputParser) ParseApplyOutput(output string) (*ApplyOutput, error) {
	result := &ApplyOutput{
		RawOutput: output,
	}

	// Check for success indicators
	if strings.Contains(output, "Apply complete!") {
		result.Success = true

		// Look for the apply summary line (e.g., "Apply complete! Resources: 2 added, 1 changed, 0 destroyed.")
		applyRegex := regexp.MustCompile(`Apply complete! Resources: (\d+) added, (\d+) changed, (\d+) destroyed\.`)
		matches := applyRegex.FindStringSubmatch(output)

		if len(matches) == 4 {
			if added, err := strconv.Atoi(matches[1]); err == nil {
				result.ResourceChanges.Added = added
			}
			if changed, err := strconv.Atoi(matches[2]); err == nil {
				result.ResourceChanges.Changed = changed
			}
			if destroyed, err := strconv.Atoi(matches[3]); err == nil {
				result.ResourceChanges.Destroyed = destroyed
			}
		}
	} else {
		result.Success = false

		// Extract detailed error information
		result.Error = p.extractApplyErrors(output)
	}

	// Extract exit code information
	result.ExitCode = p.extractApplyExitCode(output)

	return result, nil
}

// extractApplyErrors extracts detailed error information from apply output
func (p *DefaultOutputParser) extractApplyErrors(output string) string {
	lines := strings.Split(output, "\n")
	var errorLines []string
	var errorSections []string
	inError := false
	currentSection := []string{}

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Start of an error section
		if strings.Contains(trimmedLine, "Error:") ||
			strings.Contains(trimmedLine, "│ Error:") ||
			strings.Contains(trimmedLine, "╷") {
			if len(currentSection) > 0 {
				errorSections = append(errorSections, strings.Join(currentSection, "\n"))
				currentSection = []string{}
			}
			inError = true
		}

		// End of an error section
		if inError && (strings.Contains(trimmedLine, "╵") ||
			(trimmedLine == "" && len(currentSection) > 0)) {
			if len(currentSection) > 0 {
				errorSections = append(errorSections, strings.Join(currentSection, "\n"))
				currentSection = []string{}
			}
			inError = false
		}

		// Collect error lines
		if inError {
			currentSection = append(currentSection, line)
		}

		// Also collect lines that contain common error patterns
		if strings.Contains(trimmedLine, "Error:") ||
			strings.Contains(trimmedLine, "Failed to") ||
			strings.Contains(trimmedLine, "could not") ||
			strings.Contains(trimmedLine, "unable to") {
			errorLines = append(errorLines, line)
		}
	}

	// Add any remaining section
	if len(currentSection) > 0 {
		errorSections = append(errorSections, strings.Join(currentSection, "\n"))
	}

	// Prefer structured error sections, fall back to individual error lines
	if len(errorSections) > 0 {
		return strings.Join(errorSections, "\n\n")
	}

	if len(errorLines) > 0 {
		return strings.Join(errorLines, "\n")
	}

	// Check for specific error patterns
	if strings.Contains(output, "timeout") {
		return "Apply operation timed out"
	}

	if strings.Contains(output, "interrupted") {
		return "Apply operation was interrupted"
	}

	if strings.Contains(output, "cancelled") {
		return "Apply operation was cancelled"
	}

	return "Apply failed with unknown error"
}

// extractApplyExitCode determines the exit code based on apply output patterns
func (p *DefaultOutputParser) extractApplyExitCode(output string) int {
	if strings.Contains(output, "Apply complete!") {
		return 0
	}

	// Check for specific error patterns that indicate different exit codes
	if strings.Contains(output, "timeout") {
		return 124 // Timeout exit code
	}

	if strings.Contains(output, "interrupted") || strings.Contains(output, "cancelled") {
		return 130 // Interrupted exit code
	}

	if strings.Contains(output, "Error:") {
		return 1 // General error exit code
	}

	return 1 // Default error exit code
}
