# Examples

Practical examples and use cases demonstrating the go-output library's capabilities in real-world scenarios.

## Basic Examples

### Simple Data Export

```go
package main

import (
    "github.com/ArjenSchwarz/go-output"
)

func exportUserData() {
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"ID", "Name", "Email", "Role", "Active"},
    }

    // Sample user data
    users := []map[string]interface{}{
        {"ID": 1, "Name": "Alice Johnson", "Email": "alice@example.com", "Role": "Admin", "Active": true},
        {"ID": 2, "Name": "Bob Smith", "Email": "bob@example.com", "Role": "User", "Active": true},
        {"ID": 3, "Name": "Carol Davis", "Email": "carol@example.com", "Role": "User", "Active": false},
    }

    for _, user := range users {
        output.AddContents(user)
    }

    // Export as different formats
    output.Settings.SetOutputFormat("table")
    output.Settings.Title = "User Directory"
    output.Write()

    // Save as CSV
    output.Settings.SetOutputFormat("csv")
    output.Settings.OutputFile = "users.csv"
    output.Write()
}
```

### Configuration Report

```go
func generateConfigReport() {
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"Service", "Environment", "Config Key", "Value", "Sensitive"},
    }

    configs := []map[string]interface{}{
        {"Service": "api", "Environment": "prod", "Config Key": "PORT", "Value": "8080", "Sensitive": false},
        {"Service": "api", "Environment": "prod", "Config Key": "DB_PASSWORD", "Value": "***", "Sensitive": true},
        {"Service": "web", "Environment": "staging", "Config Key": "DEBUG", "Value": "true", "Sensitive": false},
    }

    for _, config := range configs {
        output.AddContents(config)
    }

    output.Settings.SetOutputFormat("markdown")
    output.Settings.Title = "Service Configuration"
    output.Settings.SortKey = "Service"
    output.Settings.OutputFile = "config-report.md"
    output.Write()
}
```

## Cloud Infrastructure Examples

### AWS Resource Inventory

```go
package main

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    "github.com/ArjenSchwarz/go-output"
)

func awsInventory() {
    // Initialize AWS client
    cfg, _ := config.LoadDefaultConfig(context.TODO())
    ec2Client := ec2.NewFromConfig(cfg)

    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"Instance ID", "Type", "State", "Name", "Launch Time", "Cost/Hour"},
    }

    // Fetch EC2 instances
    result, _ := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})

    for _, reservation := range result.Reservations {
        for _, instance := range reservation.Instances {
            name := "Unknown"
            for _, tag := range instance.Tags {
                if *tag.Key == "Name" {
                    name = *tag.Value
                    break
                }
            }

            output.AddContents(map[string]interface{}{
                "Instance ID": *instance.InstanceId,
                "Type":        string(instance.InstanceType),
                "State":       string(instance.State.Name),
                "Name":        name,
                "Launch Time": instance.LaunchTime.Format("2006-01-02 15:04"),
                "Cost/Hour":   getInstanceCost(string(instance.InstanceType)),
            })
        }
    }

    // Generate multiple output formats
    output.Settings.Title = "AWS EC2 Inventory"
    output.Settings.SortKey = "Launch Time"

    // HTML report
    output.Settings.SetOutputFormat("html")
    output.Settings.OutputFile = "aws-inventory.html"
    output.Write()

    // CSV for spreadsheet analysis
    output.Settings.SetOutputFormat("csv")
    output.Settings.OutputFile = "aws-inventory.csv"
    output.Write()
}

func getInstanceCost(instanceType string) float64 {
    costs := map[string]float64{
        "t3.micro":  0.0104,
        "t3.small":  0.0208,
        "t3.medium": 0.0416,
        "m5.large":  0.096,
        "m5.xlarge": 0.192,
    }
    if cost, exists := costs[instanceType]; exists {
        return cost
    }
    return 0.0
}
```

### Kubernetes Resource Dashboard

