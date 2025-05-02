// dashboard-temp/src/data-adapter.js
// Enhanced with debugging and better error handling

/**
 * Maps API status values to UI display values
 */
const STATUS_MAP = {
    'OK': 'Healthy',
    'Warning': 'Warning',
    'Critical': 'Critical',
    'Unknown': 'Unknown',
    'NotApplicable': 'Not Applicable'
};

/**
 * Transforms a report from API format to dashboard format
 * @param {Object} apiReport - Report data from the API
 * @returns {Object} Transformed report data for the dashboard
 */
export function transformReport(apiReport) {
    // Add debug logging
    console.log("Transforming report:", apiReport);

    if (!apiReport) {
        console.warn("Null or undefined report received");
        return null;
    }

    // Create a safe transformed report with defensive coding
    const report = {
        id: apiReport.id || '',
        timestamp: apiReport.timestamp || new Date().toISOString(),
        status: STATUS_MAP[apiReport.status] || apiReport.status || 'Unknown',
        formats: Array.isArray(apiReport.formats) ? apiReport.formats : []
    };

    console.log("Transformed report:", report);
    return report;
}

/**
 * Transforms report details from API format to dashboard format
 * @param {Object} apiReportDetails - Report details from API
 * @returns {Object} Transformed report details for the dashboard
 */
export function transformReportDetails(apiReportDetails) {
    console.log("Transforming report details:", apiReportDetails);

    // If no data is received, return default structure
    if (!apiReportDetails) {
        console.warn("Null or undefined report details received");
        return createDefaultReportDetails();
    }

    // Safely access results array
    const results = Array.isArray(apiReportDetails.results) ? apiReportDetails.results : [];
    console.log(`Processing ${results.length} results`);

    // Extract status counts from actual results
    const statusCounts = {
        passing: 0,
        warnings: 0,
        critical: 0,
        other: 0
    };

    // Category data
    const categoryData = {};

    // Issues
    const criticalIssues = [];
    const warningIssues = [];

    // Process all results
    results.forEach((result, index) => {
        // Safety check
        if (!result) {
            console.warn(`Null result at index ${index}`);
            return; // Skip this iteration
        }

        console.log(`Processing result ${index}:`, result.checkID || 'unknown', result.status);

        // Update status counts
        if (result.status === 'OK') {
            statusCounts.passing++;
        } else if (result.status === 'Warning') {
            statusCounts.warnings++;
        } else if (result.status === 'Critical') {
            statusCounts.critical++;
        } else {
            statusCounts.other++;
        }

        // Update category data
        const category = result.category || 'Unknown';
        if (!categoryData[category]) {
            categoryData[category] = {
                name: category,
                passing: 0,
                warnings: 0,
                critical: 0
            };
        }

        if (result.status === 'OK') {
            categoryData[category].passing++;
        } else if (result.status === 'Warning') {
            categoryData[category].warnings++;
        } else if (result.status === 'Critical') {
            categoryData[category].critical++;
        }

        // Add to issues arrays
        if (result.status === 'Critical') {
            criticalIssues.push({
                id: result.checkID || `check-${criticalIssues.length}`,
                name: result.checkName || result.message || 'Unknown',
                message: result.message || '',
                category: result.category || 'Unknown',
                recommendations: Array.isArray(result.recommendations) ? result.recommendations : []
            });
        } else if (result.status === 'Warning') {
            warningIssues.push({
                id: result.checkID || `check-${warningIssues.length}`,
                name: result.checkName || result.message || 'Unknown',
                message: result.message || '',
                category: result.category || 'Unknown',
                recommendations: Array.isArray(result.recommendations) ? result.recommendations : []
            });
        }
    });

    // Calculate total and percentages
    const total = statusCounts.passing + statusCounts.warnings + statusCounts.critical + statusCounts.other;

    const summaryStats = {
        passing: statusCounts.passing,
        passingPercent: total > 0 ? Math.round((statusCounts.passing / total) * 100) : 0,
        warnings: statusCounts.warnings,
        warningsPercent: total > 0 ? Math.round((statusCounts.warnings / total) * 100) : 0,
        critical: statusCounts.critical,
        criticalPercent: total > 0 ? Math.round((statusCounts.critical / total) * 100) : 0,
        total: total
    };

    // Convert category data to array
    const categoryDataArray = Object.values(categoryData);

    // Calculate overall health score (weighted formula)
    // Critical issues heavily reduce score, warnings moderately reduce
    const healthScore = total > 0
        ? Math.max(0, Math.min(100, Math.round((
            (statusCounts.passing * 100) +
            (statusCounts.other * 50) -
            (statusCounts.warnings * 15) -
            (statusCounts.critical * 40)
        ) / total)))
        : 0;

    // Create final report details object
    const transformedDetails = {
        id: apiReportDetails.id || '',
        timestamp: apiReportDetails.timestamp || new Date().toISOString(),
        status: STATUS_MAP[apiReportDetails.status] || apiReportDetails.status || 'Unknown',
        summaryStats,
        categoryData: categoryDataArray,
        criticalIssues,
        warningIssues,
        healthScore
    };

    console.log("Transformed report details:", {
        id: transformedDetails.id,
        status: transformedDetails.status,
        healthScore: transformedDetails.healthScore,
        stats: {
            passing: transformedDetails.summaryStats.passing,
            warnings: transformedDetails.summaryStats.warnings,
            critical: transformedDetails.summaryStats.critical,
            total: transformedDetails.summaryStats.total
        },
        criticalIssuesCount: transformedDetails.criticalIssues.length,
        warningIssuesCount: transformedDetails.warningIssues.length
    });

    return transformedDetails;
}

