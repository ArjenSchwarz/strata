# Strata Project Guidelines and Design Patterns

## Core Design Principles

### 1. Read-Only Analysis Tool
- Strata NEVER modifies Terraform workflows
- Acts as a passive analyzer providing insights
- No side effects on infrastructure or state

### 2. Progressive Disclosure
- Show essential information by default
- Provide detailed information in collapsible sections
- Auto-expand high-risk changes for immediate visibility
- Support global expand control for power users

### 3. Risk-Based Prioritization
- Automatically detect and highlight dangerous changes
- Calculate risk levels (critical, high, medium, low)
- Auto-expand high-risk sections
- Support configurable sensitive resources/properties

## Architecture Patterns

### Command Pattern with Cobra
- Root command defines global configuration
- Subcommands implement specific functionality
- Commands delegate to library code (separation of concerns)

### Layered Architecture
1. **Command Layer** (cmd/): CLI interaction only
2. **Library Layer** (lib/): Core business logic
3. **Configuration Layer** (config/): Settings management

### Data Processing Pipeline
```
Plan File → Parser → Analyzer → Formatter → Output
```

## Configuration Philosophy
- YAML-based configuration for readability
- Support multiple config locations (./strata.yaml, ~/.strata.yaml)
- CLI flags override configuration file settings
- Sensible defaults that work out of the box

## Output Format Strategy
- Support multiple formats (table, JSON, HTML, Markdown)
- Format-specific adaptations (e.g., collapsible sections)
- Consistent data structure across formats
- Progressive disclosure adapts to each format

## Testing Strategy
- Unit tests alongside code (*_test.go)
- Table-driven tests for multiple scenarios
- Integration tests for end-to-end workflows
- Sample files for development and testing
- High coverage in critical paths (parser, analyzer)

## Error Handling Philosophy
- Return errors, don't panic
- Wrap errors with context
- User-friendly error messages
- Fail fast with clear explanations

## Performance Considerations
- Efficient plan file parsing
- Lazy loading where possible
- Binary caching for GitHub Action
- Minimal dependencies

## Security Considerations
- Input validation and sanitization
- No logging of sensitive data
- Secure handling of plan files
- GitHub Action security-first design

## Documentation Philosophy
- Documentation aids understanding (not overly concise)
- Update README for user-facing features
- Implementation notes in docs/implementation/
- Code comments only when explicitly needed

## Feature Development Process
1. Create feature specification in specs/ directory
2. Document requirements, design, and tasks
3. Implement with tests
4. Update documentation
5. Validate with make check

## GitHub Action Design
- Modular shell scripts for maintainability
- Security-first with input validation
- Comprehensive error handling
- GitHub-specific integrations (PR comments, summaries)
- Binary caching for performance

## Current Development Focus
- Code cleanup and modernization (specs/code-cleanup-and-modernization)
- Using modern Go patterns and idioms
- Improving test coverage and quality
- Enhancing output formatting capabilities

## Important Project-Specific Rules
- NO comments in code unless explicitly requested
- ALWAYS run validation before completing tasks
- Follow existing patterns in the codebase
- Prefer editing existing files over creating new ones
- Documentation only when explicitly requested