```go
import (
    "context"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func k8sResourceDashboard() {
    // Initialize Kubernetes client
    config, _ := clientcmd.BuildConfigFromFlags("", "~/.kube/config")
    clientset, _ := kubernetes.NewForConfig(config)

    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"Namespace", "Resource Type", "Name", "Status", "Age", "Resources"},
    }

    // Get pods
    pods, _ := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
    for _, pod := range pods.Items {
        age := time.Since(pod.CreationTimestamp.Time).Round(time.Hour)

        resources := fmt.Sprintf("CPU: %s, Memory: %s",
            pod.Spec.Containers[0].Resources.Requests.Cpu(),
            pod.Spec.Containers[0].Resources.Requests.Memory())

        output.AddContents(map[string]interface{}{
            "Namespace":     pod.Namespace,
            "Resource Type": "Pod",
            "Name":          pod.Name,
            "Status":        string(pod.Status.Phase),
            "Age":           age.String(),
            "Resources":     resources,
        })
    }

    // Get services
    services, _ := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
    for _, service := range services.Items {
        age := time.Since(service.CreationTimestamp.Time).Round(time.Hour)

        output.AddContents(map[string]interface{}{
            "Namespace":     service.Namespace,
            "Resource Type": "Service",
            "Name":          service.Name,
            "Status":        string(service.Spec.Type),
            "Age":           age.String(),
            "Resources":     fmt.Sprintf("Ports: %v", len(service.Spec.Ports)),
        })
    }

    output.Settings.SetOutputFormat("html")
    output.Settings.Title = "Kubernetes Resources"
    output.Settings.SortKey = "Namespace"
    output.Settings.OutputFile = "k8s-dashboard.html"
    output.Write()
}
```

## Monitoring and Alerting

### Performance Metrics Report

```go
type MetricsCollector struct {
    output *format.OutputArray
}

func NewMetricsCollector() *MetricsCollector {
    return &MetricsCollector{
        output: &format.OutputArray{
            Settings: format.NewOutputSettings(),
            Keys:     []string{"Service", "Metric", "Value", "Unit", "Threshold", "Status"},
        },
    }
}

func (mc *MetricsCollector) AddMetric(service, metric string, value float64, unit string, threshold float64) {
    status := "OK"
    if value > threshold {
        status = "WARNING"
    }
    if value > threshold*1.5 {
        status = "CRITICAL"
    }

    mc.output.AddContents(map[string]interface{}{
        "Service":   service,
        "Metric":    metric,
        "Value":     value,
        "Unit":      unit,
        "Threshold": threshold,
        "Status":    status,
    })
}

func (mc *MetricsCollector) GenerateReport() {
    mc.output.Settings.UseColors = true
    mc.output.Settings.UseEmoji = true
    mc.output.Settings.Title = "System Metrics Report"
    mc.output.Settings.SortKey = "Service"

    // Color-coded table output
    mc.output.Settings.SetOutputFormat("table")
    mc.output.Write()

    // Generate alert summary
    criticalCount := 0
    warningCount := 0

    for _, holder := range mc.output.Contents {
        status := holder.Contents["Status"].(string)
        switch status {
        case "CRITICAL":
            criticalCount++
        case "WARNING":
            warningCount++
        }
    }

    if criticalCount > 0 {
        fmt.Printf("\n%s %d critical alerts!\n",
            mc.output.Settings.StringFailure("🚨"), criticalCount)
    }
    if warningCount > 0 {
        fmt.Printf("%s %d warnings\n",
            mc.output.Settings.StringWarning("⚠️"), warningCount)
    }
}

func generateMetricsReport() {
    collector := NewMetricsCollector()

    // Simulate collecting metrics
    collector.AddMetric("web-server", "CPU Usage", 85.5, "%", 80.0)
    collector.AddMetric("web-server", "Memory Usage", 67.2, "%", 90.0)
    collector.AddMetric("database", "CPU Usage", 45.8, "%", 80.0)
    collector.AddMetric("database", "Disk Usage", 92.1, "%", 85.0)
    collector.AddMetric("cache", "Memory Usage", 78.9, "%", 90.0)

    collector.GenerateReport()
}
```

### Log Analysis Summary

```go
func analyzeLogFiles(logFiles []string) {
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"File", "Total Lines", "Errors", "Warnings", "Last Error", "Error Rate"},
    }

    for _, logFile := range logFiles {
        stats := analyzeLogFile(logFile)

        output.AddContents(map[string]interface{}{
            "File":        filepath.Base(logFile),
            "Total Lines": stats.TotalLines,
            "Errors":      stats.ErrorCount,
            "Warnings":    stats.WarningCount,
            "Last Error":  stats.LastError.Format("2006-01-02 15:04"),
            "Error Rate":  fmt.Sprintf("%.2f%%", stats.ErrorRate*100),
        })
    }

    output.Settings.SetOutputFormat("markdown")
    output.Settings.Title = "Log Analysis Summary"
    output.Settings.OutputFile = "log-analysis.md"
    output.Write()
}

type LogStats struct {
    TotalLines   int
    ErrorCount   int
    WarningCount int
    LastError    time.Time
    ErrorRate    float64
}

func analyzeLogFile(filename string) LogStats {
    // Implementation for parsing log files
    // Return statistics about the log file
    return LogStats{
        TotalLines:   1000,
        ErrorCount:   15,
        WarningCount: 45,
        LastError:    time.Now().Add(-2 * time.Hour),
        ErrorRate:    0.015,
    }
}
```

