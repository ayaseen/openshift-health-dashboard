// app/dashboard/src/utils/scoreUtils.js

// Get color based on score value
export const getScoreColor = (score) => {
    if (score >= 90) return '#10B981'; // Excellent - Green
    if (score >= 75) return '#34D399'; // Good - Light Green
    if (score >= 60) return '#F59E0B'; // Fair - Amber
    if (score >= 40) return '#FB923C'; // Poor - Orange
    return '#EF4444'; // Critical - Red
};

// Get rating text based on score value
export const getScoreRating = (score) => {
    if (score >= 90) return 'Excellent';
    if (score >= 75) return 'Good';
    if (score >= 60) return 'Fair';
    if (score >= 40) return 'Poor';
    return 'Critical';
};

// Get health assessment text based on score value
export const getHealthAssessment = (score) => {
    if (score >= 90) return 'Your cluster is highly optimized and follows all best practices.';
    if (score >= 75) return 'Your cluster is well configured with only minor improvements needed.';
    if (score >= 60) return 'Your cluster is generally healthy but has several areas needing attention.';
    if (score >= 40) return 'Your cluster needs significant improvements to meet best practices.';
    return 'Your cluster requires urgent attention to critical configuration issues.';
};

// Format score to remove excess decimals
export const formatScore = (score) => {
    const numericScore = parseFloat(score);
    if (isNaN(numericScore)) return '0';

    return Number.isInteger(numericScore)
        ? Math.round(numericScore)
        : parseFloat(numericScore.toFixed(1));
};

// Get category description based on name and score
export const getCategoryDescription = (categoryName, score) => {
    if (score >= 90) {
        return `${categoryName} is excellent with best practices in place.`;
    } else if (score >= 80) {
        return `${categoryName} is well-configured with only minor improvements needed.`;
    } else if (score >= 70) {
        return `${categoryName} meets most requirements but has some areas that could be improved.`;
    } else if (score >= 60) {
        return `${categoryName} has several areas that need attention to meet best practices.`;
    } else {
        return `${categoryName} requires significant improvements to ensure stability and security.`;
    }
};

// Get severity level object with color and label
export const getSeverityLevel = (type) => {
    switch (type.toLowerCase()) {
        case 'required':
            return { color: '#EF4444', label: 'Required', bgColor: '#FEF2F2' };
        case 'recommended':
            return { color: '#F59E0B', label: 'Recommended', bgColor: '#FFFBEB' };
        case 'advisory':
            return { color: '#3B82F6', label: 'Advisory', bgColor: '#EFF6FF' };
        default:
            return { color: '#6B7280', label: 'Info', bgColor: '#F9FAFB' };
    }
};

// Generate customized health action recommendations
export const getHealthActionRecommendations = (reportData) => {
    const requiredCount = reportData.itemsRequired?.length || 0;
    const recommendedCount = reportData.itemsRecommended?.length || 0;
    const overallScore = reportData.overallScore;

    let recommendations = [];

    // Critical actions for required items
    if (requiredCount > 0) {
        recommendations.push({
            priority: 'high',
            title: 'Address critical security issues',
            description: `Fix ${requiredCount} required change${requiredCount > 1 ? 's' : ''} immediately to maintain cluster security and stability.`
        });
    }

    // Recommendations based on score
    if (overallScore < 75) {
        recommendations.push({
            priority: 'medium',
            title: 'Improve overall cluster health',
            description: 'Schedule a comprehensive cluster review to address multiple configuration issues.'
        });
    }

    // Category specific recommendations
    if (reportData.scoreInfra < 70) {
        recommendations.push({
            priority: 'medium',
            title: 'Infrastructure optimization needed',
            description: 'Review infrastructure setup to improve cluster stability and performance.'
        });
    }

    if (reportData.scoreGovernance < 70) {
        recommendations.push({
            priority: 'medium',
            title: 'Enhance policy governance',
            description: 'Strengthen security policies and access controls to improve governance.'
        });
    }

    if (reportData.scoreMonitoring < 70) {
        recommendations.push({
            priority: 'medium',
            title: 'Improve monitoring capabilities',
            description: 'Enhance monitoring and alerting to ensure better operational visibility.'
        });
    }

    // Best practices for recommended items
    if (recommendedCount > 0) {
        recommendations.push({
            priority: 'low',
            title: 'Align with best practices',
            description: `Address ${recommendedCount} recommended change${recommendedCount > 1 ? 's' : ''} to optimize your cluster.`
        });
    }

    // If everything looks good
    if (recommendations.length === 0) {
        recommendations.push({
            priority: 'info',
            title: 'Maintain excellent configuration',
            description: 'Your cluster is well-configured. Continue monitoring for new best practices.'
        });
    }

    return recommendations;
};

// Analyze report data to extract key insights
export const analyzeReportData = (reportData) => {
    if (!reportData) return null;

    // Extract item categories
    const itemCategories = {};
    const extractCategory = (item) => {
        const parts = item.split(':');
        if (parts.length > 1) {
            const category = parts[0].trim();
            if (!itemCategories[category]) {
                itemCategories[category] = 1;
            } else {
                itemCategories[category]++;
            }
        }
    };

    reportData.itemsRequired?.forEach(extractCategory);
    reportData.itemsRecommended?.forEach(extractCategory);

    // Get top categories with issues
    const topCategories = Object.entries(itemCategories)
        .sort((a, b) => b[1] - a[1])
        .slice(0, 3)
        .map(([category, count]) => ({ category, count }));

    // Calculate category averages
    const categoryScores = [
        reportData.scoreInfra || 0,
        reportData.scoreGovernance || 0,
        reportData.scoreCompliance || 0,
        reportData.scoreMonitoring || 0,
        reportData.scoreBuildSecurity || 0
    ];

    const averageCategoryScore = categoryScores.reduce((sum, score) => sum + score, 0) / categoryScores.length;

    // Find strongest and weakest categories
    const strongestCategory = {
        name: getHighestScoreCategory(reportData),
        score: Math.max(...categoryScores)
    };

    const weakestCategory = {
        name: getLowestScoreCategory(reportData),
        score: Math.min(...categoryScores.filter(score => score > 0)) // Filter out zeros
    };

    return {
        topCategories,
        averageCategoryScore,
        strongestCategory,
        weakestCategory,
        recommendations: getHealthActionRecommendations(reportData)
    };
};

// Helper functions for the analyzer
function getHighestScoreCategory(reportData) {
    const scores = [
        { name: 'Infrastructure', score: reportData.scoreInfra || 0 },
        { name: 'Governance', score: reportData.scoreGovernance || 0 },
        { name: 'Compliance', score: reportData.scoreCompliance || 0 },
        { name: 'Monitoring', score: reportData.scoreMonitoring || 0 },
        { name: 'Build/Deploy', score: reportData.scoreBuildSecurity || 0 }
    ];

    return scores.sort((a, b) => b.score - a.score)[0].name;
}

function getLowestScoreCategory(reportData) {
    const scores = [
        { name: 'Infrastructure', score: reportData.scoreInfra || 0 },
        { name: 'Governance', score: reportData.scoreGovernance || 0 },
        { name: 'Compliance', score: reportData.scoreCompliance || 0 },
        { name: 'Monitoring', score: reportData.scoreMonitoring || 0 },
        { name: 'Build/Deploy', score: reportData.scoreBuildSecurity || 0 }
    ].filter(item => item.score > 0); // Filter out zeros

    return scores.sort((a, b) => a.score - b.score)[0].name;
}