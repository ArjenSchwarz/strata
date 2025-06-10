# go-output Library Documentation

A comprehensive Go library for transforming input data into various output formats. This library provides a unified interface for generating output in multiple formats including JSON, YAML, CSV, HTML, Markdown, tables, DOT graphs, Mermaid diagrams, and Draw.io/Diagrams.net CSV imports.

## Table of Contents

1. [Overview](overview.md) - High-level architecture and concepts
2. [Quick Start](quick-start.md) - Get started with basic usage
3. [Core Concepts](core-concepts.md) - Understanding the fundamental data structures
4. [Output Formats](output-formats.md) - Complete guide to all supported output formats
5. [Settings and Configuration](settings.md) - Comprehensive configuration options
6. [Color and Formatting](colors.md) - Text styling and color output
7. [Mermaid Diagrams](mermaid.md) - Creating flowcharts, pie charts, and Gantt charts
8. [Draw.io Integration](drawio.md) - Generating CSV files for Draw.io import
9. [Advanced Usage](advanced-usage.md) - Complex scenarios and best practices
10. [API Reference](api-reference.md) - Complete API documentation
11. [Examples](examples.md) - Practical examples and use cases

## Key Features

- **Multiple Output Formats**: Support for 10+ output formats including structured data (JSON, YAML, CSV), visual formats (HTML, Markdown, tables), and diagram formats (Mermaid, DOT, Draw.io)
- **Unified Interface**: Single API for all output formats with consistent configuration
- **Extensible Architecture**: Modular design allowing easy addition of new output formats
- **Rich Styling**: Color support, emoji support, and customizable table styles
- **Diagram Generation**: Native support for Mermaid flowcharts, pie charts, and Gantt charts
- **Draw.io Integration**: Generate CSV files that can be imported directly into Draw.io/Diagrams.net
- **S3 Support**: Direct output to Amazon S3 buckets
- **Flexible Data Handling**: Support for complex data structures and relationships

## Package Structure

```
github.com/ArjenSchwarz/go-output/
├── format/               # Main package
├── mermaid/             # Mermaid diagram generation
├── drawio/              # Draw.io CSV generation
├── templates/           # HTML templates
└── docs/                # Documentation (this directory)
```

## Dependencies

- `github.com/jedib0t/go-pretty/v6/table` - Table rendering
- `github.com/emicklei/dot` - DOT graph generation
- `github.com/fatih/color` - Terminal color support
- `github.com/gosimple/slug` - URL slug generation
- `gopkg.in/yaml.v3` - YAML marshaling
- `github.com/aws/aws-sdk-go-v2/service/s3` - S3 integration

## License

This library is part of the go-output project by [ArjenSchwarz](https://github.com/ArjenSchwarz/). Check the main repository for license information.