## Data Visualization

### Network Topology Diagram

```go
func generateNetworkTopology() {
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"Node", "Type", "ConnectedTo", "Bandwidth"},
    }

    // Network infrastructure data
    networkNodes := []map[string]interface{}{
        {"Node": "Internet", "Type": "external", "ConnectedTo": "Router", "Bandwidth": "1Gbps"},
        {"Node": "Router", "Type": "network", "ConnectedTo": "Firewall", "Bandwidth": "1Gbps"},
        {"Node": "Firewall", "Type": "security", "ConnectedTo": "Switch", "Bandwidth": "1Gbps"},
        {"Node": "Switch", "Type": "network", "ConnectedTo": "Web-Server,DB-Server", "Bandwidth": "1Gbps"},
        {"Node": "Web-Server", "Type": "server", "ConnectedTo": "DB-Server", "Bandwidth": "100Mbps"},
        {"Node": "DB-Server", "Type": "database", "ConnectedTo": "", "Bandwidth": ""},
    }

    for _, node := range networkNodes {
        output.AddContents(node)
    }

    // Generate Mermaid diagram
    output.Settings.AddFromToColumns("Node", "ConnectedTo")
    output.Settings.SetOutputFormat("mermaid")
    output.Settings.MermaidSettings.ChartType = "flowchart"
    output.Settings.MermaidSettings.Direction = "TB"
    output.Settings.Title = "Network Topology"
    output.Settings.OutputFile = "network-topology.mmd"
    output.Write()

    // Generate Draw.io diagram
    header := drawio.NewHeader("%Node%\n%Type%", "%Type%", "Type")
    header.SetLayout(drawio.LayoutVerticalFlow)

    connection := drawio.NewConnection()
    connection.From = "Node"
    connection.To = "ConnectedTo"
    connection.Label = "%Bandwidth%"
    header.AddConnection(connection)

    output.Settings.DrawIOHeader = header
    output.Settings.SetOutputFormat("drawio")
    output.Settings.OutputFile = "network-topology.csv"
    output.Write()
}
```

### Project Timeline (Gantt Chart)

```go
func generateProjectTimeline() {
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"Task", "Start Date", "Duration", "Status", "Assignee"},
    }

    tasks := []map[string]interface{}{
        {"Task": "Requirements Analysis", "Start Date": "2024-01-01", "Duration": "2w", "Status": "done", "Assignee": "Alice"},
        {"Task": "System Design", "Start Date": "2024-01-15", "Duration": "3w", "Status": "done", "Assignee": "Bob"},
        {"Task": "Backend Development", "Start Date": "2024-02-05", "Duration": "6w", "Status": "active", "Assignee": "Carol"},
        {"Task": "Frontend Development", "Start Date": "2024-02-19", "Duration": "4w", "Status": "active", "Assignee": "Dave"},
        {"Task": "Testing", "Start Date": "2024-03-18", "Duration": "2w", "Status": "crit", "Assignee": "Eve"},
        {"Task": "Deployment", "Start Date": "2024-04-01", "Duration": "1w", "Status": "", "Assignee": "Frank"},
    }

    for _, task := range tasks {
        output.AddContents(task)
    }

    // Configure Mermaid Gantt chart
    output.Settings.SetOutputFormat("mermaid")
    output.Settings.MermaidSettings = &mermaid.Settings{
        ChartType: "ganttchart",
        GanttSettings: &mermaid.GanttSettings{
            LabelColumn:     "Task",
            StartDateColumn: "Start Date",
            DurationColumn:  "Duration",
            StatusColumn:    "Status",
        },
    }
    output.Settings.Title = "Project Timeline"
    output.Settings.OutputFile = "project-timeline.mmd"
    output.Write()
}
```

## Web Application Integration

### REST API Response Formatter

