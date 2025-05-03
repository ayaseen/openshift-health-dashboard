// app/server/types/types.go
package types

// ReportSummary represents the extracted summary data from an AsciiDoc report
type ReportSummary struct {
	ClusterName              string   `json:"clusterName"`
	CustomerName             string   `json:"customerName"`
	OverallScore             float64  `json:"overallScore"`
	ScoreInfra               int      `json:"scoreInfra"`
	ScoreGovernance          int      `json:"scoreGovernance"`
	ScoreCompliance          int      `json:"scoreCompliance"`
	ScoreMonitoring          int      `json:"scoreMonitoring"`
	ScoreBuildSecurity       int      `json:"scoreBuildSecurity"`
	InfraDescription         string   `json:"infraDescription"`
	GovernanceDescription    string   `json:"governanceDescription"`
	ComplianceDescription    string   `json:"complianceDescription"`
	MonitoringDescription    string   `json:"monitoringDescription"`
	BuildSecurityDescription string   `json:"buildSecurityDescription"`
	ItemsRequired            []string `json:"itemsRequired"`
	ItemsRecommended         []string `json:"itemsRecommended"`
	ItemsAdvisory            []string `json:"itemsAdvisory"`
	NoChangeCount            int      `json:"noChangeCount"`
}

// Category represents a category in the health check report
type Category struct {
	Name        string
	Score       int
	Description string
}

// Status represents the status of a health check
type Status string

const (
	// StatusOK indicates everything is working correctly
	StatusOK Status = "OK"

	// StatusWarning indicates a potential issue that should be addressed
	StatusWarning Status = "Warning"

	// StatusCritical indicates a critical issue that requires immediate attention
	StatusCritical Status = "Critical"

	// StatusUnknown indicates the status could not be determined
	StatusUnknown Status = "Unknown"

	// StatusNotApplicable indicates the check does not apply to this environment
	StatusNotApplicable Status = "NotApplicable"
)

// ResultKey represents the level of importance for a result
type ResultKey string

const (
	// ResultKeyNoChange indicates no changes are needed
	ResultKeyNoChange ResultKey = "nochange"

	// ResultKeyRecommended indicates changes are recommended
	ResultKeyRecommended ResultKey = "recommended"

	// ResultKeyRequired indicates changes are required
	ResultKeyRequired ResultKey = "required"

	// ResultKeyAdvisory indicates additional information
	ResultKeyAdvisory ResultKey = "advisory"

	// ResultKeyNotApplicable indicates the check does not apply
	ResultKeyNotApplicable ResultKey = "na"

	// ResultKeyEvaluate indicates the result needs evaluation
	ResultKeyEvaluate ResultKey = "eval"
)
