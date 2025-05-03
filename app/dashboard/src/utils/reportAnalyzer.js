// app/dashboard/src/utils/reportAnalyzer.js

/**
 * Report Analyzer Utility
 *
 * This utility provides advanced analysis functions for OpenShift health check reports.
 * It helps identify patterns, extract insights, and generate actionable recommendations.
 */

/**
 * Analyzes an OpenShift health check report and extracts additional insights
 * @param {Object} reportData - The parsed report data
 * @returns {Object} Analysis results including insights and recommendations
 */
export const analyzeReport = (reportData) => {
    if (!reportData) {
        return null;
    }

    return {
        summary: generateSummary(reportData),
        insights: extractInsights(reportData),
        recommendations: generateRecommendations(reportData),
        securityStatus: analyzeSecurityStatus(reportData),
        performanceInsights: analyzePerformance(reportData)
    };
};

/**
 * Generates a concise summary of the report
 */
const generateSummary = (reportData) => {
    const requiredCount = reportData.itemsRequired?.length || 0;
    const recommendedCount = reportData.itemsRecommended?.length || 0;
    const advisoryCount = reportData.itemsAdvisory?.length || 0;

    const categoryScores = [
        { name: 'Infrastructure', score: reportData.scoreInfra || 0 },
        { name: 'Governance', score: reportData.scoreGovernance || 0 },
        { name: 'Compliance', score: reportData.scoreCompliance || 0 },
        { name: 'Monitoring', score: reportData.scoreMonitoring || 0 },
        { name: 'Build/Deploy', score: reportData.scoreBuildSecurity || 0 }
    ];

    // Sort to find highest and lowest scoring categories
    categoryScores.sort((a, b) => b.score - a.score);
    const highestCategory = categoryScores[0];
    const lowestCategory = categoryScores[categoryScores.length - 1];

    // Calculate average category score
    const avgCategoryScore = categoryScores.reduce((sum, cat) => sum + cat.score, 0) / categoryScores.length;

    return {
        totalIssues: requiredCount + recommendedCount + advisoryCount,
        criticalIssuesCount: requiredCount,
        healthStatus: getHealthStatus(reportData.overallScore),
        highestCategory,
        lowestCategory,
        avgCategoryScore
    };
};

/**
 * Extracts specific insights from the report data
 */
const extractInsights = (reportData) => {
    // Extract issue categories
    const issueCategories = {};

    // Function to categorize an issue
    const categorizeIssue = (item) => {
        const categoryKeywords = {
            'Security': ['security', 'rbac', 'policy', 'user', 'kubeadmin', 'access', 'encryption', 'secret'],
            'Networking': ['network', 'ingress', 'route', 'dns', 'firewall', 'proxy', 'load balancer'],
            'Storage': ['storage', 'persistent', 'volume', 'pv', 'pvc', 'block', 'file'],
            'Configuration': ['config', 'setup', 'install', 'version', 'operator'],
            'Resources': ['resource', 'quota', 'limit', 'cpu', 'memory', 'node', 'capacity'],
            'Monitoring': ['monitor', 'logging', 'alert', 'metric', 'prometheus', 'grafana']
        };

        const lowerItem = item.toLowerCase();

        for (const [category, keywords] of Object.entries(categoryKeywords)) {
            if (keywords.some(keyword => lowerItem.includes(keyword))) {
                if (!issueCategories[category]) {
                    issueCategories[category] = 1;
                } else {
                    issueCategories[category]++;
                }
                return;
            }
        }

        // Default category if no match
        if (!issueCategories['Other']) {
            issueCategories['Other'] = 1;
        } else {
            issueCategories['Other']++;
        }
    };

    // Process each issue
    reportData.itemsRequired?.forEach(categorizeIssue);
    reportData.itemsRecommended?.forEach(categorizeIssue);

    // Analyze security-specific items
    const securityItems = [];
    const hasKubeadminIssue = reportData.itemsRequired?.some(item =>
        item.toLowerCase().includes('kubeadmin') || item.toLowerCase().includes('kube-admin')
    ) || false;

    if (hasKubeadminIssue) {
        securityItems.push({
            name: 'Kubeadmin Removal',
            priority: 'High',
            impact: 'Critical security vulnerability if compromised'
        });
    }

    const hasNetworkPolicyIssue = reportData.itemsRecommended?.some(item =>
        item.toLowerCase().includes('network policy')
    ) || false;

    if (hasNetworkPolicyIssue) {
        securityItems.push({
            name: 'Network Policies',
            priority: 'Medium',
            impact: 'Potential lateral movement within cluster'
        });
    }

    // Get top categories as an array of {name, count}
    const topIssueCategories = Object.entries(issueCategories)
        .sort((a, b) => b[1] - a[1])
        .slice(0, 3)
        .map(([name, count]) => ({ name, count }));

    return {
        issuesByCategory: issueCategories,
        topIssueCategories,
        securityItems,
        hasKubeadminIssue,
        hasNetworkPolicyIssue
    };
};

