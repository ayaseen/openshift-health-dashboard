// app/server/utils/report_parser.go
package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"dashboard/types"
)

// ParseAsciiDocExecutiveSummary parses an AsciiDoc file and extracts the executive summary
func ParseAsciiDocExecutiveSummary(filePath string) (*types.ReportSummary, error) {
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Look for executive summary section
	var summaryLines []string
	inSummary := false
	foundSummary := false

	for _, line := range lines {
		// Check for executive summary section start
		if strings.Contains(line, "Red Hat Consulting conducted") ||
			strings.Contains(line, "Overall Cluster Health:") {
			inSummary = true
			foundSummary = true
			summaryLines = append(summaryLines, line)
			continue
		}

		// If in summary section, keep collecting lines
		if inSummary {
			summaryLines = append(summaryLines, line)
		}
	}

	// If we couldn't find an executive summary, try to extract whatever data we can
	if !foundSummary {
		return extractBasicReport(lines)
	}

	// Initialize the summary
	summary := &types.ReportSummary{
		ClusterName:        ExtractClusterName(summaryLines),
		CustomerName:       ExtractCustomerName(summaryLines),
		OverallScore:       ExtractOverallScore(summaryLines),
		ScoreInfra:         ExtractCategoryScore(summaryLines, "Infrastructure Setup"),
		ScoreGovernance:    ExtractCategoryScore(summaryLines, "Policy Governance"),
		ScoreCompliance:    ExtractCategoryScore(summaryLines, "Compliance Benchmarking"),
		ScoreMonitoring:    ExtractCategoryScore(summaryLines, "Central Monitoring"),
		ScoreBuildSecurity: ExtractCategoryScore(summaryLines, "Build/Deploy Security"),
	}

	// Extract category descriptions
	summary.InfraDescription = ExtractCategoryDescription(summaryLines, "Infrastructure Setup")
	summary.GovernanceDescription = ExtractCategoryDescription(summaryLines, "Policy Governance")
	summary.ComplianceDescription = ExtractCategoryDescription(summaryLines, "Compliance Benchmarking")
	summary.MonitoringDescription = ExtractCategoryDescription(summaryLines, "Central Monitoring")
	summary.BuildSecurityDescription = ExtractCategoryDescription(summaryLines, "Build/Deploy Security")

	// Extract action items
	summary.ItemsRequired = ExtractActionItems(summaryLines, "Changes Required")
	summary.ItemsRecommended = ExtractActionItems(summaryLines, "Changes Recommended")
	summary.ItemsAdvisory = ExtractActionItems(summaryLines, "Advisory Actions")

	return summary, nil
}

// extractBasicReport tries to extract basic information from any health check report
func extractBasicReport(lines []string) (*types.ReportSummary, error) {
	// Initialize with default values
	summary := &types.ReportSummary{
		ClusterName:        "Unknown Cluster",
		CustomerName:       "Unknown Customer",
		OverallScore:       CalculateOverallScore(lines),
		ScoreInfra:         ExtractGeneralCategoryScore(lines, "Cluster", "Infrastructure", "Infra"),
		ScoreGovernance:    ExtractGeneralCategoryScore(lines, "Security", "Policy", "Governance"),
		ScoreCompliance:    ExtractGeneralCategoryScore(lines, "Compliance", "Benchmarking"),
		ScoreMonitoring:    ExtractGeneralCategoryScore(lines, "Monitoring", "Logging", "Observability"),
		ScoreBuildSecurity: ExtractGeneralCategoryScore(lines, "Build", "Deploy", "Application", "App"),
	}

	// Extract issues by severity
	summary.ItemsRequired = ExtractIssuesBySeverity(lines, "Critical", "Required", "Error")
	summary.ItemsRecommended = ExtractIssuesBySeverity(lines, "Warning", "Recommended")
	summary.ItemsAdvisory = ExtractIssuesBySeverity(lines, "Advisory", "Info", "Information")

	// Generate basic descriptions based on scores
	summary.InfraDescription = GenerateDescription("Infrastructure", summary.ScoreInfra)
	summary.GovernanceDescription = GenerateDescription("Policy Governance", summary.ScoreGovernance)
	summary.ComplianceDescription = GenerateDescription("Compliance", summary.ScoreCompliance)
	summary.MonitoringDescription = GenerateDescription("Monitoring and Logging", summary.ScoreMonitoring)
	summary.BuildSecurityDescription = GenerateDescription("Build/Deploy Security", summary.ScoreBuildSecurity)

	return summary, nil
}