```go
import (
    "net/http"
    "encoding/json"
    "github.com/gorilla/mux"
)

type APIServer struct {
    router *mux.Router
}

func NewAPIServer() *APIServer {
    server := &APIServer{
        router: mux.NewRouter(),
    }
    server.setupRoutes()
    return server
}

func (s *APIServer) setupRoutes() {
    s.router.HandleFunc("/api/users", s.usersHandler).Methods("GET")
    s.router.HandleFunc("/api/metrics", s.metricsHandler).Methods("GET")
}

func (s *APIServer) usersHandler(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters
    format := r.URL.Query().Get("format")
    if format == "" {
        format = "json"
    }

    // Create output
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"ID", "Name", "Email", "Role", "LastLogin"},
    }

    // Fetch users (simulated)
    users := getUsersFromDatabase()
    for _, user := range users {
        output.AddContents(user)
    }

    output.Settings.SetOutputFormat(format)
    output.Settings.Title = "User List"

    // Set appropriate content type and headers
    switch format {
    case "json":
        w.Header().Set("Content-Type", "application/json")
        // Return raw data for JSON
        json.NewEncoder(w).Encode(output.GetContentsMapRaw())
    case "csv":
        w.Header().Set("Content-Type", "text/csv")
        w.Header().Set("Content-Disposition", "attachment; filename=users.csv")
        output.AddToBuffer()
        w.Write(buffer.Bytes())
    case "html":
        w.Header().Set("Content-Type", "text/html")
        output.AddToBuffer()
        w.Write(output.bufferToHTML())
    default:
        http.Error(w, "Unsupported format", http.StatusBadRequest)
        return
    }
}

func getUsersFromDatabase() []map[string]interface{} {
    return []map[string]interface{}{
        {"ID": 1, "Name": "Alice", "Email": "alice@example.com", "Role": "Admin", "LastLogin": "2024-01-15"},
        {"ID": 2, "Name": "Bob", "Email": "bob@example.com", "Role": "User", "LastLogin": "2024-01-14"},
    }
}
```

### Database Query Results

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

func exportQueryResults(query string, outputFormat string) {
    db, err := sql.Open("postgres", "postgres://user:password@localhost/db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    rows, err := db.Query(query)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    // Get column names
    columns, _ := rows.Columns()

    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     columns,
    }

    // Process results
    for rows.Next() {
        values := make([]interface{}, len(columns))
        valuePtrs := make([]interface{}, len(columns))
        for i := range values {
            valuePtrs[i] = &values[i]
        }

        rows.Scan(valuePtrs...)

        record := make(map[string]interface{})
        for i, col := range columns {
            record[col] = values[i]
        }

        output.AddContents(record)
    }

    output.Settings.SetOutputFormat(outputFormat)
    output.Settings.Title = "Query Results"
    output.Settings.OutputFile = fmt.Sprintf("query_results.%s", outputFormat)
    output.Write()
}
```

## CLI Tools

### System Health Checker

```go
package main

import (
    "flag"
    "os"
    "github.com/ArjenSchwarz/go-output"
)

func main() {
    var (
        outputFormat = flag.String("format", "table", "Output format (table, json, csv, html)")
        outputFile   = flag.String("output", "", "Output file (default: stdout)")
        verbose      = flag.Bool("verbose", false, "Verbose output")
        useColors    = flag.Bool("colors", true, "Use colors in output")
    )
    flag.Parse()

    // Detect if we're in a terminal
    if !isTerminal() {
        *useColors = false
    }

    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
    }

    if *verbose {
        output.Keys = []string{"Component", "Status", "Message", "Response Time", "Last Check"}
    } else {
        output.Keys = []string{"Component", "Status", "Message"}
    }

    output.Settings.SetOutputFormat(*outputFormat)
    output.Settings.OutputFile = *outputFile
    output.Settings.UseColors = *useColors
    output.Settings.UseEmoji = *useColors
    output.Settings.Title = "System Health Check"

    // Perform health checks
    healthChecks := []func() map[string]interface{}{
        checkDatabaseHealth,
        checkAPIHealth,
        checkCacheHealth,
        checkDiskSpace,
        checkMemoryUsage,
    }

    for _, check := range healthChecks {
        result := check()
        output.AddContents(result)
    }

    output.Write()

    // Exit with appropriate code
    exitCode := 0
    for _, holder := range output.Contents {
        if status := holder.Contents["Status"].(string); status != "OK" {
            exitCode = 1
            break
        }
    }
    os.Exit(exitCode)
}

func checkDatabaseHealth() map[string]interface{} {
    // Simulate database health check
    return map[string]interface{}{
        "Component":     "Database",
        "Status":        "OK",
        "Message":       "Connection successful",
        "Response Time": "15ms",
        "Last Check":    time.Now().Format("15:04:05"),
    }
}

func checkAPIHealth() map[string]interface{} {
    return map[string]interface{}{
        "Component":     "API",
        "Status":        "WARNING",
        "Message":       "High response time",
        "Response Time": "850ms",
        "Last Check":    time.Now().Format("15:04:05"),
    }
}

