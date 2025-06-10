# Overview

The go-output library provides a unified interface for transforming structured data into various output formats. It's designed to be used in CLI tools and applications that need to present data in different formats based on user preferences or specific use cases.

## Architecture

The library is built around three main concepts:

### 1. Data Structures

- **OutputHolder**: Represents a single row or entity with key-value pairs
- **OutputArray**: Contains multiple OutputHolders and configuration settings
- **OutputSettings**: Configuration object that controls output format and behavior

### 2. Output Formats

The library supports multiple categories of output formats:

#### Structured Data Formats
- **JSON**: Standard JSON output with raw data types preserved
- **YAML**: YAML format output
- **CSV**: Comma-separated values for spreadsheet compatibility

#### Presentation Formats
- **Table**: Terminal-friendly formatted tables with styling options
- **HTML**: Rich HTML tables with responsive CSS styling
- **Markdown**: Markdown tables for documentation

#### Diagram Formats
- **Mermaid**: Flowcharts, pie charts, and Gantt charts
- **DOT**: GraphViz DOT notation for network diagrams
- **Draw.io**: CSV format for import into Draw.io/Diagrams.net

### 3. Configuration System

The OutputSettings object provides extensive configuration options:

- Output format selection
- File output and S3 integration
- Visual styling (colors, emoji, table styles)
- Diagram-specific settings
- Content organization (TOC, titles, sorting)

## Data Flow

1. **Data Input**: Create OutputHolders with your data
2. **Configuration**: Set up OutputSettings with desired format and options
3. **Processing**: Create OutputArray combining data and settings
4. **Output**: Call Write() to generate formatted output

```
Input Data → OutputHolder → OutputArray + OutputSettings → Write() → Formatted Output
```

## Key Benefits

- **Consistency**: Same data structure works with all output formats
- **Flexibility**: Easy to switch between formats without changing data preparation
- **Extensibility**: New output formats can be added without breaking existing code
- **Integration**: Built-in support for cloud storage and external tools
- **Styling**: Rich formatting options for professional presentation

## Use Cases

- **CLI Tools**: Provide users with choice of output formats
- **Data Export**: Export application data in various formats
- **Documentation**: Generate documentation with embedded diagrams
- **Reporting**: Create visual reports with charts and tables
- **Integration**: Export data for use in external tools like Draw.io
