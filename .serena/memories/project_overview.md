# Strata Project Overview

## Purpose
Strata is a Go CLI tool that enhances Terraform workflows by providing clear, concise summaries of Terraform plan changes. It acts as a read-only analysis tool that doesn't modify Terraform workflows but provides better visibility into what Terraform plans will do, helping users make more informed decisions before applying changes.

## Key Features
- Parse and summarize Terraform plan files (both binary and JSON formats)
- Highlight potentially destructive changes
- Generate statistical summaries of resource modifications
- Support multiple output formats (table, JSON, HTML, Markdown)
- CI/CD pipeline integration (GitHub Actions support)
- Danger highlights for sensitive resources and properties
- Progressive disclosure with collapsible sections
- Provider grouping for large plans
- Global expand-all control for comprehensive details
- Risk assessment with automatic prioritization

## Tech Stack
- **Language**: Go 1.24.1 (minimum)
- **CLI Framework**: Cobra v1.9.1 (command-line interface)
- **Configuration**: Viper v1.20.1 (configuration management)
- **Output Formatting**: go-output v2.1.0 (multi-format output generation)
- **Terraform Integration**: terraform-json v0.25.0 (plan file parsing)
- **Testing**: testify v1.10.0 (testing assertions)

## Architecture
- **Command Pattern**: Using Cobra for CLI commands
- **Layered Architecture**: 
  - Command Layer (cmd/) - CLI interaction
  - Library Layer (lib/) - Core business logic
  - Configuration Layer (config/) - Application settings
- **Data Flow**: User Input → Command Layer → Library Layer → Output Formatting → User Display

## Current Status
Active development with the following completed phases:
- ✅ CLI Foundation & Configuration
- ✅ Terraform Plan Parsing
- ✅ Summary Analysis Engine
- ✅ Output Integration
- ✅ Enhanced Summary Visualization (Progressive Disclosure)
- 🔄 Future Integration Preparation

## Platform
Running on Darwin (macOS) platform with standard Unix toolchain available.