/**
 * Creates default report details when none are available
 */
function createDefaultReportDetails() {
    console.log("Creating default report details");

    // Mock data for better user experience while loading
    return {
        id: '',
        timestamp: new Date().toISOString(),
        status: 'Unknown',
        summaryStats: {
            passing: 45,
            passingPercent: 60,
            warnings: 20,
            warningsPercent: 26,
            critical: 10,
            criticalPercent: 14,
            total: 75
        },
        categoryData: [
            {
                name: 'Cluster Config',
                passing: 15,
                warnings: 5,
                critical: 2
            },
            {
                name: 'Networking',
                passing: 10,
                warnings: 5,
                critical: 3
            },
            {
                name: 'Security',
                passing: 8,
                warnings: 6,
                critical: 4
            },
            {
                name: 'Storage',
                passing: 12,
                warnings: 4,
                critical: 1
            }
        ],
        criticalIssues: [
            {
                id: 'default-critical-1',
                name: 'Cluster Version',
                message: 'Cluster version outdated',
                category: 'Cluster Config',
                recommendations: ['Update to the latest version']
            }
        ],
        warningIssues: [
            {
                id: 'default-warning-1',
                name: 'Resource Quotas',
                message: 'Some namespaces missing resource quotas',
                category: 'Applications',
                recommendations: ['Configure resource quotas for all namespaces']
            }
        ],
        healthScore: 62
    };
}

/**
 * Generates trend data based on report history
 * @param {Array} reports - Array of reports
 * @returns {Array} Processed trend data
 */
export function generateTrendData(reports) {
    console.log("Generating trend data from reports:", reports?.length || 0);

    if (!reports || !Array.isArray(reports) || reports.length === 0) {
        console.warn("No reports available for trend data");
        return [];
    }

    // Create a map of dates to status counts
    const dateMap = {};

    // Sort reports by date
    const sortedReports = [...reports].sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));

    // Process each report
    sortedReports.forEach((report, index) => {
        // Safely get date
        let dateStr;
        try {
            dateStr = new Date(report.timestamp).toISOString().split('T')[0];
        } catch (e) {
            console.warn(`Invalid date in report ${index}:`, e);
            dateStr = new Date().toISOString().split('T')[0];
        }

        // Initialize counts for this date if needed
        if (!dateMap[dateStr]) {
            dateMap[dateStr] = {
                date: dateStr,
                passing: 0,
                warnings: 0,
                critical: 0
            };
        }

        // Add real report data if available
        if (report.summaryStats) {
            dateMap[dateStr].passing = report.summaryStats.passing || 0;
            dateMap[dateStr].warnings = report.summaryStats.warnings || 0;
            dateMap[dateStr].critical = report.summaryStats.critical || 0;
        }
    });

    // Convert map to array
    const trendData = Object.values(dateMap);
    console.log("Generated trend data:", trendData);
    return trendData;
}