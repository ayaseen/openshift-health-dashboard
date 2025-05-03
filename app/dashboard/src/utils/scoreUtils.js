// app/dashboard/src/utils/scoreUtils.js

// Get color based on score value - adjusted to match the orange color in your screenshot
export const getScoreColor = (score) => {
    const numericScore = parseFloat(score);
    if (isNaN(numericScore)) return '#EF4444'; // Default to red for invalid scores

    if (numericScore >= 90) return '#10B981'; // Excellent - Green
    if (numericScore >= 60) return '#F97316'; // Good/Fair - Orange (matches your screenshot)
    if (numericScore >= 40) return '#FB923C'; // Poor - Light Orange
    return '#EF4444'; // Critical - Red
};

// Get rating text based on score value
export const getScoreRating = (score) => {
    const numericScore = parseFloat(score);
    if (isNaN(numericScore)) return 'Unknown';

    if (numericScore >= 90) return 'Excellent';
    if (numericScore >= 75) return 'Good';
    if (numericScore >= 60) return 'Fair';
    if (numericScore >= 40) return 'Poor';
    return 'Critical';
};

// Get health assessment text based on score value
export const getHealthAssessment = (score) => {
    const numericScore = parseFloat(score);
    if (isNaN(numericScore)) return 'Unable to assess cluster health.';

    if (numericScore >= 90) return 'Your cluster is highly optimized and follows all best practices.';
    if (numericScore >= 75) return 'Your cluster is well configured with only minor improvements needed.';
    if (numericScore >= 60) return 'Your cluster is generally healthy but has several areas needing attention.';
    if (numericScore >= 40) return 'Your cluster needs significant improvements to meet best practices.';
    return 'Your cluster requires urgent attention to critical configuration issues.';
};

// Format score to ensure it's displayed correctly
export const formatScore = (score) => {
    const numericScore = parseFloat(score);
    if (isNaN(numericScore)) return '0';

    // Return as integer if it's a whole number, otherwise with one decimal place
    return Number.isInteger(numericScore)
        ? Math.round(numericScore)
        : parseFloat(numericScore.toFixed(1));
};

// Calculate a weighted overall score based on category scores
export const calculateOverallScore = (reportData) => {
    if (!reportData) return 75;

    const weights = {
        infra: 0.25,
        governance: 0.2,
        compliance: 0.2,
        monitoring: 0.15,
        buildSecurity: 0.2
    };

    const weightedSum =
        (reportData.scoreInfra || 75) * weights.infra +
        (reportData.scoreGovernance || 75) * weights.governance +
        (reportData.scoreCompliance || 75) * weights.compliance +
        (reportData.scoreMonitoring || 75) * weights.monitoring +
        (reportData.scoreBuildSecurity || 75) * weights.buildSecurity;

    return parseFloat(weightedSum.toFixed(1));
};

// Get severity level colors and labels
export const getSeverityLevel = (type) => {
    switch (type.toLowerCase()) {
        case 'required':
            return { color: 'text-red-700', bgColor: 'bg-red-50', borderColor: 'border-red-100' };
        case 'recommended':
            return { color: 'text-yellow-700', bgColor: 'bg-yellow-50', borderColor: 'border-yellow-100' };
        case 'advisory':
            return { color: 'text-blue-700', bgColor: 'bg-blue-50', borderColor: 'border-blue-100' };
        case 'no change':
            return { color: 'text-green-700', bgColor: 'bg-green-50', borderColor: 'border-green-100' };
        default:
            return { color: 'text-gray-700', bgColor: 'bg-gray-50', borderColor: 'border-gray-100' };
    }
};

// Generate a summary statement based on the report data
export const generateSummaryStatement = (reportData) => {
    if (!reportData) return '';

    const score = formatScore(reportData.overallScore);
    const rating = getScoreRating(score);
    const requiredCount = reportData.itemsRequired?.length || 0;
    const recommendedCount = reportData.itemsRecommended?.length || 0;

    let summary = `Cluster health is ${rating.toLowerCase()} (${score}%).`;

    if (requiredCount > 0) {
        summary += ` There ${requiredCount === 1 ? 'is' : 'are'} ${requiredCount} critical issue${requiredCount === 1 ? '' : 's'} requiring immediate attention.`;
    } else {
        summary += ' No critical issues were found.';
    }

    if (recommendedCount > 0) {
        summary += ` ${recommendedCount} recommended change${recommendedCount === 1 ? '' : 's'} would improve cluster health.`;
    }

    return summary;
};