func isTerminal() bool {
    fileInfo, _ := os.Stdout.Stat()
    return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
```

### File System Analyzer

```go
func analyzeFileSystem(rootPath string) {
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"Path", "Type", "Size", "Modified", "Permissions"},
    }

    err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil // Skip errors
        }

        fileType := "File"
        if info.IsDir() {
            fileType = "Directory"
        }

        size := formatBytes(info.Size())
        if info.IsDir() {
            size = "-"
        }

        output.AddContents(map[string]interface{}{
            "Path":        path,
            "Type":        fileType,
            "Size":        size,
            "Modified":    info.ModTime().Format("2006-01-02 15:04"),
            "Permissions": info.Mode().String(),
        })

        return nil
    })

    if err != nil {
        log.Fatal(err)
    }

    output.Settings.SetOutputFormat("table")
    output.Settings.Title = fmt.Sprintf("File System Analysis: %s", rootPath)
    output.Settings.SortKey = "Size"
    output.Write()
}

func formatBytes(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
```

## Data Processing Pipelines

### CSV Data Transformation

```go
func transformCSVData(inputFile, outputFile string) {
    // Read input CSV
    file, err := os.Open(inputFile)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        log.Fatal(err)
    }

    if len(records) == 0 {
        log.Fatal("No data found")
    }

    // First row is headers
    headers := records[0]

    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"Original Field", "Transformed Field", "Data Type", "Sample Value"},
    }

    // Analyze and transform each column
    for i, header := range headers {
        if i >= len(records[1]) {
            continue
        }

        sampleValue := records[1][i]
        dataType := detectDataType(sampleValue)
        transformedField := transformFieldName(header)

        output.AddContents(map[string]interface{}{
            "Original Field":    header,
            "Transformed Field": transformedField,
            "Data Type":         dataType,
            "Sample Value":      sampleValue,
        })
    }

    output.Settings.SetOutputFormat("markdown")
    output.Settings.Title = "CSV Data Transformation Analysis"
    output.Settings.OutputFile = outputFile
    output.Write()
}

func detectDataType(value string) string {
    if _, err := strconv.Atoi(value); err == nil {
        return "Integer"
    }
    if _, err := strconv.ParseFloat(value, 64); err == nil {
        return "Float"
    }
    if _, err := time.Parse("2006-01-02", value); err == nil {
        return "Date"
    }
    if len(value) > 100 {
        return "Text (Long)"
    }
    return "String"
}

func transformFieldName(field string) string {
    // Convert to snake_case
    return strings.ToLower(strings.ReplaceAll(field, " ", "_"))
}
```

## Error Reporting

### Application Error Summary

```go
type ErrorReporter struct {
    errors []ErrorEntry
    output *format.OutputArray
}

type ErrorEntry struct {
    Timestamp time.Time
    Level     string
    Component string
    Message   string
    Count     int
}

func NewErrorReporter() *ErrorReporter {
    return &ErrorReporter{
        errors: make([]ErrorEntry, 0),
        output: &format.OutputArray{
            Settings: format.NewOutputSettings(),
            Keys:     []string{"Time", "Level", "Component", "Message", "Count"},
        },
    }
}

func (er *ErrorReporter) AddError(level, component, message string) {
    er.errors = append(er.errors, ErrorEntry{
        Timestamp: time.Now(),
        Level:     level,
        Component: component,
        Message:   message,
        Count:     1,
    })
}

func (er *ErrorReporter) GenerateReport() {
    // Aggregate similar errors
    aggregated := er.aggregateErrors()

    for _, error := range aggregated {
        er.output.AddContents(map[string]interface{}{
            "Time":      error.Timestamp.Format("2006-01-02 15:04:05"),
            "Level":     error.Level,
            "Component": error.Component,
            "Message":   error.Message,
            "Count":     error.Count,
        })
    }

    er.output.Settings.UseColors = true
    er.output.Settings.Title = "Error Report"
    er.output.Settings.SortKey = "Count"
    er.output.Settings.SetOutputFormat("html")
    er.output.Settings.OutputFile = "error-report.html"
    er.output.Write()
}

func (er *ErrorReporter) aggregateErrors() []ErrorEntry {
    counts := make(map[string]*ErrorEntry)

    for _, err := range er.errors {
        key := fmt.Sprintf("%s-%s-%s", err.Level, err.Component, err.Message)
        if existing, exists := counts[key]; exists {
            existing.Count++
        } else {
            errCopy := err
            counts[key] = &errCopy
        }
    }

    result := make([]ErrorEntry, 0, len(counts))
    for _, err := range counts {
        result = append(result, *err)
    }

    return result
}
```

These examples demonstrate the versatility and power of the go-output library across various domains and use cases. Each example can be adapted and extended based on specific requirements.