/**
 * Generates actionable recommendations based on report data
 */
const generateRecommendations = (reportData) => {
    const recommendations = [];
    const requiredCount = reportData.itemsRequired?.length || 0;

    // Critical security recommendations first
    if (reportData.itemsRequired?.some(item => item.toLowerCase().includes('kubeadmin'))) {
        recommendations.push({
            title: 'Remove kubeadmin user immediately',
            description: 'This default admin account should be removed after setting up alternative admin access.',
            steps: [
                'Ensure other cluster-admin users exist',
                'Run: oc delete secret kubeadmin -n kube-system'
            ],
            priority: 'critical',
            category: 'Security'
        });
    }

    // Network policy recommendations
    if (reportData.itemsRecommended?.some(item => item.toLowerCase().includes('network policy'))) {
        recommendations.push({
            title: 'Implement network policies',
            description: 'Network policies provide micro-segmentation within the cluster.',
            steps: [
                'Define policies for each namespace',
                'Limit ingress/egress traffic between namespaces',
                'Follow zero-trust principles'
            ],
            priority: 'high',
            category: 'Networking'
        });
    }

    // Monitoring recommendations
    if (reportData.scoreMonitoring < 80 ||
        reportData.itemsRecommended?.some(item => item.toLowerCase().includes('monitor'))) {
        recommendations.push({
            title: 'Enhance monitoring configuration',
            description: 'Improve monitoring stack to ensure proper visibility and alerting.',
            steps: [
                'Configure persistent storage for Prometheus',
                'Set up alert forwarding to external systems',
                'Enable user workload monitoring'
            ],
            priority: 'medium',
            category: 'Monitoring'
        });
    }

    // Resource management recommendations
    if (reportData.itemsRecommended?.some(item =>
        item.toLowerCase().includes('limitrange') ||
        item.toLowerCase().includes('resource quota'))) {
        recommendations.push({
            title: 'Implement resource controls',
            description: 'Ensure fair resource allocation and prevent resource exhaustion.',
            steps: [
                'Configure LimitRange in each namespace',
                'Set up ResourceQuotas for CPU, memory and storage',
                'Apply defaults for resource requests and limits'
            ],
            priority: 'medium',
            category: 'Resources'
        });
    }

    // Cluster update recommendations
    if (reportData.itemsRecommended?.some(item => item.toLowerCase().includes('cluster version'))) {
        recommendations.push({
            title: 'Update cluster version',
            description: 'Keep cluster updated to benefit from security fixes and new features.',
            steps: [
                'Review release notes for the latest version',
                'Test upgrade in a non-production environment',
                'Schedule maintenance window for the update',
                'Follow standard upgrade process'
            ],
            priority: 'medium',
            category: 'Maintenance'
        });
    }

    // Application health recommendations
    if (reportData.itemsRecommended?.some(item => item.toLowerCase().includes('probe'))) {
        recommendations.push({
            title: 'Implement application health probes',
            description: 'Configure readiness and liveness probes for improved application reliability.',
            steps: [
                'Add appropriate readiness probes to all deployments',
                'Configure liveness probes with appropriate thresholds',
                'Consider adding startup probes for slow-starting applications'
            ],
            priority: 'medium',
            category: 'Applications'
        });
    }

    return recommendations;
};

