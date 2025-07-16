# Strata: Terraform Plan Analysis Tool

Strata is a CLI tool that enhances Terraform workflows by providing clear, concise summaries of plan changes. It doesn't modify your Terraform workflow or require special configuration - it simply provides better visibility into what your Terraform plans will do, helping you make more informed decisions before applying changes.

## Core Features

### Plan Analysis
- Parse and summarize Terraform plan files (binary and JSON formats)
- Generate comprehensive statistical summaries of resource modifications
- Highlight potentially destructive changes with configurable danger thresholds
- Display detailed resource changes with IDs, modules, and replacement indicators

### Output Formats & File Management
- Support multiple output formats: table, JSON, HTML, markdown
- Dual output capability (stdout + file with different formats)
- Dynamic file naming with placeholders ($TIMESTAMP, $AWS_REGION, $AWS_ACCOUNTID)
- Flexible file output for CI/CD integration and documentation

### Danger Detection & Sensitivity
- Configurable sensitive resources and properties
- Automatic highlighting of high-risk changes (⚠️ indicators)
- Separate tracking of sensitive dangerous changes in statistics
- Always-show-sensitive option for critical resource visibility

### Integration & Automation
- GitHub Action for CI/CD workflows with PR comment integration
- Version information with build details and Git commit tracking
- Configuration file support with YAML format
- CLI help system with detailed usage examples

## Key Concepts

- **Plan Summary**: Comprehensive overview including plan info, statistics, and detailed changes
- **Destructive Changes**: Operations that remove or replace resources
- **Sensitive Resources**: Configurable list of critical infrastructure components
- **Danger Threshold**: Configurable number of destructive changes that trigger warnings
- **High-Risk Changes**: Sensitive resources with danger flags tracked separately
- **File Output**: Dual output system supporting different formats for stdout and files

## Architecture

### Command Structure
- Root command with global flags and configuration
- Plan command group with summary subcommand
- Version command with detailed build information
- Layered architecture: Command → Library → Output

### Core Components
- **Plan Processing**: Parser, analyzer, formatter, and models
- **Configuration Management**: YAML-based with validation
- **Output System**: Multi-format support with file management
- **Version System**: Build-time injection of version metadata

## Project Status

The project is feature-complete and actively maintained:
1. ✅ CLI Foundation & Configuration
2. ✅ Terraform Plan Parsing (Binary & JSON)
3. ✅ Summary Analysis Engine with Danger Detection
4. ✅ Multi-format Output Integration
5. ✅ GitHub Action Integration
6. ✅ File Output System with Placeholders
7. ✅ Version Management System
8. 🔄 Ongoing maintenance and feature enhancements

## License

Strata is licensed under the MIT License.