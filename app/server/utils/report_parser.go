// app/server/utils/report_parser.go
package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ayaseen/openshift-health-dashboard/app/server/types"
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

	// Look for the summary information
	var summaryLines []string
	inSummary := false
	foundSummary := false

	// First pass: collect all lines as potential summary content
	for _, line := range lines {
		// Check for executive summary section start
		if (strings.Contains(line, "Red Hat Consulting conducted") ||
			strings.Contains(line, "OpenShift Health Check") ||
			strings.Contains(line, "Overall Cluster Health")) && !inSummary {
			inSummary = true
			foundSummary = true
		}

		// If in summary section, keep collecting lines
		if inSummary || !foundSummary {
			summaryLines = append(summaryLines, line)
		}
	}

	// If we couldn't find an explicit executive summary, we'll use all lines
	if !foundSummary {
		summaryLines = lines
	}

	// Extract scores and data
	overallScore := ExtractOverallScore(summaryLines)

	// Extract required changes
	itemsRequired := ExtractRequiredChanges(summaryLines)
	if len(itemsRequired) == 0 {
		// Fall back to ActionItems extraction if the direct method fails
		itemsRequired = ExtractActionItems(summaryLines, "Changes Required")
	}

	// Extract recommended changes
	itemsRecommended := ExtractRecommendedChanges(summaryLines)
	if len(itemsRecommended) == 0 {
		// Fall back to ActionItems extraction if the direct method fails
		itemsRecommended = ExtractActionItems(summaryLines, "Changes Recommended")
	}

	// Extract advisory actions
	itemsAdvisory := ExtractAdvisoryActions(summaryLines)
	if len(itemsAdvisory) == 0 {
		// Fall back to ActionItems extraction if the direct method fails
		itemsAdvisory = ExtractActionItems(summaryLines, "Advisory Actions")
	}

	// Initialize the summary with extracted data
	summary := &types.ReportSummary{
		ClusterName:              ExtractClusterName(summaryLines),
		CustomerName:             ExtractCustomerName(summaryLines),
		OverallScore:             overallScore,
		ScoreInfra:               ExtractCategoryScore(summaryLines, "Infrastructure Setup"),
		ScoreGovernance:          ExtractCategoryScore(summaryLines, "Policy Governance"),
		ScoreCompliance:          ExtractCategoryScore(summaryLines, "Compliance Benchmarking"),
		ScoreMonitoring:          ExtractCategoryScore(summaryLines, "Central Monitoring"),
		ScoreBuildSecurity:       ExtractCategoryScore(summaryLines, "Build/Deploy Security"),
		InfraDescription:         ExtractCategoryDescription(summaryLines, "Infrastructure Setup"),
		GovernanceDescription:    ExtractCategoryDescription(summaryLines, "Policy Governance"),
		ComplianceDescription:    ExtractCategoryDescription(summaryLines, "Compliance Benchmarking"),
		MonitoringDescription:    ExtractCategoryDescription(summaryLines, "Central Monitoring"),
		BuildSecurityDescription: ExtractCategoryDescription(summaryLines, "Build/Deploy Security"),
		ItemsRequired:            itemsRequired,
		ItemsRecommended:         itemsRecommended,
		ItemsAdvisory:            itemsAdvisory,
	}

	// If no items were found, try to extract them by severity
	if len(summary.ItemsRequired) == 0 {
		summary.ItemsRequired = ExtractIssuesBySeverity(summaryLines, "Critical", "Required", "Error")
	}
	if len(summary.ItemsRecommended) == 0 {
		summary.ItemsRecommended = ExtractIssuesBySeverity(summaryLines, "Warning", "Recommended")
	}
	if len(summary.ItemsAdvisory) == 0 {
		summary.ItemsAdvisory = ExtractIssuesBySeverity(summaryLines, "Advisory", "Info", "Information")
	}

	// If category scores weren't found, try to extract them using alternative methods
	if summary.ScoreInfra == 0 {
		summary.ScoreInfra = ExtractGeneralCategoryScore(summaryLines, "Infrastructure", "Infra", "Setup")
	}
	if summary.ScoreGovernance == 0 {
		summary.ScoreGovernance = ExtractGeneralCategoryScore(summaryLines, "Governance", "Policy", "Security")
	}
	if summary.ScoreCompliance == 0 {
		summary.ScoreCompliance = ExtractGeneralCategoryScore(summaryLines, "Compliance", "Benchmark")
	}
	if summary.ScoreMonitoring == 0 {
		summary.ScoreMonitoring = ExtractGeneralCategoryScore(summaryLines, "Monitoring", "Logging", "Observability")
	}
	if summary.ScoreBuildSecurity == 0 {
		summary.ScoreBuildSecurity = ExtractGeneralCategoryScore(summaryLines, "Build", "Deploy", "App", "Application")
	}

	return summary, nil
}

// extractBasicReport tries to extract basic information from any health check report
func extractBasicReport(lines []string) (*types.ReportSummary, error) {
	// Initialize with default values
	summary := &types.ReportSummary{
		ClusterName:        "Unknown Cluster",
		CustomerName:       "Unknown Customer",
		OverallScore:       CalculateScoreFromStatusCounts(lines),
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