/**
 * Analyzes the security posture based on the report
 */
const analyzeSecurityStatus = (reportData) => {
    // Count security-related issues
    const securityIssues = {
        required: 0,
        recommended: 0
    };

    const securityKeywords = [
        'security', 'rbac', 'policy', 'user', 'kubeadmin', 'access',
        'encryption', 'secret', 'auth', 'identity', 'certificate'
    ];

    reportData.itemsRequired?.forEach(item => {
        if (securityKeywords.some(keyword => item.toLowerCase().includes(keyword))) {
            securityIssues.required++;
        }
    });

    reportData.itemsRecommended?.forEach(item => {
        if (securityKeywords.some(keyword => item.toLowerCase().includes(keyword))) {
            securityIssues.recommended++;
        }
    });

    // Determine security status
    let status = 'good';
    let riskLevel = 'low';

    if (securityIssues.required > 0) {
        status = 'critical';
        riskLevel = 'high';
    } else if (securityIssues.recommended > 2) {
        status = 'warning';
        riskLevel = 'medium';
    }

    return {
        status,
        riskLevel,
        securityIssues,
        encryptionEnabled: !reportData.itemsRecommended?.some(item =>
            item.toLowerCase().includes('etcd encryption')
        )
    };
};

/**
 * Analyzes performance aspects of the cluster
 */
const analyzePerformance = (reportData) => {
    // Look for performance-related items
    const performanceIssues = reportData.itemsRecommended?.filter(item =>
        item.toLowerCase().includes('performance') ||
        item.toLowerCase().includes('resource') ||
        item.toLowerCase().includes('cpu') ||
        item.toLowerCase().includes('memory') ||
        item.toLowerCase().includes('storage')
    ) || [];

    return {
        hasPerformanceIssues: performanceIssues.length > 0,
        performanceIssues,
        resourceUtilization: 'normal', // Default, since detailed metrics are not available
        recommendations: performanceIssues.length > 0 ? [
            'Review resource allocations and limits',
            'Consider infrastructure scaling if consistently high utilization',
            'Implement horizontal pod autoscaling for dynamic workloads'
        ] : []
    };
};

/**
 * Gets the overall health status label
 */
const getHealthStatus = (score) => {
    if (score >= 90) return 'Excellent';
    if (score >= 75) return 'Good';
    if (score >= 60) return 'Fair';
    if (score >= 40) return 'Poor';
    return 'Critical';
};

/**
 * Determines if a specific issue needs attention
 * @param {string} issue - The issue description
 * @returns {boolean} Whether this is an important issue
 */
export const isHighPriorityIssue = (issue) => {
    const highPriorityKeywords = [
        'kubeadmin', 'security', 'critical', 'vulnerability',
        'exposed', 'breach', 'outdated', 'deprecated'
    ];

    return highPriorityKeywords.some(keyword =>
        issue.toLowerCase().includes(keyword)
    );
};

/**
 * Gets a simple textual summary of the report
 */
export const getReportSummary = (reportData) => {
    if (!reportData) return '';

    const analysis = analyzeReport(reportData);
    const requiredCount = reportData.itemsRequired?.length || 0;

    let summary = `Cluster health is ${getHealthStatus(reportData.overallScore)} (${reportData.overallScore}%). `;

    if (requiredCount > 0) {
        summary += `There ${requiredCount === 1 ? 'is' : 'are'} ${requiredCount} critical issue${requiredCount === 1 ? '' : 's'} requiring immediate attention. `;
    } else {
        summary += 'No critical issues were found. ';
    }

    if (analysis.insights.hasKubeadminIssue) {
        summary += 'The kubeadmin user should be removed as a security best practice. ';
    }

    summary += `The strongest category is ${analysis.summary.highestCategory.name} (${analysis.summary.highestCategory.score}%) and the weakest is ${analysis.summary.lowestCategory.name} (${analysis.summary.lowestCategory.score}%).`;

    return summary;
};

export default {
    analyzeReport,
    getReportSummary,
    isHighPriorityIssue
};